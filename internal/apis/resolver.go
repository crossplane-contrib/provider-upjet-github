package apis

import (
	v1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/fieldpath"
	xpresource "github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	upjetresource "github.com/crossplane/upjet/v2/pkg/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var _ xpresource.Managed = &managed{}
var _ xpresource.ManagedList = &managedList{}
var _ upjetresource.Terraformed = &managed{}

// GetManagedResource returns lightweight unstructured managed resource shells
// for generated reference resolvers. It avoids static imports between API
// packages while still giving controller-runtime a concrete GVK to read/list.
func GetManagedResource(group, version, kind, listKind string) (xpresource.Managed, xpresource.ManagedList, error) {
	return newManaged(schema.GroupVersionKind{
			Group:   group,
			Version: version,
			Kind:    kind,
		}),
		newManagedList(schema.GroupVersionKind{
			Group:   group,
			Version: version,
			Kind:    listKind,
		}),
		nil
}

type managed struct {
	unstructured.Unstructured
}

func newManaged(gvk schema.GroupVersionKind) *managed {
	m := &managed{Unstructured: unstructured.Unstructured{Object: map[string]any{}}}
	m.SetGroupVersionKind(gvk)
	return m
}

func (m *managed) GetCondition(ct v1.ConditionType) v1.Condition {
	status := &v1.ConditionedStatus{}
	if err := fieldpath.Pave(m.object()).GetValueInto("status", status); err != nil {
		return v1.Condition{}
	}
	return status.GetCondition(ct)
}

func (m *managed) SetConditions(conditions ...v1.Condition) {
	status := &v1.ConditionedStatus{}
	_ = fieldpath.Pave(m.object()).GetValueInto("status", status)
	status.SetConditions(conditions...)
	_ = fieldpath.Pave(m.object()).SetValue("status.conditions", status.Conditions)
}

func (m *managed) GetManagementPolicies() v1.ManagementPolicies {
	var policies v1.ManagementPolicies
	_ = fieldpath.Pave(m.object()).GetValueInto("spec.managementPolicies", &policies)
	return policies
}

func (m *managed) SetManagementPolicies(policies v1.ManagementPolicies) {
	_ = fieldpath.Pave(m.object()).SetValue("spec.managementPolicies", policies)
}

func (m *managed) GetObservation() (map[string]any, error) {
	return m.getMap("status.atProvider"), nil
}

func (m *managed) SetObservation(obs map[string]any) error {
	return fieldpath.Pave(m.object()).SetValue("status.atProvider", obs)
}

func (m *managed) GetID() string {
	id, err := fieldpath.Pave(m.object()).GetString("status.atProvider.id")
	if err != nil {
		return ""
	}
	return id
}

func (m *managed) GetParameters() (map[string]any, error) {
	return m.getMap("spec.forProvider"), nil
}

func (m *managed) SetParameters(params map[string]any) error {
	return fieldpath.Pave(m.object()).SetValue("spec.forProvider", params)
}

func (m *managed) GetInitParameters() (map[string]any, error) {
	return m.getMap("spec.initProvider"), nil
}

func (m *managed) GetMergedParameters(mergeInitProvider bool) (map[string]any, error) {
	params := m.getMap("spec.forProvider")
	if !mergeInitProvider {
		return params, nil
	}
	for k, v := range m.getMap("spec.initProvider") {
		if _, ok := params[k]; !ok {
			params[k] = v
		}
	}
	return params, nil
}

func (m *managed) GetTerraformResourceType() string {
	return ""
}

func (m *managed) GetTerraformSchemaVersion() int {
	return 0
}

func (m *managed) GetConnectionDetailsMapping() map[string]string {
	return nil
}

func (m *managed) LateInitialize([]byte) (bool, error) {
	return false, nil
}

func (m *managed) object() map[string]any {
	if m.Object == nil {
		m.Object = map[string]any{}
	}
	return m.Object
}

func (m *managed) getMap(path string) map[string]any {
	values := map[string]any{}
	if err := fieldpath.Pave(m.object()).GetValueInto(path, &values); err != nil {
		return map[string]any{}
	}
	return values
}

type managedList struct {
	unstructured.UnstructuredList
}

func newManagedList(gvk schema.GroupVersionKind) *managedList {
	l := &managedList{UnstructuredList: unstructured.UnstructuredList{Object: map[string]any{}}}
	l.SetGroupVersionKind(gvk)
	return l
}

func (l *managedList) GetItems() []xpresource.Managed {
	items := make([]xpresource.Managed, len(l.Items))
	for i := range l.Items {
		items[i] = &managed{Unstructured: l.Items[i]}
	}
	return items
}
