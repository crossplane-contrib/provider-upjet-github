/*
Copyright 2021 Upbound Inc.
*/

package config

import (
	"context"
	"fmt"
	"testing"

	tpg "github.com/integrations/terraform-provider-github/v6/github"
)

// TestDirectGrantWorkaroundResourceIsWired guards against the silent-skip
// failure mode: withDirectGrantWorkaround does
// `if !ok || r == nil || r.Read == nil { return p }`, so if
// github_repository_collaborator is renamed, removed, or switches from the
// legacy Read field to ReadContext upstream, the wrapper would silently stop
// doing anything at all -- no build failure, no test failure, just a quiet
// return to the wedging/misreported-permission bug direct_grant.go exists to
// fix. This builds the real terraform-provider-github provider and asserts
// the resource is present and still exposes a non-nil legacy Read.
func TestDirectGrantWorkaroundResourceIsWired(t *testing.T) {
	p := tpg.NewProvider("dev", "none")()
	r, ok := p.ResourcesMap[directGrantWorkaroundResourceName]
	if !ok {
		t.Fatalf("%q is not a registered terraform-provider-github resource (would be silently skipped by withDirectGrantWorkaround -- renamed key?)", directGrantWorkaroundResourceName)
	}
	if r.Read == nil { //nolint:staticcheck // SA1019: verifying the legacy Read field the wrapper depends on is present.
		t.Fatalf("%q has no legacy Read func; withDirectGrantWorkaround would silently skip it", directGrantWorkaroundResourceName)
	}
}

// TestWithDirectGrantWorkaroundReplacesRead asserts the composition:
// withDirectGrantWorkaround must actually replace
// github_repository_collaborator's Read function pointer, not merely leave
// it wired (which the test above already covers).
func TestWithDirectGrantWorkaroundReplacesRead(t *testing.T) {
	p := tpg.NewProvider("dev", "none")()
	before := fmt.Sprintf("%p", p.ResourcesMap[directGrantWorkaroundResourceName].Read) //nolint:staticcheck // SA1019: see above.

	withDirectGrantWorkaround(p)

	after := fmt.Sprintf("%p", p.ResourcesMap[directGrantWorkaroundResourceName].Read) //nolint:staticcheck // SA1019: see above.
	if after == before {
		t.Errorf("%q Read was not wrapped (pointer unchanged: %s)", directGrantWorkaroundResourceName, after)
	}
}

// TestDirectGrantWorkaroundInstalledOnBothProviders is the test that catches
// a one-sided edit: github_repository_collaborator's configuration is
// doubled across config/provider_cluster.go and config/provider_namespaced.go,
// each composing withDirectGrantWorkaround at its own
// ujconfig.WithTerraformProvider call site. Wiring only one means the
// namespaced (or cluster) managed resources keep wedging/misreporting
// permission silently forever -- nothing else would fail.
//
// Each entrypoint is asserted independently against the same freshly built,
// unwrapped baseline provider's Read pointer for this resource. Because the
// two subtests share nothing but that baseline, removing the wrapper from
// either call site alone -- and only that one -- fails this test: dropping it
// from provider_cluster.go fails only the "cluster" subtest, dropping it from
// provider_namespaced.go fails only "namespaced". Checking a shared helper
// (e.g. withDirectGrantWorkaround itself) instead of both real entrypoints
// would not catch a call site that never invokes it in the first place --
// which is exactly the failure mode this test exists to catch.
func TestDirectGrantWorkaroundInstalledOnBothProviders(t *testing.T) {
	baseline := fmt.Sprintf("%p", tpg.NewProvider("dev", "none")().ResourcesMap[directGrantWorkaroundResourceName].Read) //nolint:staticcheck // SA1019: see above.

	t.Run("cluster", func(t *testing.T) {
		pc, err := GetProvider(context.Background())
		if err != nil {
			t.Fatalf("GetProvider: %v", err)
		}
		got := fmt.Sprintf("%p", pc.TerraformProvider.ResourcesMap[directGrantWorkaroundResourceName].Read) //nolint:staticcheck // SA1019: see above.
		if got == baseline {
			t.Errorf("GetProvider (cluster): %q Read is not wrapped (pointer matches unwrapped baseline: %s) -- is withDirectGrantWorkaround composed at the WithTerraformProvider call site in provider_cluster.go?", directGrantWorkaroundResourceName, got)
		}
	})

	t.Run("namespaced", func(t *testing.T) {
		pc, err := GetProviderNamespaced(context.Background())
		if err != nil {
			t.Fatalf("GetProviderNamespaced: %v", err)
		}
		got := fmt.Sprintf("%p", pc.TerraformProvider.ResourcesMap[directGrantWorkaroundResourceName].Read) //nolint:staticcheck // SA1019: see above.
		if got == baseline {
			t.Errorf("GetProviderNamespaced: %q Read is not wrapped (pointer matches unwrapped baseline: %s) -- is withDirectGrantWorkaround composed at the WithTerraformProvider call site in provider_namespaced.go?", directGrantWorkaroundResourceName, got)
		}
	})
}
