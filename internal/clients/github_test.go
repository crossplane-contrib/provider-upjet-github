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

	"github.com/crossplane/upjet/v2/pkg/terraform"
)

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
