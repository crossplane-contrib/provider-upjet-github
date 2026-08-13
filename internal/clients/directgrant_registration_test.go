/*
Copyright 2021 Upbound Inc.
*/

package clients

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/crossplane/crossplane-runtime/v2/pkg/logging"
	"github.com/crossplane/upjet/v2/pkg/terraform"
	"github.com/integrations/terraform-provider-github/v6/github"

	"github.com/crossplane-contrib/provider-upjet-github/internal/directgrant"
)

// testAppPEM generates a throwaway RSA private key in PKCS#1 PEM form -- the
// shape a GitHub App's pem_file carries, and what both upstream's
// generateAppJWT (go-jose) and go-githubauth's ParseRSAPrivateKeyFromPEM
// accept. No real credential is involved.
func testAppPEM(t *testing.T) string {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate RSA key: %v", err)
	}
	return string(pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}))
}

// appAuthServer stands in for GitHub's installation-token endpoint. It answers
// any path ending in access_tokens with 201 Created (the status
// getInstallationAccessToken and go-githubauth both require; a 200 is treated
// as a failure by each) and a token carrying a real future expiry.
//
// The expiry matters: go-githubauth's ReuseTokenSourceWithSkew treats a zero
// Expiry as "valid forever", so omitting expires_at would silently make this
// test assert the captured-token behaviour the token provider exists to avoid.
//
// The path is matched by suffix so the one server serves both upstream's mint
// during p.Configure (which routes through GHESRESTAPIPath, "api/v3/", because
// getBaseURL classifies an httptest host as GHES) and the mint our own token
// provider makes.
func appAuthServer(t *testing.T, token string, mints *int32) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "access_tokens") {
			http.Error(w, fmt.Sprintf("unexpected path %q", r.URL.Path), http.StatusNotFound)
			return
		}
		atomic.AddInt32(mints, 1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"token":      token,
			"expires_at": time.Now().Add(time.Hour).UTC().Format(time.RFC3339),
		})
	}))
	t.Cleanup(srv.Close)
	return srv
}

func appAuthConfiguration(baseURL, pemFile string) terraform.ProviderConfiguration {
	return terraform.ProviderConfiguration{
		keyBaseURL: baseURL + "/",
		keyOwner:   "octo-org",
		keyAppAuth: []map[string]any{{
			keyAppAuthID:             "12345",
			keyAppAuthInstallationID: "67890",
			keyAppAuthPemFile:        pemFile,
		}},
	}
}

// app_auth is what production uses, and setCredentialConfigs never writes
// ps.Configuration["token"] on that path -- it writes app_auth and nothing
// else. Upstream never writes a minted token back either (p.Configure takes
// ps.Configuration as read-only input, and GenerateOAuthTokenFromApp assigns
// to its own internal Config.Token). So gating registration on
// ps.Configuration["token"] would make the whole direct-grant fix a silent
// no-op in production.
//
// This exercises the real configure path end to end and asserts both halves:
// a client is registered for the resulting meta, and its token provider yields
// a usable token.
func TestConfigureNoForkGithubClient_AppAuthRegistersUsableDirectGrantClient(t *testing.T) {
	const wantToken = "ghs-installation-token"
	var mints int32
	srv := appAuthServer(t, wantToken, &mints)

	ps := terraform.Setup{Configuration: appAuthConfiguration(srv.URL, testAppPEM(t))}
	p := github.NewProvider("dev", "none")()

	if err := configureNoForkGithubClient(context.Background(), &ps, *p, logging.NewNopLogger()); err != nil {
		t.Fatalf("configureNoForkGithubClient: %v", err)
	}
	t.Cleanup(func() { directgrant.Deregister(ps.Meta) })

	if !directgrant.IsRegistered(ps.Meta) {
		t.Fatal("no direct-grant client registered for an app_auth ProviderConfig: every LookupDirectGrant would fail and the wrapper's fail-safe would silently do nothing")
	}

	got, err := directgrant.Token(context.Background(), ps.Meta)
	if err != nil {
		t.Fatalf("registered token provider failed: %v", err)
	}
	if got != wantToken {
		t.Fatalf("token provider yielded %q, want %q", got, wantToken)
	}
	if n := atomic.LoadInt32(&mints); n == 0 {
		t.Fatal("expected the token provider to mint an installation token")
	}
}

