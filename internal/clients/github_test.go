/*
Copyright 2021 Upbound Inc.
*/

package clients

import (
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/crossplane/crossplane-runtime/v2/pkg/logging"
	"github.com/crossplane/upjet/v2/pkg/terraform"
	"github.com/integrations/terraform-provider-github/v6/github"

	"github.com/crossplane-contrib/provider-upjet-github/internal/directgrant"
)

// capturingLogger implements logging.Logger, recording every Info call so a
// test can assert on what actually reached it -- rather than on a
// logging.Logger-shaped stub that was never wired to anything real.
type capturingLogger struct {
	mu    sync.Mutex
	infos []string
}

func (l *capturingLogger) Info(msg string, _ ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.infos = append(l.infos, msg)
}

func (l *capturingLogger) Debug(string, ...any) {}

func (l *capturingLogger) WithValues(...any) logging.Logger { return l }

func (l *capturingLogger) contains(substr string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, m := range l.infos {
		if strings.Contains(m, substr) {
			return true
		}
	}
	return false
}

// TestGetOrBuildTerraformSetup_ConcurrentColdStart reproduces the cold-start
// thundering-herd bug: when many managed resources reconcile at once against a
// ProviderConfig whose setup has not been cached yet, only one goroutine should
// build the setup and every other caller should receive that same setup — never
// an "github token not ready yet" error, and never a data race on the cache map.
//
// Run with -race; the buggy implementation reads the cache map without a lock
// while another goroutine writes it under the lock, which is a fatal Go panic in
// production.
func TestGetOrBuildTerraformSetup_ConcurrentColdStart(t *testing.T) {
	const n = 50

	var lock sync.RWMutex
	cache := map[string]CachedTerraformSetup{}

	var buildCount int32
	var firstBuild sync.Once
	entered := make(chan struct{}) // closed when the first build starts (lock held)
	release := make(chan struct{}) // holds the first build open to widen contention

	build := func() (terraform.Setup, error) {
		atomic.AddInt32(&buildCount, 1)
		firstBuild.Do(func() { close(entered) })
		<-release
		return terraform.Setup{
			Configuration: terraform.ProviderConfiguration{"token": "test-token"},
		}, nil
	}

	// Barrier so all goroutines are released as close to simultaneously as
	// possible, maximizing contention on the cache/lock.
	var gate sync.WaitGroup
	gate.Add(1)

	errs := make(chan error, n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			gate.Wait()
			_, err := getOrBuildTerraformSetup(&lock, cache, "provider-config", time.Now, time.Minute, build)
			errs <- err
		}()
	}

	gate.Done() // release the herd
	<-entered   // a build is now in flight, holding the lock

	// Give the contending goroutines time to pile onto the lock while the build
	// is held open, so the buggy TryLock path deterministically reports the error.
	time.Sleep(25 * time.Millisecond)
	close(release)

	wg.Wait()
	close(errs)

	var failures int
	for err := range errs {
		if err != nil {
			failures++
			if !strings.Contains(err.Error(), errGitHubTokenNotReady) {
				t.Errorf("unexpected error: %v", err)
			}
		}
	}

	if failures != 0 {
		t.Errorf("expected 0 callers to fail, got %d/%d with %q", failures, n, errGitHubTokenNotReady)
	}
	if got := atomic.LoadInt32(&buildCount); got != 1 {
		t.Errorf("expected exactly 1 build (single-flight), got %d", got)
	}
}

// TestGetOrBuildTerraformSetup_CachesUntilExpiry verifies the cache fast path:
// a second call within the TTL returns the cached setup without rebuilding, and
// a call after expiry rebuilds.
func TestGetOrBuildTerraformSetup_CachesUntilExpiry(t *testing.T) {
	var lock sync.RWMutex
	cache := map[string]CachedTerraformSetup{}

	var buildCount int32
	build := func() (terraform.Setup, error) {
		atomic.AddInt32(&buildCount, 1)
		return terraform.Setup{}, nil
	}

	now := time.Now()
	clock := func() time.Time { return now }

	if _, err := getOrBuildTerraformSetup(&lock, cache, "pc", clock, time.Minute, build); err != nil {
		t.Fatalf("first call: unexpected error: %v", err)
	}
	if _, err := getOrBuildTerraformSetup(&lock, cache, "pc", clock, time.Minute, build); err != nil {
		t.Fatalf("second call: unexpected error: %v", err)
	}
	if got := atomic.LoadInt32(&buildCount); got != 1 {
		t.Fatalf("expected cached second call (1 build), got %d builds", got)
	}

	// Advance past the TTL: the setup should be rebuilt.
	now = now.Add(2 * time.Minute)
	if _, err := getOrBuildTerraformSetup(&lock, cache, "pc", clock, time.Minute, build); err != nil {
		t.Fatalf("post-expiry call: unexpected error: %v", err)
	}
	if got := atomic.LoadInt32(&buildCount); got != 2 {
		t.Fatalf("expected rebuild after expiry (2 builds), got %d builds", got)
	}
}

// TestTFSetupCacheTTLOutlivedByToken verifies the cache expires far enough ahead
// of the GitHub App installation token that a caller still holding a Setup —
// notably upjet's async external client, which keeps one for a whole create or
// update — cannot make calls with an already-expired token and get back
// 401 Bad credentials.
func TestTFSetupCacheTTLOutlivedByToken(t *testing.T) {
	if tfSetupCacheTTL <= 0 {
		t.Fatalf("tfSetupCacheTTL must be positive, got %s", tfSetupCacheTTL)
	}

	if got := tfSetupCacheTTL + tfSetupMaxHold; got > githubInstallationTokenLifetime {
		t.Errorf("a Setup taken at the end of the TTL and held for %s outlives its %s token by %s; lower tfSetupCacheTTL",
			tfSetupMaxHold, githubInstallationTokenLifetime, got-githubInstallationTokenLifetime)
	}
}

// TerraformSetupBuilder installs its Logger into directgrant.SetLogger, so
// wrapReadForDirectGrant's fail-safe (config/direct_grant.go) can report
// through the same Logger as everything else in this provider. That wrapper
// has no logger of its own to use instead -- it runs inside the terraform
// SDK's legacy schema.ReadFunc signature -- so without this wiring its
// failures would go to the stdlib log package, which cmd/provider/main.go
// discards (log.Default().SetOutput(io.Discard)).
//
// Must fail against a build that never calls directgrant.SetLogger:
// directgrant.Warn would then be a silent no-op.
func TestTerraformSetupBuilder_InstallsDirectGrantLogger(t *testing.T) {
	logger := &capturingLogger{}
	_ = TerraformSetupBuilder(github.NewProvider("dev", "none")(), logger)
	t.Cleanup(func() { directgrant.SetLogger(nil) })

	directgrant.Warn("simulated direct-grant failure")

	if !logger.contains("simulated direct-grant failure") {
		t.Fatal("expected TerraformSetupBuilder to install its Logger into directgrant.SetLogger")
	}
}
