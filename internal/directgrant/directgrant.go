/*
Copyright 2021 Upbound Inc.
*/

package directgrant

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
)

// DirectGrant reports whether a login holds a direct (repository-sourced)
// grant, as opposed to access inherited through team or organization
// membership. REST's role_name cannot answer this -- it reports the
// effective role across every source, not the direct grant specifically --
// which is why this package asks GraphQL's permissionSources instead.
type DirectGrant struct {
	Exists   bool
	RoleName string // the direct grant's role, e.g. "write", "maintain", or a custom repository role name
}

// Warnf reports an operator-visible warning, in the same shape as
// crossplane-runtime's logging.Logger.Info -- its only "important" level,
// there being no Warn. A plain function type rather than logging.Logger
// itself, so this leaf package does not couple to that dependency.
type Warnf func(msg string, keysAndValues ...any)

var (
	warnMu sync.RWMutex
	warnFn Warnf
)

// SetLogger installs w as the logger every Warn call routes through,
// process-wide. TerraformSetupBuilder (internal/clients) calls this once per
// controller setup with the same crossplane-runtime Logger the rest of the
// provider uses. Without it, this package's failures would go to the stdlib
// log package, which cmd/provider/main.go silences
// (log.Default().SetOutput(io.Discard)).
func SetLogger(w Warnf) {
	warnMu.Lock()
	defer warnMu.Unlock()
	warnFn = w
}

// Warn reports an operator-visible warning through whatever logger
// SetLogger last installed. A logger that was never installed -- true for
// most tests -- makes this a silent no-op.
func Warn(msg string, keysAndValues ...any) {
	warnMu.RLock()
	w := warnFn
	warnMu.RUnlock()
	if w == nil {
		return
	}
	w(msg, keysAndValues...)
}

// query asks for up to 10 candidate collaborators matching login
// and their permission provenance. query: is a fuzzy filter (hence first:10
// and the case-insensitive exact match Lookup does on the result),
// not an exact lookup, so a returned edge is a candidate, not a match.
//
// pageInfo { hasNextPage } is what makes the first:10 bound safe in the other
// direction. collaborators(query:) matches display names as well as logins, so
// on a repository with many collaborators a common fragment can push the login
// we actually want past the window -- and an absent edge is indistinguishable
// from a revoked grant. When no exact match is found and the connection has
// more pages, match reports an error rather than "gone", so a live
// collaborator is never reaped on a truncated answer.
const query = `query($owner:String!, $name:String!, $login:String!) {
  repository(owner:$owner, name:$name) {
    collaborators(first:10, query:$login) {
      pageInfo { hasNextPage }
      edges {
        node { login }
        permissionSources { roleName source { __typename } }
      }
    }
  }
}`

type graphQLRequest struct {
	Query     string            `json:"query"`
	Variables map[string]string `json:"variables"`
}

type graphQLError struct {
	Message string `json:"message"`
}

type permissionSource struct {
	RoleName *string `json:"roleName"`
	Source   struct {
		Typename string `json:"__typename"`
	} `json:"source"`
}

type collaboratorEdge struct {
	Node struct {
		Login string `json:"login"`
	} `json:"node"`
	PermissionSources []permissionSource `json:"permissionSources"`
}

type directGrantResponse struct {
	Data struct {
		Repository *struct {
			Collaborators struct {
				PageInfo struct {
					HasNextPage bool `json:"hasNextPage"`
				} `json:"pageInfo"`
				Edges []collaboratorEdge `json:"edges"`
			} `json:"collaborators"`
		} `json:"repository"`
	} `json:"data"`
	Errors []graphQLError `json:"errors"`
}

