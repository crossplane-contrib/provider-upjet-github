/*
Copyright 2022 Upbound Inc.
*/

// Code generated by upjet. DO NOT EDIT.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	v1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

type EnvironmentVariableInitParameters struct {

	// Name of the environment.
	// Name of the environment.
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-upjet-github/apis/repo/v1alpha1.Environment
	Environment *string `json:"environment,omitempty" tf:"environment,omitempty"`

	// Reference to a Environment in repo to populate environment.
	// +kubebuilder:validation:Optional
	EnvironmentRef *v1.Reference `json:"environmentRef,omitempty" tf:"-"`

	// Selector for a Environment in repo to populate environment.
	// +kubebuilder:validation:Optional
	EnvironmentSelector *v1.Selector `json:"environmentSelector,omitempty" tf:"-"`

	// Name of the repository.
	// Name of the repository.
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-upjet-github/apis/repo/v1alpha1.Repository
	Repository *string `json:"repository,omitempty" tf:"repository,omitempty"`

	// Reference to a Repository in repo to populate repository.
	// +kubebuilder:validation:Optional
	RepositoryRef *v1.Reference `json:"repositoryRef,omitempty" tf:"-"`

	// Selector for a Repository in repo to populate repository.
	// +kubebuilder:validation:Optional
	RepositorySelector *v1.Selector `json:"repositorySelector,omitempty" tf:"-"`

	// Value of the variable
	// Value of the variable.
	Value *string `json:"value,omitempty" tf:"value,omitempty"`

	// Name of the variable.
	// Name of the variable.
	VariableName *string `json:"variableName,omitempty" tf:"variable_name,omitempty"`
}

type EnvironmentVariableObservation struct {

	// Date of actions_environment_secret creation.
	// Date of 'actions_variable' creation.
	CreatedAt *string `json:"createdAt,omitempty" tf:"created_at,omitempty"`

	// Name of the environment.
	// Name of the environment.
	Environment *string `json:"environment,omitempty" tf:"environment,omitempty"`

	ID *string `json:"id,omitempty" tf:"id,omitempty"`

	// Name of the repository.
	// Name of the repository.
	Repository *string `json:"repository,omitempty" tf:"repository,omitempty"`

	// Date of actions_environment_secret update.
	// Date of 'actions_variable' update.
	UpdatedAt *string `json:"updatedAt,omitempty" tf:"updated_at,omitempty"`

	// Value of the variable
	// Value of the variable.
	Value *string `json:"value,omitempty" tf:"value,omitempty"`

	// Name of the variable.
	// Name of the variable.
	VariableName *string `json:"variableName,omitempty" tf:"variable_name,omitempty"`
}

type EnvironmentVariableParameters struct {

	// Name of the environment.
	// Name of the environment.
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-upjet-github/apis/repo/v1alpha1.Environment
	// +kubebuilder:validation:Optional
	Environment *string `json:"environment,omitempty" tf:"environment,omitempty"`

	// Reference to a Environment in repo to populate environment.
	// +kubebuilder:validation:Optional
	EnvironmentRef *v1.Reference `json:"environmentRef,omitempty" tf:"-"`

	// Selector for a Environment in repo to populate environment.
	// +kubebuilder:validation:Optional
	EnvironmentSelector *v1.Selector `json:"environmentSelector,omitempty" tf:"-"`

	// Name of the repository.
	// Name of the repository.
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-upjet-github/apis/repo/v1alpha1.Repository
	// +kubebuilder:validation:Optional
	Repository *string `json:"repository,omitempty" tf:"repository,omitempty"`

	// Reference to a Repository in repo to populate repository.
	// +kubebuilder:validation:Optional
	RepositoryRef *v1.Reference `json:"repositoryRef,omitempty" tf:"-"`

	// Selector for a Repository in repo to populate repository.
	// +kubebuilder:validation:Optional
	RepositorySelector *v1.Selector `json:"repositorySelector,omitempty" tf:"-"`

	// Value of the variable
	// Value of the variable.
	// +kubebuilder:validation:Optional
	Value *string `json:"value,omitempty" tf:"value,omitempty"`

	// Name of the variable.
	// Name of the variable.
	// +kubebuilder:validation:Optional
	VariableName *string `json:"variableName,omitempty" tf:"variable_name,omitempty"`
}

// EnvironmentVariableSpec defines the desired state of EnvironmentVariable
type EnvironmentVariableSpec struct {
	v1.ResourceSpec `json:",inline"`
	ForProvider     EnvironmentVariableParameters `json:"forProvider"`
	// THIS IS A BETA FIELD. It will be honored
	// unless the Management Policies feature flag is disabled.
	// InitProvider holds the same fields as ForProvider, with the exception
	// of Identifier and other resource reference fields. The fields that are
	// in InitProvider are merged into ForProvider when the resource is created.
	// The same fields are also added to the terraform ignore_changes hook, to
	// avoid updating them after creation. This is useful for fields that are
	// required on creation, but we do not desire to update them after creation,
	// for example because of an external controller is managing them, like an
	// autoscaler.
	InitProvider EnvironmentVariableInitParameters `json:"initProvider,omitempty"`
}

// EnvironmentVariableStatus defines the observed state of EnvironmentVariable.
type EnvironmentVariableStatus struct {
	v1.ResourceStatus `json:",inline"`
	AtProvider        EnvironmentVariableObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// EnvironmentVariable is the Schema for the EnvironmentVariables API. Creates and manages an Action variable within a GitHub repository environment
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,github}
type EnvironmentVariable struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// +kubebuilder:validation:XValidation:rule="!('*' in self.managementPolicies || 'Create' in self.managementPolicies || 'Update' in self.managementPolicies) || has(self.forProvider.value) || (has(self.initProvider) && has(self.initProvider.value))",message="spec.forProvider.value is a required parameter"
	// +kubebuilder:validation:XValidation:rule="!('*' in self.managementPolicies || 'Create' in self.managementPolicies || 'Update' in self.managementPolicies) || has(self.forProvider.variableName) || (has(self.initProvider) && has(self.initProvider.variableName))",message="spec.forProvider.variableName is a required parameter"
	Spec   EnvironmentVariableSpec   `json:"spec"`
	Status EnvironmentVariableStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// EnvironmentVariableList contains a list of EnvironmentVariables
type EnvironmentVariableList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EnvironmentVariable `json:"items"`
}

// Repository type metadata.
var (
	EnvironmentVariable_Kind             = "EnvironmentVariable"
	EnvironmentVariable_GroupKind        = schema.GroupKind{Group: CRDGroup, Kind: EnvironmentVariable_Kind}.String()
	EnvironmentVariable_KindAPIVersion   = EnvironmentVariable_Kind + "." + CRDGroupVersion.String()
	EnvironmentVariable_GroupVersionKind = CRDGroupVersion.WithKind(EnvironmentVariable_Kind)
)

func init() {
	SchemeBuilder.Register(&EnvironmentVariable{}, &EnvironmentVariableList{})
}
