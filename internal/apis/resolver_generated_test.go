package apis_test

import (
	"context"
	"testing"

	clusterteam "github.com/crossplane-contrib/provider-upjet-github/apis/cluster/team/v1alpha1"
	namespacedteam "github.com/crossplane-contrib/provider-upjet-github/apis/namespaced/team/v1alpha1"
	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// Exercise the generated entry points, including namespace isolation, rather
// than only testing the unstructured resource adapter in isolation.
func TestGeneratedTeamRepositoryReferences(t *testing.T) {
	const (
		repositoryName = "resolver-test-repository"
		externalName   = "resolved-repository"
		namespace      = "resolver-test"
		version        = "v1alpha1"
	)
	for _, namespaced := range []bool{false, true} {
		for _, selector := range []bool{false, true} {
			name := "cluster/reference"
			group := "repo.github.upbound.io"
			if namespaced {
				name = "namespaced/reference"
				group = "repo.github.m.upbound.io"
			}
			if selector {
				name = name[:len(name)-len("reference")] + "selector"
			}
			t.Run(name, func(t *testing.T) {
				repo := &unstructured.Unstructured{}
				repo.SetGroupVersionKind(schema.GroupVersionKind{Group: group, Version: version, Kind: "Repository"})
				repo.SetName(repositoryName)
				repo.SetLabels(map[string]string{"test": repositoryName})
				meta.SetExternalName(repo, externalName)
				teamID := "123"
				var got *string
				if namespaced {
					repo.SetNamespace(namespace)
					decoy := repo.DeepCopy()
					decoy.SetNamespace("other-namespace")
					meta.SetExternalName(decoy, "wrong-namespace-repository")
					c := fake.NewClientBuilder().WithObjects(repo, decoy).Build()
					mg := &namespacedteam.TeamRepository{}
					mg.SetNamespace(namespace)
					mg.Spec.ForProvider.TeamID = &teamID
					if selector {
						mg.Spec.ForProvider.RepositorySelector = &xpv1.NamespacedSelector{MatchLabels: repo.GetLabels()}
					} else {
						mg.Spec.ForProvider.RepositoryRef = &xpv1.NamespacedReference{Name: repositoryName}
					}
					if err := mg.ResolveReferences(context.Background(), c); err != nil {
						t.Fatalf("ResolveReferences(): %v", err)
					}
					got = mg.Spec.ForProvider.Repository
				} else {
					c := fake.NewClientBuilder().WithObjects(repo).Build()
					mg := &clusterteam.TeamRepository{}
					mg.Spec.ForProvider.TeamID = &teamID
					if selector {
						mg.Spec.ForProvider.RepositorySelector = &xpv1.Selector{MatchLabels: repo.GetLabels()}
					} else {
						mg.Spec.ForProvider.RepositoryRef = &xpv1.Reference{Name: repositoryName}
					}
					if err := mg.ResolveReferences(context.Background(), c); err != nil {
						t.Fatalf("ResolveReferences(): %v", err)
					}
					got = mg.Spec.ForProvider.Repository
				}
				if got == nil || *got != externalName {
					t.Errorf("resolved repository = %v, want %q", got, externalName)
				}
			})
		}
	}
}
