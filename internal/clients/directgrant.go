/*
Copyright 2021 Upbound Inc.
*/

package clients

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jferrl/go-githubauth"
	"github.com/pkg/errors"

	"github.com/crossplane/upjet/v2/pkg/terraform"

	"github.com/crossplane-contrib/provider-upjet-github/internal/directgrant"
)

// appTokenMintTimeout bounds a GitHub App installation-token mint.
//
// It exists because go-githubauth's installationTokenSource captures its
// context at construction rather than taking one per Token() call, so the
// per-lookup deadline the Read wrapper sets (directGrantLookupTimeout,
// config/direct_grant.go) cannot reach the mint. Without this, a hung mint
// would block the reconciling goroutine indefinitely. A mint failure surfaces
// as an ordinary lookup error, which the wrapper's fail-safe already handles
// correctly: state untouched, ID never cleared.
const appTokenMintTimeout = 30 * time.Second

// registerDirectGrantClient registers a GraphQL client for
// directgrant.Lookup, keyed by the same meta pointer the config/ Read wrapper
// is handed. configureNoForkGithubClient is the only place that holds both the
// credentials (ps.Configuration) and the resulting meta, because
// github.Owner's v3client/v4client and name are unexported and unreachable
// from anywhere else.
//
// A ProviderConfig with neither a token nor an app_auth block registers
// nothing, and directgrant.Lookup then fails loudly for that meta rather than
// silently answering "no direct grant" with no credential to have asked GitHub
// with.
func registerDirectGrantClient(ps *terraform.Setup) error {
	baseURL, _ := ps.Configuration[keyBaseURL].(string)
	owner, _ := ps.Configuration[keyOwner].(string)

	token, err := directGrantTokenProvider(ps.Configuration, baseURL)
	if err != nil {
		return err
	}
	if token == nil {
		return nil
	}

	endpoint, err := directgrant.GraphQLEndpoint(baseURL)
	if err != nil {
		return errors.Wrap(err, "cannot derive GraphQL endpoint for direct-grant lookups")
	}
	directgrant.Register(ps.Meta, endpoint, token, owner)
	return nil
}

// directGrantTokenProvider builds the TokenProvider for this ProviderConfig's
// credentials, or returns nil when there are none to build one from.
//
// The two credential shapes are the two branches of setCredentialConfigs: a
// plain token, or an app_auth block. Checking app_auth first and letting a
// valid block win outright matches upstream's own precedence: it reads the
// token attribute only when config.AppID is nil (github/provider.go:397-416
// at v6.13.0). This must authenticate our GraphQL lookup as the same
// identity the terraform provider itself uses -- GitHub answers 404, and an
// empty collaborator list, for things a credential cannot see, so a narrower
// token here could report a live direct grant as absent and reap the managed
// resource. Gating on the token key alone would also miss app_auth entirely:
// setCredentialConfigs writes cnf[keyToken] only when creds.Token != nil,
// and upstream never writes its minted installation token back into this map
// (p.Configure receives ps.Configuration as read-only input, and
// GenerateOAuthTokenFromApp assigns to upstream's own internal Config.Token
// field, a different object).
func directGrantTokenProvider(cnf terraform.ProviderConfiguration, baseURL string) (directgrant.TokenProvider, error) {
	if appID, installationID, pemFile, ok := appAuthFromConfiguration(cnf); ok {
		return appInstallationTokenProvider(appID, installationID, pemFile, baseURL)
	}
	if token, _ := cnf[keyToken].(string); token != "" {
		return directgrant.StaticToken(token), nil
	}
	return nil, nil
}

// appAuthFromConfiguration reads the app_auth block, reporting whether it is
// complete. setCredentialConfigs is the only writer of this key, and it writes
// exactly this shape. The completeness check mirrors upstream's validateAppAuth
// (github/provider.go:711): all three fields, all non-empty, or the block does
// not count as app authentication at all.
func appAuthFromConfiguration(cnf terraform.ProviderConfiguration) (appID, installationID, pemFile string, ok bool) {
	aaList, isList := cnf[keyAppAuth].([]map[string]any)
	if !isList || len(aaList) == 0 || aaList[0] == nil {
		return "", "", "", false
	}
	aa := aaList[0]
	appID, _ = aa[keyAppAuthID].(string)
	installationID, _ = aa[keyAppAuthInstallationID].(string)
	pemFile, _ = aa[keyAppAuthPemFile].(string)
	if appID == "" || installationID == "" || pemFile == "" {
		return "", "", "", false
	}
	return appID, installationID, pemFile, true
}

// appInstallationTokenProvider mints GitHub App installation tokens on demand,
// through go-githubauth -- already in the module graph, and the library
// upstream itself uses for this (internal/ghclient/rest.go at
// terraform-provider-github v6.13.0).
//
// The token must be minted per call, not once at configure time. Even though
// tfSetupCacheTTL keeps this registration well inside the installation
// token's lifetime today, a captured token would silently stop being
// refreshed if that ever changed -- and the Read wrapper's fail-safe
// swallows lookup errors, so an expired captured token would degrade
// silently rather than loudly. NewInstallationTokenSource already returns a
// ReuseTokenSourceWithSkew, so the source is built once here and re-mints
// itself ahead of expiry rather than on every Observe.
func appInstallationTokenProvider(appID, installationID, pemFile, baseURL string) (directgrant.TokenProvider, error) {
	id, err := strconv.ParseInt(installationID, 10, 64)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot parse app_auth installation_id %q for direct-grant lookups", installationID)
	}

	// pem_file carries PEM content, not a path, and upstream un-escapes a
	// literal \n in it before use (validateAppAuth, github/provider.go:716 at
	// v6.13.0; the schema description reads "The GitHub App's PEM file
	// content; `\n` can be used for newlines"). Real newlines are unaffected.
	pemData := strings.ReplaceAll(pemFile, `\n`, "\n")

	// appID is passed as a string, which go-githubauth uses verbatim as the
	// JWT issuer -- correct for both an App ID and a Client ID, the two things
	// GitHub accepts there.
	appTokenSource, err := githubauth.NewApplicationTokenSource(appID, []byte(pemData))
	if err != nil {
		return nil, errors.Wrap(err, "cannot build a GitHub App token source for direct-grant lookups")
	}

	opts := []githubauth.InstallationTokenSourceOpt{
		githubauth.WithHTTPClient(&http.Client{Timeout: appTokenMintTimeout}),
	}
	if baseURL != "" {
		// WithBaseURL, not WithEnterpriseURL: the URL is used verbatim, and
		// RESTEndpoint has already applied the same normalization upstream
		// applies before its own mint, so both hit the same path.
		restURL, err := directgrant.RESTEndpoint(baseURL)
		if err != nil {
			return nil, errors.Wrap(err, "cannot derive REST endpoint for direct-grant app token minting")
		}
		opts = append(opts, githubauth.WithBaseURL(restURL))
	}

	tokenSource := githubauth.NewInstallationTokenSource(id, appTokenSource, opts...)
	return func(context.Context) (string, error) {
		token, err := tokenSource.Token()
		if err != nil {
			return "", errors.Wrap(err, "cannot mint a GitHub App installation token for a direct-grant lookup")
		}
		return token.AccessToken, nil
	}, nil
}