// TokenProvider yields the GitHub token to authenticate a direct-grant lookup
// with, at the moment the lookup is made.
//
// It is deliberately a function rather than a captured string. A GitHub App
// installation token lives one hour, while the terraform setup that produced
// this registration is cached for up to that same lifetime
// (tfSetupCacheTTL, internal/clients/github.go). A token captured at
// registration time would therefore authenticate for only part of that
// window and 401 for the rest -- and because the Read wrapper's fail-safe
// swallows lookup errors, that would degrade silently back to the bug this
// whole mechanism exists to fix. Refresh has to run on the token's clock,
// not the setup cache's.
//
// Note for the app-auth path: the token source underneath (go-githubauth's
// installationTokenSource) captures its own context at construction, so ctx
// here does not bound a mint already in flight. internal/clients bounds it
// with an http.Client timeout instead (appTokenMintTimeout). Both budgets are
// 30s, so worst case is ~30s per Observe, not 60: a mint that exhausts its
// own budget has also exhausted the wrapper's directGrantLookupTimeout, so
// the GraphQL call that follows fails immediately on an already-expired ctx.
type TokenProvider func(ctx context.Context) (string, error)

// registeredClient is what the registry keeps per ProviderConfig meta: just
// enough to authenticate and address a GraphQL request, since github.Owner
// (the meta value Lookup is handed) has all fields unexported and
// no accessors.
type registeredClient struct {
	endpoint string
	token    TokenProvider
	owner    string
}

var (
	registryMu sync.RWMutex
	registry   = map[any]registeredClient{}
)

// Register records the GraphQL endpoint, token provider and
// ProviderConfig owner for meta (the value returned by
// schema.Provider.Meta(), which configureNoForkGithubClient assigns to
// ps.Meta), so Lookup can resolve them later without reaching into
// github.Owner's unexported fields.
//
// Registering with no token provider is refused: Lookup must fail loudly
// for that meta (fail-safe, never a silent "no direct grant" with no
// credential to have asked GitHub with). The provider yielding an empty
// token or an error is caught later, by Token, for the same reason.
func Register(meta any, endpoint string, token TokenProvider, owner string) {
	if meta == nil || token == nil {
		return
	}
	registryMu.Lock()
	defer registryMu.Unlock()
	registry[meta] = registeredClient{endpoint: endpoint, token: token, owner: owner}
}

// StaticToken adapts a fixed token string to a TokenProvider. It is what the
// plain-token ProviderConfig path registers; app_auth registers a minting
// provider instead.
func StaticToken(token string) TokenProvider {
	return func(context.Context) (string, error) { return token, nil }
}

// Deregister removes meta's registry entry, if any. Called
// when the setup cache evicts the CachedTerraformSetup that produced meta, so
// the registry does not grow without bound across setup rebuilds.
func Deregister(meta any) {
	if meta == nil {
		return
	}
	registryMu.Lock()
	defer registryMu.Unlock()
	delete(registry, meta)
}

func lookupClient(meta any) (registeredClient, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	c, ok := registry[meta]
	return c, ok
}

// IsRegistered reports whether meta currently has a registered client. It
// exists so the setup-cache eviction path (internal/clients) can be tested
// without exporting the registry itself.
func IsRegistered(meta any) bool {
	_, ok := lookupClient(meta)
	return ok
}

// Token resolves the GitHub token registered for meta, minting a fresh one if
// that is what the registered provider does.
//
// An unregistered meta, a provider that errors, and a provider that yields an
// empty token are all errors: Register cannot inspect the credential a
// TokenProvider will eventually produce, so an app whose mint silently
// returns nothing must be caught here, or an unauthenticated GraphQL request
// would produce a "no direct grant" answer that gets trusted.
func Token(ctx context.Context, meta any) (string, error) {
	c, ok := lookupClient(meta)
	if !ok {
		return "", fmt.Errorf("no GraphQL client registered for this ProviderConfig")
	}
	token, err := c.token(ctx)
	if err != nil {
		return "", fmt.Errorf("resolve GitHub token for direct-grant lookup: %w", err)
	}
	if token == "" {
		return "", fmt.Errorf("resolved an empty GitHub token for direct-grant lookup")
	}
	return token, nil
}

