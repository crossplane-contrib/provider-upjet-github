/*
Copyright 2021 Upbound Inc.
*/

package clients

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/crossplane/upjet/v2/pkg/terraform"

	"github.com/crossplane-contrib/provider-upjet-github/internal/directgrant"
)

// When the setup cache rebuilds a ProviderConfig's CachedTerraformSetup,
// replacing the map entry keyed by configRefName must deregister the
// superseded entry's setup.Meta, so the direct-grant registry does not grow
// without bound across setup rebuilds. This must fail against
// getOrBuildTerraformSetup with the eviction call removed: the old meta would
// stay resolvable forever.
func TestDirectGrantRegistryEvictedOnSetupReplacement(t *testing.T) {
	var lock sync.RWMutex
	cache := map[string]CachedTerraformSetup{}

	oldMeta := new(int)
	newMeta := new(int)
	directgrant.Register(oldMeta, "https://api.github.com/graphql", directgrant.StaticToken("old-token"), "octo-org")
	directgrant.Register(newMeta, "https://api.github.com/graphql", directgrant.StaticToken("new-token"), "octo-org")
	t.Cleanup(func() {
		directgrant.Deregister(oldMeta)
		directgrant.Deregister(newMeta)
	})

	now := time.Now()
	clock := func() time.Time { return now }

	var buildCount int32
	build := func(meta any) func() (terraform.Setup, error) {
		return func() (terraform.Setup, error) {
			atomic.AddInt32(&buildCount, 1)
			return terraform.Setup{Meta: meta}, nil
		}
	}

	if _, err := getOrBuildTerraformSetup(&lock, cache, "pc", clock, time.Minute, build(oldMeta)); err != nil {
		t.Fatalf("first build: unexpected error: %v", err)
	}
	if ok := directgrant.IsRegistered(oldMeta); !ok {
		t.Fatalf("expected oldMeta to be registered after the first build")
	}

	// Expire the cache entry so the next call rebuilds and replaces it.
	now = now.Add(2 * time.Minute)
	if _, err := getOrBuildTerraformSetup(&lock, cache, "pc", clock, time.Minute, build(newMeta)); err != nil {
		t.Fatalf("second build: unexpected error: %v", err)
	}

	if ok := directgrant.IsRegistered(oldMeta); ok {
		t.Fatalf("expected oldMeta to be evicted from the direct-grant registry after its CachedTerraformSetup was replaced")
	}
	if ok := directgrant.IsRegistered(newMeta); !ok {
		t.Fatalf("expected newMeta to remain registered")
	}
}

// build() always registers the new meta (via configureNoForkGithubClient ->
// registerDirectGrantClient) before getOrBuildTerraformSetup deregisters the
// superseded entry. That ordering is only safe while a rebuild's meta
// pointer never equals the one it replaces; if it ever did, deregistering
// unconditionally would evict the registration just installed rather than a
// stale one. This pins the guard against that: a rebuild that returns the
// same meta as the cached entry must leave it registered.
func TestDirectGrantRegistryNotEvictedWhenRebuildReusesSameMeta(t *testing.T) {
	var lock sync.RWMutex
	cache := map[string]CachedTerraformSetup{}

	meta := new(int)
	directgrant.Register(meta, "https://api.github.com/graphql", directgrant.StaticToken("token"), "octo-org")
	t.Cleanup(func() { directgrant.Deregister(meta) })

	now := time.Now()
	clock := func() time.Time { return now }
	build := func() (terraform.Setup, error) { return terraform.Setup{Meta: meta}, nil }

	if _, err := getOrBuildTerraformSetup(&lock, cache, "pc", clock, time.Minute, build); err != nil {
		t.Fatalf("first build: unexpected error: %v", err)
	}

	now = now.Add(2 * time.Minute)
	if _, err := getOrBuildTerraformSetup(&lock, cache, "pc", clock, time.Minute, build); err != nil {
		t.Fatalf("second build: unexpected error: %v", err)
	}

	if ok := directgrant.IsRegistered(meta); !ok {
		t.Fatal("expected meta to remain registered: the rebuild reused the same meta pointer, so it must not be deregistered as if it were superseded")
	}
}
