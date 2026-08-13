/*
Copyright 2021 Upbound Inc.
*/

package directgrant

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// testRepoName and testOwner are fixtures shared across the cases below --
// named so the many table-driven cases asserting against them don't repeat
// the literal past goconst's threshold.
const (
	testRepoName = "hello-world"
	testOwner    = "octo-org"
)

// capturedRequest is what a jsonServer records about the last request it
// received, so tests can assert on the GraphQL request Lookup
// actually sent (method, Authorization header, body) rather than just its
// parsed result.
type capturedRequest struct {
	method        string
	authorization string
	body          []byte
}

// jsonServer starts an httptest.Server that always answers with status and
// body, and records the last request it received (via the returned pointer,
// which the caller reads after Lookup returns).
func jsonServer(t *testing.T, status int, body string) (*httptest.Server, *capturedRequest) {
	t.Helper()
	captured := &capturedRequest{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		captured.method = r.Method
		captured.authorization = r.Header.Get("Authorization")
		captured.body = b
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv, captured
}

// registerTestClient registers a unique meta with the direct-grant registry
// pointed at srv, and deregisters it when the test ends so the registry does
// not accumulate entries across the test binary's run.
func registerTestClient(t *testing.T, endpoint, token, owner string) any {
	t.Helper()
	meta := new(int) // any comparable, unique pointer per call
	Register(meta, endpoint, StaticToken(token), owner)
	t.Cleanup(func() { Deregister(meta) })
	return meta
}

// A Repository-typed permission source is a direct grant; its roleName is
// returned verbatim.
func TestLookupDirectGrant_RepositorySourceExists(t *testing.T) {
	const body = `{"data":{"repository":{"collaborators":{"edges":[
		{"node":{"login":"octocat"},"permissionSources":[
			{"roleName":null,"source":{"__typename":"Organization"}},
			{"roleName":"write","source":{"__typename":"Repository"}}
		]}
	]}}}}`
	srv, captured := jsonServer(t, http.StatusOK, body)
	meta := registerTestClient(t, srv.URL, "test-token", testOwner)

	got, err := Lookup(context.Background(), meta, "octo-org/hello-world", "octocat")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := DirectGrant{Exists: true, RoleName: "write"}
	if got != want {
		t.Fatalf("Lookup() = %+v, want %+v", got, want)
	}

	// The request itself must carry the right shape: POST, bearer auth, and
	// the owner/name/login split into GraphQL variables.
	if captured.method != http.MethodPost {
		t.Errorf("request method = %q, want %q", captured.method, http.MethodPost)
	}
	if want := "Bearer test-token"; captured.authorization != want {
		t.Errorf("Authorization header = %q, want %q", captured.authorization, want)
	}
	var sent graphQLRequest
	if err := json.Unmarshal(captured.body, &sent); err != nil {
		t.Fatalf("captured body is not valid JSON: %v (%s)", err, captured.body)
	}
	if sent.Variables["owner"] != testOwner || sent.Variables["name"] != testRepoName || sent.Variables["login"] != "octocat" {
		t.Fatalf("unexpected GraphQL variables: %+v", sent.Variables)
	}
}

// Only Organization and Team sources, no Repository source, must report
// Exists: false with a nil error. The login is only inherited access, not a
// direct grant.
func TestLookupDirectGrant_OnlyInheritedSources_NoDirectGrant(t *testing.T) {
	const body = `{"data":{"repository":{"collaborators":{"edges":[
		{"node":{"login":"octocat"},"permissionSources":[
			{"roleName":null,"source":{"__typename":"Organization"}},
			{"roleName":"admin","source":{"__typename":"Team"}}
		]}
	]}}}}`
	srv, _ := jsonServer(t, http.StatusOK, body)
	meta := registerTestClient(t, srv.URL, "test-token", testOwner)

	got, err := Lookup(context.Background(), meta, "octo-org/widget-api", "octocat")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Exists {
		t.Fatalf("Lookup() = %+v, want Exists: false (inherited-only access)", got)
	}
}

// roleName must be read, not the lossy permission enum -- which cannot
// express "maintain" at all. This is the case that fails if the
// implementation reads a permission field instead of roleName.
func TestLookupDirectGrant_RoleNameNotPermissionEnum(t *testing.T) {
	const body = `{"data":{"repository":{"collaborators":{"edges":[
		{"node":{"login":"octocat"},"permissionSources":[
			{"roleName":"maintain","source":{"__typename":"Repository"}}
		]}
	]}}}}`
	srv, _ := jsonServer(t, http.StatusOK, body)
	meta := registerTestClient(t, srv.URL, "test-token", testOwner)

	got, err := Lookup(context.Background(), meta, "octo-org/hello-world", "octocat")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := DirectGrant{Exists: true, RoleName: "maintain"}
	if got != want {
		t.Fatalf("Lookup() = %+v, want %+v", got, want)
	}
}

// A custom repository role name must survive verbatim -- roleName is a
// String, not an enum, precisely so custom repository roles come through.
func TestLookupDirectGrant_CustomRoleNameSurvivesVerbatim(t *testing.T) {
	const body = `{"data":{"repository":{"collaborators":{"edges":[
		{"node":{"login":"octocat"},"permissionSources":[
			{"roleName":"Custom Repo Role","source":{"__typename":"Repository"}}
		]}
	]}}}}`
	srv, _ := jsonServer(t, http.StatusOK, body)
	meta := registerTestClient(t, srv.URL, "test-token", testOwner)

	got, err := Lookup(context.Background(), meta, "octo-org/hello-world", "octocat")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := DirectGrant{Exists: true, RoleName: "Custom Repo Role"}
	if got != want {
		t.Fatalf("Lookup() = %+v, want %+v", got, want)
	}
}

// A GraphQL 200 carrying a populated errors array (here the real
// INSUFFICIENT_SCOPES shape GitHub returns when permissionSources is
// requested without admin:org) must be a non-nil error. It must never be
// readable as "no direct grant" -- an error and an absent grant are different
// facts, and conflating them would let a scope withdrawal masquerade as a
// legitimate finding of "not a direct grant".
func TestLookupDirectGrant_GraphQLErrorsArray_IsError(t *testing.T) {
	const body = `{"data":{"repository":null},"errors":[{"type":"INSUFFICIENT_SCOPES","message":"Your token has not been granted the required scopes to execute this query. The 'permissionSources' field requires one of the following scopes: ['admin:org'], but your token has only been granted the: ['repo'] scopes. Please modify your token's scopes at: https://github.com/settings/tokens."}]}`
	srv, _ := jsonServer(t, http.StatusOK, body)
	meta := registerTestClient(t, srv.URL, "test-token", testOwner)

	got, err := Lookup(context.Background(), meta, "octo-org/hello-world", "octocat")
	if err == nil {
		t.Fatalf("expected a non-nil error for a populated GraphQL errors array, got DirectGrant %+v", got)
	}
	if !strings.Contains(err.Error(), "INSUFFICIENT_SCOPES") && !strings.Contains(err.Error(), "scopes") {
		t.Fatalf("expected the error to carry the GraphQL error detail, got: %v", err)
	}
}

// An HTTP 500, and a 200 whose body is not JSON, must each be a non-nil
// error.
func TestLookupDirectGrant_HTTPFailures(t *testing.T) {
	t.Run("HTTP 500", func(t *testing.T) {
		srv, _ := jsonServer(t, http.StatusInternalServerError, `{"message":"Internal Server Error"}`)
		meta := registerTestClient(t, srv.URL, "test-token", testOwner)

		_, err := Lookup(context.Background(), meta, "octo-org/hello-world", "octocat")
		if err == nil {
			t.Fatal("expected a non-nil error for HTTP 500")
		}
	})

	t.Run("non-JSON body", func(t *testing.T) {
		srv, _ := jsonServer(t, http.StatusOK, `not json at all`)
		meta := registerTestClient(t, srv.URL, "test-token", testOwner)

		_, err := Lookup(context.Background(), meta, "octo-org/hello-world", "octocat")
		if err == nil {
			t.Fatal("expected a non-nil error for a non-JSON response body")
		}
	})
}

// query: is a fuzzy filter. An edge for "octocat-bot" is a
// candidate, not a match, when "octocat" was asked for -- a substring or
// prefix match here would misreport inherited-only access as a direct grant
// (or vice versa) for any login that is a fuzzy neighbour of another.
func TestLookupDirectGrant_FuzzyQueryNearMiss_NotAMatch(t *testing.T) {
	const body = `{"data":{"repository":{"collaborators":{"edges":[
		{"node":{"login":"octocat-bot"},"permissionSources":[
			{"roleName":"write","source":{"__typename":"Repository"}}
		]}
	]}}}}`
	srv, _ := jsonServer(t, http.StatusOK, body)
	meta := registerTestClient(t, srv.URL, "test-token", testOwner)

	got, err := Lookup(context.Background(), meta, "octo-org/hello-world", "octocat")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Exists {
		t.Fatalf("Lookup() = %+v, want Exists: false (fuzzy near-miss, not a match)", got)
	}
}

// An unregistered meta -- no ProviderConfig ever configured a client for it
// -- must be an error, never a silent "no direct grant".
func TestLookupDirectGrant_UnregisteredMeta_Errors(t *testing.T) {
	meta := new(int) // never registered

	_, err := Lookup(context.Background(), meta, "octo-org/hello-world", "octocat")
	if err == nil {
		t.Fatal("expected a non-nil error for an unregistered meta")
	}
}

// The GraphQL endpoint must be derived the same way
// terraform-provider-github derives its own getBaseURL + NewGraphQLClient
// (github/config.go:187-220, 111 at TERRAFORM_PROVIDER_VERSION v6.13.0):
// base_url + "/graphql" for github.com (and ghe.com) hosts, base_url +
// "api/graphql" for GHES hosts, and https://api.github.com/ as the default
// when base_url is absent.
//
// Cases exercise upstream's normalization, which runs before the path
// suffix is appended, not just the pre-normalized shape:
//   - a GHES base_url carries the REST suffix "api/v3/"
//     (GHESRESTAPIPath) -- upstream strips it before appending "api/graphql",
//     so naively appending would double it: ".../api/v3/api/graphql".
//   - a bare "github.com" base_url (not "api.github.com") is rewritten to
//     the API host before "/graphql" is appended; GitHub does not serve
//     GraphQL at github.com itself.
//   - a GHEC Data Residency host ("api.<slug>.ghe.com" or "<slug>.ghe.com")
//     is a distinct branch from the github.com rewrite (ghecHostMatch vs.
//     dotComHost) and is not GHES, so it takes the plain "graphql" suffix,
//     not "api/graphql".
func TestGraphQLEndpoint(t *testing.T) {
	cases := map[string]struct {
		baseURL string
		want    string
	}{
		"api.github.com base URL": {
			baseURL: "https://api.github.com/",
			want:    "https://api.github.com/graphql",
		},
		"bare github.com host is rewritten to api.github.com": {
			baseURL: "https://github.com/",
			want:    "https://api.github.com/graphql",
		},
		"GHES base URL carrying the api/v3/ REST suffix": {
			baseURL: "https://ghes.example.com/api/v3/",
			want:    "https://ghes.example.com/api/graphql",
		},
		"absent base_url defaults to github.com": {
			baseURL: "",
			want:    "https://api.github.com/graphql",
		},
		"GHEC Data Residency API host is used verbatim": {
			baseURL: "https://api.acme.ghe.com/",
			want:    "https://api.acme.ghe.com/graphql",
		},
		"GHEC Data Residency bare host is rewritten to its api. host": {
			baseURL: "https://acme.ghe.com/",
			want:    "https://api.acme.ghe.com/graphql",
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := GraphQLEndpoint(tc.baseURL)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("GraphQLEndpoint(%q) = %q, want %q", tc.baseURL, got, tc.want)
			}
		})
	}
}