// Lookup resolves the GraphQL client registered for meta and asks
// GitHub which permission sources apply to login on repository. repository
// may be "owner/repo" or a bare "repo", in which case the ProviderConfig's
// owner is used.
//
// Every failure mode -- no registered client, a transport error, a non-200
// response, an unparseable body, or a populated GraphQL errors array --
// returns a non-nil error. A nil error with Exists: false means GitHub
// answered and no Repository-typed permission source exists for login; it
// never means the answer is unknown. Callers must not read an error as
// absence of a grant.
func Lookup(ctx context.Context, meta any, repository, login string) (DirectGrant, error) {
	c, ok := lookupClient(meta)
	if !ok {
		return DirectGrant{}, fmt.Errorf("no GraphQL client registered for this ProviderConfig")
	}

	owner, name, err := splitRepository(repository, c.owner)
	if err != nil {
		return DirectGrant{}, err
	}

	token, err := Token(ctx, meta)
	if err != nil {
		return DirectGrant{}, err
	}

	reqBody, err := json.Marshal(graphQLRequest{
		Query: query,
		Variables: map[string]string{
			"owner": owner,
			"name":  name,
			"login": login,
		},
	})
	if err != nil {
		return DirectGrant{}, fmt.Errorf("marshal GraphQL request: %w", err)
	}

	respBody, err := doGraphQLRequest(ctx, c.endpoint, token, reqBody)
	if err != nil {
		return DirectGrant{}, err
	}

	var parsed directGrantResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return DirectGrant{}, fmt.Errorf("decode GraphQL response: %w", err)
	}

	// A 200 with a populated errors array (e.g. INSUFFICIENT_SCOPES) is a
	// GraphQL-level failure, not an answer. It must be distinguishable from
	// "no direct grant" -- returning a non-nil error here is what makes that
	// so; callers must never read this as Exists: false.
	if len(parsed.Errors) > 0 {
		return DirectGrant{}, fmt.Errorf("GraphQL errors: %s", formatGraphQLErrors(parsed.Errors))
	}

	return match(parsed, login)
}

// doGraphQLRequest sends reqBody to endpoint, authenticated with token, and
// returns the raw response body. Split out of Lookup so Lookup's error
// handling stays within gocyclo's threshold; the request/response handling
// (build request, execute, read body, check status) is unchanged from what
// Lookup did inline.
func doGraphQLRequest(ctx context.Context, endpoint, token string, reqBody []byte) ([]byte, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("build GraphQL request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+token)

	// http.DefaultClient, not a bare &http.Client{}: any request-metrics or
	// rate-limit instrumentation a deployment installs on
	// http.DefaultClient.Transport should see this GraphQL traffic too,
	// alongside the terraform clients' REST traffic. A separate client would
	// make it invisible to that instrumentation.
	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("GraphQL request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read GraphQL response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GraphQL request failed: %s: %s", resp.Status, string(respBody))
	}
	return respBody, nil
}

func formatGraphQLErrors(errs []graphQLError) string {
	msgs := make([]string, 0, len(errs))
	for _, e := range errs {
		msgs = append(msgs, e.Message)
	}
	return strings.Join(msgs, "; ")
}