// pem_file carries PEM *content*, and upstream's validateAppAuth
// (provider.go:716 at v6.13.0) un-escapes a literal \n in it before use --
// "The GitHub App's PEM file content; `\n` can be used for newlines", per the
// schema description. A ProviderConfig written that way must work here too,
// or app_auth registration fails for exactly the credential format the
// upstream docs recommend.
func TestConfigureNoForkGithubClient_AppAuthEscapedNewlinePEM(t *testing.T) {
	const wantToken = "ghs-escaped-pem-token"
	var mints int32
	srv := appAuthServer(t, wantToken, &mints)

	escaped := strings.ReplaceAll(testAppPEM(t), "\n", `\n`)
	ps := terraform.Setup{Configuration: appAuthConfiguration(srv.URL, escaped)}
	p := github.NewProvider("dev", "none")()

	if err := configureNoForkGithubClient(context.Background(), &ps, *p, logging.NewNopLogger()); err != nil {
		t.Fatalf("configureNoForkGithubClient: %v", err)
	}
	t.Cleanup(func() { directgrant.Deregister(ps.Meta) })

	got, err := directgrant.Token(context.Background(), ps.Meta)
	if err != nil {
		t.Fatalf("registered token provider failed for a \\n-escaped pem_file: %v", err)
	}
	if got != wantToken {
		t.Fatalf("token provider yielded %q, want %q", got, wantToken)
	}
}

// The plain-token path must keep working, and must not mint anything.
//
// base_url points at the fixture server for the same reason the app_auth tests
// do: configureProviderMeta looks the owner up over REST (provider.go:618),
// and while it tolerates a 404 it fails the whole Configure on anything else.
func TestConfigureNoForkGithubClient_PlainTokenRegistersStaticProvider(t *testing.T) {
	var mints int32
	srv := appAuthServer(t, "unused", &mints)

	ps := terraform.Setup{Configuration: terraform.ProviderConfiguration{
		keyBaseURL: srv.URL + "/",
		keyOwner:   "octo-org",
		keyToken:   "ghp-plain-token",
	}}
	p := github.NewProvider("dev", "none")()

	if err := configureNoForkGithubClient(context.Background(), &ps, *p, logging.NewNopLogger()); err != nil {
		t.Fatalf("configureNoForkGithubClient: %v", err)
	}
	t.Cleanup(func() { directgrant.Deregister(ps.Meta) })

	got, err := directgrant.Token(context.Background(), ps.Meta)
	if err != nil {
		t.Fatalf("registered token provider failed: %v", err)
	}
	if want := "ghp-plain-token"; got != want {
		t.Fatalf("token provider yielded %q, want %q", got, want)
	}
	if n := atomic.LoadInt32(&mints); n != 0 {
		t.Fatalf("plain-token path minted %d installation tokens, want 0", n)
	}
}

// A ProviderConfig carrying both a token and an app_auth block must
// authenticate the direct-grant lookup the way upstream authenticates itself:
// as the App. Upstream reads the token attribute only when config.AppID is nil
// (github/provider.go:405-416 at v6.13.0). Diverging here would point the
// GraphQL lookup at a different identity than the terraform provider uses --
// and since GitHub answers 404, and an empty collaborator list, for anything a
// credential cannot see, a narrower token could report a live direct grant as
// absent and reap the managed resource.
func TestConfigureNoForkGithubClient_AppAuthWinsOverToken(t *testing.T) {
	const wantToken = "ghs-app-wins"
	var mints int32
	srv := appAuthServer(t, wantToken, &mints)

	cnf := appAuthConfiguration(srv.URL, testAppPEM(t))
	cnf[keyToken] = "ghp-should-not-be-used"
	ps := terraform.Setup{Configuration: cnf}
	p := github.NewProvider("dev", "none")()

	if err := configureNoForkGithubClient(context.Background(), &ps, *p, logging.NewNopLogger()); err != nil {
		t.Fatalf("configureNoForkGithubClient: %v", err)
	}
	t.Cleanup(func() { directgrant.Deregister(ps.Meta) })

	got, err := directgrant.Token(context.Background(), ps.Meta)
	if err != nil {
		t.Fatalf("registered token provider failed: %v", err)
	}
	if got != wantToken {
		t.Fatalf("token provider yielded %q, want the minted installation token %q: app_auth must win over token, as it does upstream", got, wantToken)
	}
}

// An incomplete app_auth block never reaches registration at all: upstream
// rejects it outright rather than falling back to the token attribute
// (provider.go:406-408), so the whole setup build fails first.
//
// This pins that, because it is the reason appAuthFromConfiguration's
// completeness check cannot be observed choosing the token instead -- the
// check is defence in depth, not a reachable branch, and a future upstream
// that softened this error is exactly when it would start to matter.
func TestConfigureNoForkGithubClient_IncompleteAppAuthIsRejectedUpstream(t *testing.T) {
	var mints int32
	srv := appAuthServer(t, "unused", &mints)

	ps := terraform.Setup{Configuration: terraform.ProviderConfiguration{
		keyBaseURL: srv.URL + "/",
		keyOwner:   "octo-org",
		keyToken:   "ghp-plain-token",
		keyAppAuth: []map[string]any{{
			keyAppAuthID:             "12345",
			keyAppAuthInstallationID: "",
			keyAppAuthPemFile:        "",
		}},
	}}
	p := github.NewProvider("dev", "none")()

	err := configureNoForkGithubClient(context.Background(), &ps, *p, logging.NewNopLogger())
	if err == nil {
		t.Fatal("expected an incomplete app_auth block to fail the provider configure")
	}
	if !strings.Contains(err.Error(), "app_auth block is set but required fields are missing") {
		t.Fatalf("unexpected error: %v", err)
	}
	if directgrant.IsRegistered(ps.Meta) {
		t.Fatal("registered a direct-grant client for a ProviderConfig whose configure failed")
	}
}