// A 200 whose data.repository is null, with no errors array, is not an
// answer: any intermediary answering 200 with unrelated JSON decodes
// straight to the zero value. A genuinely missing repository comes back as a
// GraphQL NOT_FOUND *error*, which the errors-array check already catches --
// so there is no legitimate shape that reaches here, and treating it as "no
// direct grant" would let a proxy's error page reap a live collaborator.
func TestLookup_NullRepository_IsError(t *testing.T) {
	cases := map[string]string{
		"data.repository is null":      `{"data":{"repository":null}}`,
		"data is empty":                `{"data":{}}`,
		"body is unrelated JSON":       `{"message":"upstream proxy error"}`,
		"body is an empty JSON object": `{}`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			srv, _ := jsonServer(t, http.StatusOK, body)
			meta := registerTestClient(t, srv.URL, "test-token", testOwner)

			got, err := Lookup(context.Background(), meta, "octo-org/hello-world", "octocat")
			if err == nil {
				t.Fatalf("expected a non-nil error for %s, got DirectGrant %+v", name, got)
			}
		})
	}
}

// collaborators(query:) is bounded at first:10 and matches
// display names as well as logins, so a real direct grant can fall outside the
// window on a repository with many collaborators whose names share a fragment.
// Reporting "gone" there would reap a live collaborator. When no exact login
// match was found and the connection has more pages, the answer is unknown --
// which must be an error, not Exists: false.
func TestLookup_TruncatedResultWithoutMatch_IsError(t *testing.T) {
	const body = `{"data":{"repository":{"collaborators":{
		"pageInfo":{"hasNextPage":true},
		"edges":[
			{"node":{"login":"octocat-bot"},"permissionSources":[
				{"roleName":"write","source":{"__typename":"Repository"}}
			]}
		]}}}}`
	srv, _ := jsonServer(t, http.StatusOK, body)
	meta := registerTestClient(t, srv.URL, "test-token", testOwner)

	got, err := Lookup(context.Background(), meta, "octo-org/hello-world", "octocat")
	if err == nil {
		t.Fatalf("expected a non-nil error when the result is truncated and no exact login matched, got DirectGrant %+v", got)
	}
}