// match finds the edge whose node login matches login
// case-insensitively -- GitHub logins are case-insensitive, and upstream's
// schema carries DiffSuppressFunc: caseInsensitive() for this same reason --
// and reports whether that edge carries a Repository-typed permission source.
// Team and Organization sources are inherited access and are ignored.
//
// query: is a fuzzy filter, so an edge appearing in the response is a
// candidate, not a confirmed match; only an exact (case-insensitive) login
// match counts.
//
// Two shapes are errors rather than answers, both for the same reason: the
// caller clears the managed resource's ID on Exists: false, so anything short
// of a definite "GitHub says there is no direct grant" must not reach it.
//
//   - A nil data.repository. A genuinely missing repository comes back as a
//     GraphQL NOT_FOUND error, which the errors-array check has already
//     caught, so nothing legitimate arrives here -- but any intermediary
//     answering 200 with unrelated JSON decodes silently to this zero value.
//   - No exact match on a truncated connection. See the query constant.
func match(resp directGrantResponse, login string) (DirectGrant, error) {
	if resp.Data.Repository == nil {
		return DirectGrant{}, fmt.Errorf("GraphQL response contained no repository data")
	}
	for _, edge := range resp.Data.Repository.Collaborators.Edges {
		if !strings.EqualFold(edge.Node.Login, login) {
			continue
		}
		// An exact login match is authoritative whether or not the connection
		// has more pages: a login appears at most once in it, so no later page
		// can contradict this edge.
		for _, src := range edge.PermissionSources {
			if src.Source.Typename == "Repository" {
				role := ""
				if src.RoleName != nil {
					role = *src.RoleName
				}
				return DirectGrant{Exists: true, RoleName: role}, nil
			}
		}
		return DirectGrant{}, nil
	}
	if resp.Data.Repository.Collaborators.PageInfo.HasNextPage {
		return DirectGrant{}, fmt.Errorf("no exact match for login %q within the first %d collaborators and more pages remain: the answer is unknown, not absent", login, len(resp.Data.Repository.Collaborators.Edges))
	}
	return DirectGrant{}, nil
}

// splitRepository splits repository into owner and name.
// "owner/repo" is used verbatim; a bare "repo" falls back to owner, the
// registered ProviderConfig owner -- needed because github.Owner.name is
// unexported and unreachable from the caller.
//
// The bare form is the production path, not an edge case: upstream's
// parseRepoName (terraform-provider-github v6.13.0) resolves it against
// meta.(*Owner).name, which upstream populates in order of precedence:
// organization attribute, GITHUB_OWNER env var, owner attribute.
// Our resolution deliberately differs: the owner captured from ProviderConfig
// at registration time wins, and GITHUB_OWNER is only a fallback when that is
// empty. This divergence is harmless because under app_auth, setCredentialConfigs
// (internal/clients/github.go) returns an error when Owner is nil, so the
// registered owner is always populated and the env fallback is unreachable there.
func splitRepository(repository, owner string) (string, string, error) {
	if before, after, found := strings.Cut(repository, "/"); found {
		if before == "" || after == "" {
			return "", "", fmt.Errorf("invalid repository %q", repository)
		}
		return before, after, nil
	}
	if owner == "" {
		owner = os.Getenv(envGitHubOwner)
	}
	if owner == "" {
		return "", "", fmt.Errorf("repository %q has no owner, no ProviderConfig owner is registered, and %s is unset", repository, envGitHubOwner)
	}
	return owner, repository, nil
}

// envGitHubOwner is terraform-provider-github's own environment variable for
// the owner setting (github/provider.go at v6.13.0). Upstream's parseRepoName
// resolves a bare repository name against meta.(*Owner).name in order of
// precedence: organization attribute, GITHUB_OWNER env var, owner attribute.
// Our implementation uses the owner captured from ProviderConfig at registration
// time first, and falls back to this variable only when that is empty -- and
// that direction matches upstream's behavior when no organization or owner
// attribute is provided.
const envGitHubOwner = "GITHUB_OWNER"

// defaultGitHubBaseURL is what terraform-provider-github's own schema
// defaults base_url to (github/provider.go, DotComAPIURL) when it is absent
// from the ProviderConfig.
const defaultGitHubBaseURL = "https://api.github.com/"

// ghesRESTAPIPath is the REST API path suffix a raw GHES base_url normally
// carries (terraform-provider-github's GHESRESTAPIPath, github/config.go).
// It is a REST-only path segment -- GraphQL lives at a sibling path, "api/
// graphql" -- so it must be stripped before appending the GraphQL suffix, or
// the result doubles it: ".../api/v3/api/graphql".
const ghesRESTAPIPath = "api/v3/"

const (
	dotComHost    = "github.com"
	dotComAPIHost = "api.github.com"
)

