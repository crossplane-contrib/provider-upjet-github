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

type EnvironmentDeploymentPolicyInitParameters struct {

	// The name pattern that branches must match in order to deploy to the environment. If not specified, tag_pattern must be specified.
	// The name pattern that branches must match in order to deploy to the environment.
	BranchPattern *string `json:"branchPattern,omitempty" tf:"branch_pattern,omitempty"`

	// The name of the environment.
	// The name of the environment.
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-upjet-github/apis/repo/v1alpha1.Environment
	Environment *string `json:"environment,omitempty" tf:"environment,omitempty"`

	// Reference to a Environment in repo to populate environment.
	// +kubebuilder:validation:Optional
	EnvironmentRef *v1.Reference `json:"environmentRef,omitempty" tf:"-"`

	// Selector for a Environment in repo to populate environment.
	// +kubebuilder:validation:Optional
	EnvironmentSelector *v1.Selector `json:"environmentSelector,omitempty" tf:"-"`

	// The repository of the environment.
	// The name of the repository. The name is not case sensitive.
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-upjet-github/apis/repo/v1alpha1.Repository
	Repository *string `json:"repository,omitempty" tf:"repository,omitempty"`

	// Reference to a Repository in repo to populate repository.
	// +kubebuilder:validation:Optional
	RepositoryRef *v1.Reference `json:"repositoryRef,omitempty" tf:"-"`

	// Selector for a Repository in repo to populate repository.
	// +kubebuilder:validation:Optional
	RepositorySelector *v1.Selector `json:"repositorySelector,omitempty" tf:"-"`

	// The name pattern that tags must match in order to deploy to the environment. If not specified, branch_pattern must be specified.
	// The name pattern that tags must match in order to deploy to the environment.
	TagPattern *string `json:"tagPattern,omitempty" tf:"tag_pattern,omitempty"`
}

type EnvironmentDeploymentPolicyObservation struct {

	// The name pattern that branches must match in order to deploy to the environment. If not specified, tag_pattern must be specified.
	// The name pattern that branches must match in order to deploy to the environment.
	BranchPattern *string `json:"branchPattern,omitempty" tf:"branch_pattern,omitempty"`

	// The name of the environment.
	// The name of the environment.
	Environment *string `json:"environment,omitempty" tf:"environment,omitempty"`

	ID *string `json:"id,omitempty" tf:"id,omitempty"`

	// The repository of the environment.
	// The name of the repository. The name is not case sensitive.
	Repository *string `json:"repository,omitempty" tf:"repository,omitempty"`

	// The name pattern that tags must match in order to deploy to the environment. If not specified, branch_pattern must be specified.
	// The name pattern that tags must match in order to deploy to the environment.
	TagPattern *string `json:"tagPattern,omitempty" tf:"tag_pattern,omitempty"`
}

type EnvironmentDeploymentPolicyParameters struct {

	// The name pattern that branches must match in order to deploy to the environment. If not specified, tag_pattern must be specified.
	// The name pattern that branches must match in order to deploy to the environment.
	// +kubebuilder:validation:Optional
	BranchPattern *string `json:"branchPattern,omitempty" tf:"branch_pattern,omitempty"`

	// The name of the environment.
	// The name of the environment.
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-upjet-github/apis/repo/v1alpha1.Environment
	// +kubebuilder:validation:Optional
	Environment *string `json:"environment,omitempty" tf:"environment,omitempty"`

	// Reference to a Environment in repo to populate environment.
	// +kubebuilder:validation:Optional
	EnvironmentRef *v1.Reference `json:"environmentRef,omitempty" tf:"-"`

	// Selector for a Environment in repo to populate environment.
	// +kubebuilder:validation:Optional
	EnvironmentSelector *v1.Selector `json:"environmentSelector,omitempty" tf:"-"`

	// The repository of the environment.
	// The name of the repository. The name is not case sensitive.
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-upjet-github/apis/repo/v1alpha1.Repository
	// +kubebuilder:validation:Optional
	Repository *string `json:"repository,omitempty" tf:"repository,omitempty"`

	// Reference to a Repository in repo to populate repository.
	// +kubebuilder:validation:Optional
	RepositoryRef *v1.Reference `json:"repositoryRef,omitempty" tf:"-"`

	// Selector for a Repository in repo to populate repository.
	// +kubebuilder:validation:Optional
	RepositorySelector *v1.Selector `json:"repositorySelector,omitempty" tf:"-"`

	// The name pattern that tags must match in order to deploy to the environment. If not specified, branch_pattern must be specified.
	// The name pattern that tags must match in order to deploy to the environment.
	// +kubebuilder:validation:Optional
	TagPattern *string `json:"tagPattern,omitempty" tf:"tag_pattern,omitempty"`
}

// EnvironmentDeploymentPolicySpec defines the desired state of EnvironmentDeploymentPolicy
type EnvironmentDeploymentPolicySpec struct {
	v1.ResourceSpec `json:",inline"`
	ForProvider     EnvironmentDeploymentPolicyParameters `json:"forProvider"`
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
	InitProvider EnvironmentDeploymentPolicyInitParameters `json:"initProvider,omitempty"`
}

// EnvironmentDeploymentPolicyStatus defines the observed state of EnvironmentDeploymentPolicy.
type EnvironmentDeploymentPolicyStatus struct {
	v1.ResourceStatus `json:",inline"`
	AtProvider        EnvironmentDeploymentPolicyObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// EnvironmentDeploymentPolicy is the Schema for the EnvironmentDeploymentPolicys API. Creates and manages environment deployment branch policies for GitHub repositories
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,github}
type EnvironmentDeploymentPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              EnvironmentDeploymentPolicySpec   `json:"spec"`
	Status            EnvironmentDeploymentPolicyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// EnvironmentDeploymentPolicyList contains a list of EnvironmentDeploymentPolicys
type EnvironmentDeploymentPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EnvironmentDeploymentPolicy `json:"items"`
}

// Repository type metadata.
var (
	EnvironmentDeploymentPolicy_Kind             = "EnvironmentDeploymentPolicy"
	EnvironmentDeploymentPolicy_GroupKind        = schema.GroupKind{Group: CRDGroup, Kind: EnvironmentDeploymentPolicy_Kind}.String()
	EnvironmentDeploymentPolicy_KindAPIVersion   = EnvironmentDeploymentPolicy_Kind + "." + CRDGroupVersion.String()
	EnvironmentDeploymentPolicy_GroupVersionKind = CRDGroupVersion.WithKind(EnvironmentDeploymentPolicy_Kind)
)

func init() {
	SchemeBuilder.Register(&EnvironmentDeploymentPolicy{}, &EnvironmentDeploymentPolicyList{})
}