// The inverse: an exact login match is authoritative regardless of
// hasNextPage -- a login appears at most once in the connection, so later
// pages cannot contradict it. Erroring here instead would turn every lookup
// on a large repository into a permanent fail-safe no-op.
func TestLookup_TruncatedResultWithMatch_IsAnswered(t *testing.T) {
	const body = `{"data":{"repository":{"collaborators":{
		"pageInfo":{"hasNextPage":true},
		"edges":[
			{"node":{"login":"octocat"},"permissionSources":[
				{"roleName":"write","source":{"__typename":"Repository"}}
			]}
		]}}}}`
	srv, _ := jsonServer(t, http.StatusOK, body)
	meta := registerTestClient(t, srv.URL, "test-token", testOwner)

	got, err := Lookup(context.Background(), meta, "octo-org/hello-world", "octocat")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := DirectGrant{Exists: true, RoleName: "write"}
	if got != want {
		t.Fatalf("Lookup() = %+v, want %+v", got, want)
	}
}

// The query must actually ask for pageInfo, or hasNextPage is always false
// and the truncation tests above pass for the wrong reason -- GitHub simply
// never tells us the result was truncated.
func TestQueryRequestsPageInfo(t *testing.T) {
	if !strings.Contains(query, "hasNextPage") {
		t.Fatal("the GraphQL query does not request pageInfo { hasNextPage }; truncation would be undetectable")
	}
}

