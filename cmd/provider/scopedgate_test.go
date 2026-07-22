package main

import (
	"context"
	"testing"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/crossplane/crossplane-runtime/v2/pkg/logging"
)

// fakeGate records the last value set for each GVK so tests can assert on them.
type fakeGate struct {
	set map[schema.GroupVersionKind]bool
}

func (g *fakeGate) Register(_ func(), _ ...schema.GroupVersionKind) {}

func (g *fakeGate) Set(gvk schema.GroupVersionKind, ready bool) bool {
	if g.set == nil {
		g.set = map[schema.GroupVersionKind]bool{}
	}
	g.set[gvk] = ready
	return true
}

func TestMatchPattern(t *testing.T) {
	cases := map[string]struct {
		pattern string
		s       string
		want    bool
	}{
		"Exact":          {"team.github.upbound.io", "team.github.upbound.io", true},
		"ExactNoMatch":   {"team.github.upbound.io", "repo.github.upbound.io", false},
		"Star":           {"*", "anything", true},
		"SuffixMatch":    {"*.github.upbound.io", "team.github.upbound.io", true},
		"SuffixNoMatch":  {"*.github.upbound.io", "team.aws.upbound.io", false},
		"PrefixMatch":    {"team.*", "team.github.upbound.io", true},
		"PrefixNoMatch":  {"team.*", "repo.github.upbound.io", false},
		"NoWildcardDiff": {"team", "team.github.upbound.io", false},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := matchPattern(tc.pattern, tc.s); got != tc.want {
				t.Errorf("matchPattern(%q, %q) = %v, want %v", tc.pattern, tc.s, got, tc.want)
			}
		})
	}
}

func TestGVKActive(t *testing.T) {
	gvk := schema.GroupVersionKind{Group: "team.github.upbound.io", Version: "v1alpha1", Kind: "Team"}
	cases := map[string]struct {
		patterns []string
		want     bool
	}{
		"EmptyActivatesAll":  {nil, true},
		"MatchWholeGroup":    {[]string{"team.github.upbound.io"}, true},
		"MatchGroupKind":     {[]string{"team.github.upbound.io/Team"}, true},
		"MatchGroupSuffix":   {[]string{"*.github.upbound.io"}, true},
		"NoMatchOtherGroup":  {[]string{"repo.github.upbound.io"}, false},
		"NoMatchOtherKind":   {[]string{"team.github.upbound.io/Membership"}, false},
		"MatchAmongMultiple": {[]string{"repo.github.upbound.io", "team.*"}, true},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := gvkActive(gvk, tc.patterns); got != tc.want {
				t.Errorf("gvkActive(%v, %v) = %v, want %v", gvk, tc.patterns, got, tc.want)
			}
		})
	}
}

func TestParseCSV(t *testing.T) {
	cases := map[string]struct {
		in   string
		want []string
	}{
		"Empty":          {"", nil},
		"Single":         {"a", []string{"a"}},
		"Multiple":       {"a,b,c", []string{"a", "b", "c"}},
		"TrimsSpaces":    {" a , b ,c ", []string{"a", "b", "c"}},
		"DropsEmpties":   {"a,,b,", []string{"a", "b"}},
		"OnlyCommas":     {",,,", nil},
		"OnlyWhitespace": {"  ,  ", nil},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := parseCSV(tc.in)
			if len(got) != len(tc.want) {
				t.Fatalf("parseCSV(%q) = %v, want %v", tc.in, got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("parseCSV(%q)[%d] = %q, want %q", tc.in, i, got[i], tc.want[i])
				}
			}
		})
	}
}

