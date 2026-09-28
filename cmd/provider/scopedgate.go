package main

import (
	"context"
	"errors"
	"reflect"
	"strings"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/crossplane/crossplane-runtime/v2/pkg/logging"
)

// setupScopedCRDGate registers a controller that reconciles
// CustomResourceDefinitions and opens the supplied gate for a GVK once its CRD
// is Established and served.
//
// It behaves like crossplane-runtime's customresourcesgate, with one addition:
// when activeGroups is non-empty it only opens the gate for GVKs whose API group
// matches one of the supplied patterns. GVKs that do not match are never gated
// open, so their controllers are never started and never create an informer.
// This avoids waiting on (and timing out on) cache syncs for resources the
// provider is not meant to serve on this control plane.
//
// When activeGroups is empty the behaviour is identical to the upstream gate
// (every Established+served GVK is opened), so this is a no-op by default and
// safe to always use.
func setupScopedCRDGate(mgr ctrl.Manager, o controller.Options, activeGroups []string) error {
	if o.Gate == nil || reflect.ValueOf(o.Gate).IsNil() {
		return errors.New("gate is required")
	}

	r := &scopedGateReconciler{
		log:          o.Logger,
		gate:         o.Gate,
		activeGroups: activeGroups,
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&apiextensionsv1.CustomResourceDefinition{}).
		Named("crd-gate").
		Complete(reconcile.AsReconciler[*apiextensionsv1.CustomResourceDefinition](mgr.GetClient(), r))
}

// scopedGateReconciler reports GVK readiness to a Gate, optionally restricted to
// a set of API group patterns.
type scopedGateReconciler struct {
	log          logging.Logger
	gate         controller.Gate
	activeGroups []string
}

// Reconcile reports ready/unready GVKs of a CustomResourceDefinition to the gate.
func (r *scopedGateReconciler) Reconcile(_ context.Context, crd *apiextensionsv1.CustomResourceDefinition) (ctrl.Result, error) {
	established := crdEstablished(crd)
	deleting := !crd.GetDeletionTimestamp().IsZero()

	for _, v := range crd.Spec.Versions {
		gvk := schema.GroupVersionKind{Group: crd.Spec.Group, Version: v.Name, Kind: crd.Spec.Names.Kind}
		switch {
		case !established || deleting:
			r.gate.Set(gvk, false)
		case v.Served && gvkActive(gvk, r.activeGroups):
			r.log.Debug("gvk is ready", "gvk", gvk)
			r.gate.Set(gvk, true)
		default:
			// Either the version is not served, or the GVK is excluded by the
			// active-resources filter. Leave the gate closed so the controller
			// stays inert (no informer, no cache sync, no cache-sync timeout).
			r.log.Debug("gvk gated off (not served or filtered out by active-resources)", "gvk", gvk)
		}
	}

	return ctrl.Result{}, nil
}

func crdEstablished(crd *apiextensionsv1.CustomResourceDefinition) bool {
	for _, c := range crd.Status.Conditions {
		if c.Type == apiextensionsv1.Established {
			return c.Status == apiextensionsv1.ConditionTrue
		}
	}
	return false
}

// gvkActive reports whether a GVK should be activated. An empty patterns slice
// activates everything (upstream behaviour). Each pattern is matched against
// both the API group (e.g. "team.github.m.upbound.io") and the group/Kind
// (e.g. "team.github.m.upbound.io/Team"), so callers can scope by whole group
// or by individual kind. A pattern supports a single leading or trailing '*'
// wildcard (suffix/prefix match); otherwise it is an exact match.
func gvkActive(gvk schema.GroupVersionKind, patterns []string) bool {
	if len(patterns) == 0 {
		return true
	}
	groupKind := gvk.Group + "/" + gvk.Kind
	for _, p := range patterns {
		if matchPattern(p, gvk.Group) || matchPattern(p, groupKind) {
			return true
		}
	}
	return false
}

func matchPattern(pattern, s string) bool {
	switch {
	case pattern == "*" || pattern == s:
		return true
	case strings.HasPrefix(pattern, "*"):
		return strings.HasSuffix(s, strings.TrimPrefix(pattern, "*"))
	case strings.HasSuffix(pattern, "*"):
		return strings.HasPrefix(s, strings.TrimSuffix(pattern, "*"))
	default:
		return false
	}
}

// parseCSV splits a comma-separated flag value into a trimmed, non-empty slice.
func parseCSV(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if p := strings.TrimSpace(part); p != "" {
			out = append(out, p)
		}
	}
	return out
}