// ghecAPIHostMatch and ghecHostMatch mirror terraform-provider-github's own
// GHECAPIHostMatch/GHECHostMatch (github/config.go): a GitHub Enterprise
// Cloud with Data Residency host, in its "api.<slug>.ghe.com" and
// "<slug>.ghe.com" forms respectively.
var (
	ghecAPIHostMatch = regexp.MustCompile(`^api\.[a-zA-Z0-9-]+\.ghe\.com$`)
	ghecHostMatch    = regexp.MustCompile(`\.ghe\.com$`)
)

// GraphQLEndpoint derives the GraphQL endpoint from a REST base_url the same
// way terraform-provider-github derives its own (getBaseURL,
// github/config.go:187-220, feeding NewGraphQLClient at line 111, at the
// pinned TERRAFORM_PROVIDER_VERSION v6.13.0). GraphQL lives at a sibling path
// from REST, not nested under it, so a GHES base_url's REST suffix
// ("api/v3/") is stripped before "graphql" is appended.
func GraphQLEndpoint(baseURL string) (string, error) {
	u, isGHES, err := normalizeBaseURL(baseURL)
	if err != nil {
		return "", err
	}
	suffix := "graphql"
	if isGHES {
		suffix = "api/graphql"
	}
	return u.JoinPath(suffix).String(), nil
}

// RESTEndpoint derives the REST API base the same way, for the one REST call
// this package's credentials feed: minting a GitHub App installation token.
// It mirrors upstream's own choice of pathSuffix at provider.go:424-429 --
// RESTAPIPath ("/") normally, GHESRESTAPIPath ("api/v3/") for GHES -- so that
// our mint and upstream's hit the same URL.
func RESTEndpoint(baseURL string) (string, error) {
	u, isGHES, err := normalizeBaseURL(baseURL)
	if err != nil {
		return "", err
	}
	suffix := "/"
	if isGHES {
		suffix = ghesRESTAPIPath
	}
	return u.JoinPath(suffix).String(), nil
}

// normalizeBaseURL mirrors terraform-provider-github's getBaseURL
// (github/config.go:187-220 at the pinned TERRAFORM_PROVIDER_VERSION v6.13.0),
// returning the normalized base and whether the host is a GHES one. The raw
// base_url this provider stores in ps.Configuration is the unnormalized
// credential string -- upstream's own p.Configure() normalizes only its
// internal Config.BaseURL, a separate object -- so callers here must
// normalize it themselves rather than assume it already happened:
//   - a bare "github.com" host is rewritten to "api.github.com" (GitHub does
//     not serve GraphQL, or anything else, at the bare host);
//   - a GHES base_url's REST suffix ("api/v3/") is stripped, so callers can
//     append whichever sibling path they need without doubling it.
//
// An absent base_url defaults to https://api.github.com/, mirroring the
// provider's own schema default, rather than upstream's "base url must not
// be empty" error.
func normalizeBaseURL(baseURL string) (*url.URL, bool, error) {
	if baseURL == "" {
		baseURL = defaultGitHubBaseURL
	}
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, false, fmt.Errorf("invalid base_url %q: %w", baseURL, err)
	}
	// JoinPath("/") is upstream's own normalization step (getBaseURL): it
	// cleans the path and guarantees a trailing slash, so the suffix check
	// below (ghesRESTAPIPath) matches regardless of whether the input ended
	// with one.
	u = u.JoinPath("/")

	switch {
	case u.Host == dotComAPIHost:
	case u.Host == dotComHost:
		u.Host = dotComAPIHost
	case ghecAPIHostMatch.MatchString(u.Host):
	case ghecHostMatch.MatchString(u.Host):
		u.Host = "api." + u.Host
	default:
		// GHES: strip the REST-only suffix a raw GHES base_url carries.
		u.Path = strings.TrimSuffix(u.Path, ghesRESTAPIPath)
		return u, true, nil
	}
	return u, false, nil
}