func TestCRDEstablished(t *testing.T) {
	cases := map[string]struct {
		conditions []apiextensionsv1.CustomResourceDefinitionCondition
		want       bool
	}{
		"Established": {[]apiextensionsv1.CustomResourceDefinitionCondition{
			{Type: apiextensionsv1.Established, Status: apiextensionsv1.ConditionTrue},
		}, true},
		"NotEstablished": {[]apiextensionsv1.CustomResourceDefinitionCondition{
			{Type: apiextensionsv1.Established, Status: apiextensionsv1.ConditionFalse},
		}, false},
		"NoConditions": {nil, false},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			crd := &apiextensionsv1.CustomResourceDefinition{
				Status: apiextensionsv1.CustomResourceDefinitionStatus{Conditions: tc.conditions},
			}
			if got := crdEstablished(crd); got != tc.want {
				t.Errorf("crdEstablished() = %v, want %v", got, tc.want)
			}
		})
	}
}

// crd builds a CRD with the given group/kind, served versions and established state.
func crd(group, kind string, established bool, versions ...apiextensionsv1.CustomResourceDefinitionVersion) *apiextensionsv1.CustomResourceDefinition {
	status := apiextensionsv1.ConditionFalse
	if established {
		status = apiextensionsv1.ConditionTrue
	}
	return &apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{Name: kind + "." + group},
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group:    group,
			Names:    apiextensionsv1.CustomResourceDefinitionNames{Kind: kind},
			Versions: versions,
		},
		Status: apiextensionsv1.CustomResourceDefinitionStatus{
			Conditions: []apiextensionsv1.CustomResourceDefinitionCondition{
				{Type: apiextensionsv1.Established, Status: status},
			},
		},
	}
}

func TestScopedGateReconcile(t *testing.T) {
	served := apiextensionsv1.CustomResourceDefinitionVersion{Name: "v1alpha1", Served: true}
	notServed := apiextensionsv1.CustomResourceDefinitionVersion{Name: "v1beta1", Served: false}

	teamGVK := schema.GroupVersionKind{Group: "team.github.upbound.io", Version: "v1alpha1", Kind: "Team"}
	repoGVK := schema.GroupVersionKind{Group: "repo.github.upbound.io", Version: "v1alpha1", Kind: "Repository"}

	cases := map[string]struct {
		crd          *apiextensionsv1.CustomResourceDefinition
		activeGroups []string
		want         map[schema.GroupVersionKind]bool
	}{
		"EstablishedServedNoFilter": {
			crd:  crd("team.github.upbound.io", "Team", true, served),
			want: map[schema.GroupVersionKind]bool{teamGVK: true},
		},
		"NotEstablishedClosesGate": {
			crd:  crd("team.github.upbound.io", "Team", false, served),
			want: map[schema.GroupVersionKind]bool{teamGVK: false},
		},
		"NotServedLeftUntouched": {
			// Not served => gate is neither opened nor explicitly set true.
			crd:  crd("team.github.upbound.io", "Team", true, notServed),
			want: map[schema.GroupVersionKind]bool{},
		},
		"FilteredOutStaysClosed": {
			crd:          crd("repo.github.upbound.io", "Repository", true, served),
			activeGroups: []string{"team.github.upbound.io"},
			want:         map[schema.GroupVersionKind]bool{},
		},
		"FilterMatchesOpensGate": {
			crd:          crd("repo.github.upbound.io", "Repository", true, served),
			activeGroups: []string{"repo.*"},
			want:         map[schema.GroupVersionKind]bool{repoGVK: true},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			gate := &fakeGate{}
			r := &scopedGateReconciler{log: logging.NewNopLogger(), gate: gate, activeGroups: tc.activeGroups}

			if _, err := r.Reconcile(context.Background(), tc.crd); err != nil {
				t.Fatalf("Reconcile() returned unexpected error: %v", err)
			}
			if len(gate.set) != len(tc.want) {
				t.Fatalf("gate entries = %v, want %v", gate.set, tc.want)
			}
			for gvk, want := range tc.want {
				if got, ok := gate.set[gvk]; !ok || got != want {
					t.Errorf("gate[%s] = %v (present=%v), want %v", gvk, got, ok, want)
				}
			}
		})
	}
}
