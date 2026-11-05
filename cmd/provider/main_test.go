package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"

	apisCluster "github.com/crossplane-contrib/provider-upjet-github/apis/cluster"
	apisClusterv1beta1 "github.com/crossplane-contrib/provider-upjet-github/apis/cluster/v1beta1"
	apisNamespaced "github.com/crossplane-contrib/provider-upjet-github/apis/namespaced"
	apisNamespacedv1beta1 "github.com/crossplane-contrib/provider-upjet-github/apis/namespaced/v1beta1"
)

// A group missing from the CRD gate means its controllers never start.
func TestProviderGroups(t *testing.T) {
	s := runtime.NewScheme()
	require.NoError(t, apisCluster.AddToScheme(s))
	require.NoError(t, apisNamespaced.AddToScheme(s))
	require.NoError(t, apiextensionsv1.AddToScheme(s))

	got := providerGroups(s, apisClusterv1beta1.Group, apisNamespacedv1beta1.Group)

	want := map[string]struct{}{}
	for gvk := range s.AllKnownTypes() {
		if gvk.Group != apiextensionsv1.GroupName && gvk.Group != "" {
			want[gvk.Group] = struct{}{}
		}
	}
	assert.Len(t, got, len(want))
	assert.Contains(t, got, "github.upbound.io")
	assert.Contains(t, got, "repo.github.m.upbound.io")
	assert.NotContains(t, got, apiextensionsv1.GroupName)
}