// splitRepository's branches. The bare-repo branch is the production path,
// not an edge case: upstream's parseRepoName resolves the owner from
// meta.(*Owner).name, so the repository attribute this wrapper reads is
// routinely just "repo".
func TestSplitRepository(t *testing.T) {
	cases := map[string]struct {
		repository string
		owner      string
		env        string
		wantOwner  string
		wantName   string
		wantErr    bool
	}{
		"owner/repo is used verbatim": {
			repository: "octo-org/hello-world", owner: "other", wantOwner: testOwner, wantName: testRepoName,
		},
		"bare repo falls back to the registered owner": {
			repository: testRepoName, owner: testOwner, wantOwner: testOwner, wantName: testRepoName,
		},
		"bare repo falls back to GITHUB_OWNER when no owner is registered": {
			repository: testRepoName, owner: "", env: testOwner, wantOwner: testOwner, wantName: testRepoName,
		},
		"a registered owner beats GITHUB_OWNER": {
			repository: testRepoName, owner: testOwner, env: "SomeoneElse", wantOwner: testOwner, wantName: testRepoName,
		},
		"bare repo with no owner anywhere is an error": {
			repository: testRepoName, wantErr: true,
		},
		"empty owner segment is invalid": {
			repository: "/hello-world", owner: testOwner, wantErr: true,
		},
		"empty name segment is invalid": {
			repository: "octo-org/", owner: testOwner, wantErr: true,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Setenv(envGitHubOwner, tc.env)
			if tc.env == "" {
				_ = os.Unsetenv(envGitHubOwner)
			}
			gotOwner, gotName, err := splitRepository(tc.repository, tc.owner)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("splitRepository(%q, %q) = (%q, %q, nil), want an error", tc.repository, tc.owner, gotOwner, gotName)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotOwner != tc.wantOwner || gotName != tc.wantName {
				t.Fatalf("splitRepository(%q, %q) = (%q, %q), want (%q, %q)", tc.repository, tc.owner, gotOwner, gotName, tc.wantOwner, tc.wantName)
			}
		})
	}
}

