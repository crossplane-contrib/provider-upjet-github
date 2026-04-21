package apis

import (
	"context"
	"testing"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reference"
	upjetresource "github.com/crossplane/upjet/v2/pkg/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestGetManagedResourceResolveReference(t *testing.T) {
	gvk := schema.GroupVersionKind{
		Group:   "repo.github.upbound.io",
		Version: "v1alpha1",
		Kind:    "Repository",
	}

	mg, list, err := GetManagedResource(gvk.Group, gvk.Version, gvk.Kind, "RepositoryList")
	if err != nil {
		t.Fatalf("GetManagedResource(...): %v", err)
	}

	repo := &unstructured.Unstructured{}
	repo.SetGroupVersionKind(gvk)
	repo.SetName("example")
	meta.SetExternalName(repo, "example-repo")

	c := fake.NewClientBuilder().WithRuntimeObjects(repo).Build()
	r := reference.NewAPIResolver(c, newManaged(schema.GroupVersionKind{
		Group:   "team.github.upbound.io",
		Version: "v1alpha1",
		Kind:    "TeamRepository",
	}))

	got, err := r.Resolve(context.Background(), reference.ResolutionRequest{
		Reference: &xpv1.Reference{Name: "example"},
		To:        reference.To{Managed: mg, List: list},
		Extract:   reference.ExternalName(),
	})
	if err != nil {
		t.Fatalf("Resolve(...): %v", err)
	}

	if got.ResolvedValue != "example-repo" {
		t.Errorf("Resolve(...).ResolvedValue: got %q, want %q", got.ResolvedValue, "example-repo")
	}
	if got.ResolvedReference == nil || got.ResolvedReference.Name != "example" {
		t.Errorf("Resolve(...).ResolvedReference: got %#v, want name %q", got.ResolvedReference, "example")
	}
}

func TestGetManagedResourceResolveSelector(t *testing.T) {
	gvk := schema.GroupVersionKind{
		Group:   "repo.github.upbound.io",
		Version: "v1alpha1",
		Kind:    "Repository",
	}

	mg, list, err := GetManagedResource(gvk.Group, gvk.Version, gvk.Kind, "RepositoryList")
	if err != nil {
		t.Fatalf("GetManagedResource(...): %v", err)
	}

	repo := &unstructured.Unstructured{}
	repo.SetGroupVersionKind(gvk)
	repo.SetName("example")
	repo.SetLabels(map[string]string{"repo": "example"})
	meta.SetExternalName(repo, "example-repo")

	c := fake.NewClientBuilder().WithRuntimeObjects(repo).Build()
	r := reference.NewAPIResolver(c, newManaged(schema.GroupVersionKind{
		Group:   "team.github.upbound.io",
		Version: "v1alpha1",
		Kind:    "TeamRepository",
	}))

	got, err := r.Resolve(context.Background(), reference.ResolutionRequest{
		Selector: &xpv1.Selector{MatchLabels: map[string]string{"repo": "example"}},
		To:       reference.To{Managed: mg, List: list},
		Extract:  reference.ExternalName(),
	})
	if err != nil {
		t.Fatalf("Resolve(...): %v", err)
	}

	if got.ResolvedValue != "example-repo" {
		t.Errorf("Resolve(...).ResolvedValue: got %q, want %q", got.ResolvedValue, "example-repo")
	}
	if got.ResolvedReference == nil || got.ResolvedReference.Name != "example" {
		t.Errorf("Resolve(...).ResolvedReference: got %#v, want name %q", got.ResolvedReference, "example")
	}
}

func TestManagedExtractParamPath(t *testing.T) {
	mg := newManaged(schema.GroupVersionKind{
		Group:   "team.github.upbound.io",
		Version: "v1alpha1",
		Kind:    "Team",
	})
	mg.Object = map[string]any{
		"spec": map[string]any{
			"forProvider": map[string]any{
				"name": "team-a",
			},
		},
		"status": map[string]any{
			"atProvider": map[string]any{
				"slug": "team-a",
			},
		},
	}

	if got := upjetresource.ExtractParamPath("name", false)(mg); got != "team-a" {
		t.Errorf("ExtractParamPath(\"name\", false)(...): got %q, want %q", got, "team-a")
	}
	if got := upjetresource.ExtractParamPath("slug", true)(mg); got != "team-a" {
		t.Errorf("ExtractParamPath(\"slug\", true)(...): got %q, want %q", got, "team-a")
	}
}

func TestGetManagedResourceGVK(t *testing.T) {
	mg, list, err := GetManagedResource("repo.github.upbound.io", "v1alpha1", "Repository", "RepositoryList")
	if err != nil {
		t.Fatalf("GetManagedResource(...): %v", err)
	}

	wantManagedGVK := schema.GroupVersionKind{Group: "repo.github.upbound.io", Version: "v1alpha1", Kind: "Repository"}
	if got := mg.(runtime.Object).GetObjectKind().GroupVersionKind(); got != wantManagedGVK {
		t.Errorf("managed GVK: got %s, want %s", got, wantManagedGVK)
	}

	wantListGVK := schema.GroupVersionKind{Group: "repo.github.upbound.io", Version: "v1alpha1", Kind: "RepositoryList"}
	if got := list.(runtime.Object).GetObjectKind().GroupVersionKind(); got != wantListGVK {
		t.Errorf("list GVK: got %s, want %s", got, wantListGVK)
	}
}