// No credentials at all registers nothing, so directgrant.Lookup fails loudly
// for that meta rather than answering "no direct grant" with nothing to have
// asked GitHub with.
func TestConfigureNoForkGithubClient_NoCredentialsRegistersNothing(t *testing.T) {
	var mints int32
	srv := appAuthServer(t, "unused", &mints)

	ps := terraform.Setup{Configuration: terraform.ProviderConfiguration{
		keyBaseURL: srv.URL + "/",
	}}
	p := github.NewProvider("dev", "none")()

	if err := configureNoForkGithubClient(context.Background(), &ps, *p, logging.NewNopLogger()); err != nil {
		t.Fatalf("configureNoForkGithubClient: %v", err)
	}
	t.Cleanup(func() { directgrant.Deregister(ps.Meta) })

	if directgrant.IsRegistered(ps.Meta) {
		t.Fatal("registered a direct-grant client with no credentials")
	}
}

// A direct-grant client that cannot be built must not fail the whole
// terraform.Setup. registerDirectGrantClient gained three new error sources
// with the token provider -- ParseInt on installation_id, the PEM parse, and
// RESTEndpoint -- and configureNoForkGithubClient is called from the setup
// cache's build function. Propagating any of them means that ProviderConfig
// gets no Setup at all and every managed resource under it stops
// reconciling: strictly worse than the effective-role bug this whole
// mechanism exists to fix, and the wrong direction for a design whose rule is
// "degrade to today's known bug, never make things worse."
//
// installation_id is the probe because upstream's own Configure does not
// parse it as an integer, so p.Configure succeeds and control reaches our
// code -- the realistic production shape being a Secret value with a trailing
// newline.
func TestConfigureNoForkGithubClient_UnbuildableClientDoesNotFailSetup(t *testing.T) {
	// A malformed pem_file is deliberately not a case here: upstream's own
	// Configure parses the key and rejects it first ("asn1: structure error"), so
	// that input never reaches our code and failing the setup is upstream's
	// correct, pre-existing behaviour rather than a regression of ours.
	cases := map[string]func(terraform.ProviderConfiguration){
		"installation_id is not an integer": func(cnf terraform.ProviderConfiguration) {
			cnf[keyAppAuth].([]map[string]any)[0][keyAppAuthInstallationID] = "67890\n"
		},
	}

	for name, mangle := range cases {
		t.Run(name, func(t *testing.T) {
			var mints int32
			srv := appAuthServer(t, "ghs-token", &mints)
			cnf := appAuthConfiguration(srv.URL, testAppPEM(t))
			mangle(cnf)

			ps := terraform.Setup{Configuration: cnf}
			p := github.NewProvider("dev", "none")()

			if err := configureNoForkGithubClient(context.Background(), &ps, *p, logging.NewNopLogger()); err != nil {
				t.Fatalf("configureNoForkGithubClient failed the whole setup build over an unbuildable direct-grant client: %v", err)
			}
			t.Cleanup(func() { directgrant.Deregister(ps.Meta) })

			// Nothing registered is the correct degraded state: Lookup then
			// fails loudly per call and the Read wrapper's fail-safe leaves
			// upstream's answer standing.
			if directgrant.IsRegistered(ps.Meta) {
				t.Fatal("registered a direct-grant client that could not be built")
			}
		})
	}
}

// A registerDirectGrantClient failure used to report only via log.Printf on
// the stdlib default logger, which main.go discards
// (log.Default().SetOutput(io.Discard)) -- so it never reached an operator.
// This proves the failure now lands in the caller's own Logger, not just a
// logging.Logger-shaped stub: must fail against an implementation that logs
// anywhere other than the l it was handed.
func TestConfigureNoForkGithubClient_UnbuildableClientLogsThroughCallerLogger(t *testing.T) {
	var mints int32
	srv := appAuthServer(t, "ghs-token", &mints)
	cnf := appAuthConfiguration(srv.URL, testAppPEM(t))
	cnf[keyAppAuth].([]map[string]any)[0][keyAppAuthInstallationID] = "67890\n"

	ps := terraform.Setup{Configuration: cnf}
	p := github.NewProvider("dev", "none")()
	logger := &capturingLogger{}

	if err := configureNoForkGithubClient(context.Background(), &ps, *p, logger); err != nil {
		t.Fatalf("configureNoForkGithubClient: %v", err)
	}
	t.Cleanup(func() { directgrant.Deregister(ps.Meta) })

	if !logger.contains("cannot register a direct-grant client") {
		t.Fatalf("expected the registration failure to be logged through the caller's Logger, got: %v", logger.infos)
	}
}