// The token provider is what authenticates the request, and it is consulted
// per lookup rather than captured once -- an App installation token lives an
// hour while the setup cache holds this registration for 24.
func TestLookup_TokenProviderConsultedPerCall(t *testing.T) {
	const body = `{"data":{"repository":{"collaborators":{"edges":[]}}}}`
	srv, captured := jsonServer(t, http.StatusOK, body)

	var calls int
	meta := new(int)
	Register(meta, srv.URL, func(context.Context) (string, error) {
		calls++
		return fmt.Sprintf("token-%d", calls), nil
	}, testOwner)
	t.Cleanup(func() { Deregister(meta) })

	for i := 1; i <= 2; i++ {
		if _, err := Lookup(context.Background(), meta, "octo-org/hello-world", "octocat"); err != nil {
			t.Fatalf("lookup %d: unexpected error: %v", i, err)
		}
		if want := fmt.Sprintf("Bearer token-%d", i); captured.authorization != want {
			t.Fatalf("lookup %d used Authorization %q, want %q", i, captured.authorization, want)
		}
	}
}

// A token provider that fails, or yields an empty token, must be an error --
// never an unauthenticated request whose "no direct grant" answer would then
// be trusted. Register cannot inspect the credential a TokenProvider will
// eventually produce, so this guard has to live here instead.
func TestLookup_TokenProviderFailure_IsError(t *testing.T) {
	t.Run("provider errors", func(t *testing.T) {
		srv, _ := jsonServer(t, http.StatusOK, `{"data":{"repository":{"collaborators":{"edges":[]}}}}`)
		meta := new(int)
		Register(meta, srv.URL, func(context.Context) (string, error) {
			return "", errors.New("mint failed: 500 from GitHub")
		}, testOwner)
		t.Cleanup(func() { Deregister(meta) })

		if _, err := Lookup(context.Background(), meta, "octo-org/hello-world", "octocat"); err == nil {
			t.Fatal("expected a non-nil error when the token provider fails")
		}
	})

	t.Run("provider yields an empty token", func(t *testing.T) {
		srv, _ := jsonServer(t, http.StatusOK, `{"data":{"repository":{"collaborators":{"edges":[]}}}}`)
		meta := new(int)
		Register(meta, srv.URL, func(context.Context) (string, error) { return "", nil }, testOwner)
		t.Cleanup(func() { Deregister(meta) })

		if _, err := Lookup(context.Background(), meta, "octo-org/hello-world", "octocat"); err == nil {
			t.Fatal("expected a non-nil error when the token provider yields an empty token")
		}
	})

	t.Run("a nil token provider is refused at registration", func(t *testing.T) {
		meta := new(int)
		Register(meta, "https://api.github.com/graphql", nil, testOwner)
		if IsRegistered(meta) {
			t.Fatal("Register accepted a nil token provider")
		}
	})
}

// Warn must reach the logger SetLogger last installed, and carry the
// message and structured pairs through unchanged.
func TestWarn_ReachesInstalledLogger(t *testing.T) {
	t.Cleanup(func() { SetLogger(nil) })

	var gotMsg string
	var gotKVs []any
	SetLogger(func(msg string, keysAndValues ...any) {
		gotMsg = msg
		gotKVs = keysAndValues
	})

	Warn("direct-grant lookup failed", "login", "octocat", "repository", "octo-org/hello-world")

	if gotMsg != "direct-grant lookup failed" {
		t.Fatalf("logger received message %q, want %q", gotMsg, "direct-grant lookup failed")
	}
	want := []any{"login", "octocat", "repository", "octo-org/hello-world"}
	if len(gotKVs) != len(want) {
		t.Fatalf("logger received keysAndValues %v, want %v", gotKVs, want)
	}
	for i := range want {
		if gotKVs[i] != want[i] {
			t.Fatalf("logger received keysAndValues %v, want %v", gotKVs, want)
		}
	}
}

// No logger installed -- the state every test not calling SetLogger runs
// in, and what a package that never wires TerraformSetupBuilder would see
// in production -- must make Warn a silent no-op rather than panic.
func TestWarn_NoLoggerInstalled_NoOp(t *testing.T) {
	SetLogger(nil)
	Warn("should not panic", "key", "value")
}
