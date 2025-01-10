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

type OrganizationActionsSecretInitParameters struct {

	// Encrypted value of the secret using the GitHub public key in Base64 format.
	// Encrypted value of the secret using the GitHub public key in Base64 format.
	EncryptedValueSecretRef *v1.SecretKeySelector `json:"encryptedValueSecretRef,omitempty" tf:"-"`

	// Plaintext value of the secret to be encrypted
	// Plaintext value of the secret to be encrypted.
	PlaintextValueSecretRef *v1.SecretKeySelector `json:"plaintextValueSecretRef,omitempty" tf:"-"`

	// Name of the secret
	// Name of the secret.
	SecretName *string `json:"secretName,omitempty" tf:"secret_name,omitempty"`

	// An array of repository ids that can access the organization secret.
	// An array of repository ids that can access the organization secret.
	// +listType=set
	SelectedRepositoryIds []*int64 `json:"selectedRepositoryIds,omitempty" tf:"selected_repository_ids,omitempty"`

	// Configures the access that repositories have to the organization secret.
	// Must be one of all, private, selected. selected_repository_ids is required if set to selected.
	// Configures the access that repositories have to the organization secret. Must be one of 'all', 'private', or 'selected'. 'selected_repository_ids' is required if set to 'selected'.
	Visibility *string `json:"visibility,omitempty" tf:"visibility,omitempty"`
}

type OrganizationActionsSecretObservation struct {

	// Date of actions_secret creation.
	// Date of 'actions_secret' creation.
	CreatedAt *string `json:"createdAt,omitempty" tf:"created_at,omitempty"`

	ID *string `json:"id,omitempty" tf:"id,omitempty"`

	// Name of the secret
	// Name of the secret.
	SecretName *string `json:"secretName,omitempty" tf:"secret_name,omitempty"`

	// An array of repository ids that can access the organization secret.
	// An array of repository ids that can access the organization secret.
	// +listType=set
	SelectedRepositoryIds []*int64 `json:"selectedRepositoryIds,omitempty" tf:"selected_repository_ids,omitempty"`

	// Date of actions_secret update.
	// Date of 'actions_secret' update.
	UpdatedAt *string `json:"updatedAt,omitempty" tf:"updated_at,omitempty"`

	// Configures the access that repositories have to the organization secret.
	// Must be one of all, private, selected. selected_repository_ids is required if set to selected.
	// Configures the access that repositories have to the organization secret. Must be one of 'all', 'private', or 'selected'. 'selected_repository_ids' is required if set to 'selected'.
	Visibility *string `json:"visibility,omitempty" tf:"visibility,omitempty"`
}

type OrganizationActionsSecretParameters struct {

	// Encrypted value of the secret using the GitHub public key in Base64 format.
	// Encrypted value of the secret using the GitHub public key in Base64 format.
	// +kubebuilder:validation:Optional
	EncryptedValueSecretRef *v1.SecretKeySelector `json:"encryptedValueSecretRef,omitempty" tf:"-"`

	// Plaintext value of the secret to be encrypted
	// Plaintext value of the secret to be encrypted.
	// +kubebuilder:validation:Optional
	PlaintextValueSecretRef *v1.SecretKeySelector `json:"plaintextValueSecretRef,omitempty" tf:"-"`

	// Name of the secret
	// Name of the secret.
	// +kubebuilder:validation:Optional
	SecretName *string `json:"secretName,omitempty" tf:"secret_name,omitempty"`

	// An array of repository ids that can access the organization secret.
	// An array of repository ids that can access the organization secret.
	// +kubebuilder:validation:Optional
	// +listType=set
	SelectedRepositoryIds []*int64 `json:"selectedRepositoryIds,omitempty" tf:"selected_repository_ids,omitempty"`

	// Configures the access that repositories have to the organization secret.
	// Must be one of all, private, selected. selected_repository_ids is required if set to selected.
	// Configures the access that repositories have to the organization secret. Must be one of 'all', 'private', or 'selected'. 'selected_repository_ids' is required if set to 'selected'.
	// +kubebuilder:validation:Optional
	Visibility *string `json:"visibility,omitempty" tf:"visibility,omitempty"`
}

// OrganizationActionsSecretSpec defines the desired state of OrganizationActionsSecret
type OrganizationActionsSecretSpec struct {
	v1.ResourceSpec `json:",inline"`
	ForProvider     OrganizationActionsSecretParameters `json:"forProvider"`
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
	InitProvider OrganizationActionsSecretInitParameters `json:"initProvider,omitempty"`
}

// OrganizationActionsSecretStatus defines the observed state of OrganizationActionsSecret.
type OrganizationActionsSecretStatus struct {
	v1.ResourceStatus `json:",inline"`
	AtProvider        OrganizationActionsSecretObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// OrganizationActionsSecret is the Schema for the OrganizationActionsSecrets API. Creates and manages an Action Secret within a GitHub organization
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,github}
type OrganizationActionsSecret struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// +kubebuilder:validation:XValidation:rule="!('*' in self.managementPolicies || 'Create' in self.managementPolicies || 'Update' in self.managementPolicies) || has(self.forProvider.secretName) || (has(self.initProvider) && has(self.initProvider.secretName))",message="spec.forProvider.secretName is a required parameter"
	// +kubebuilder:validation:XValidation:rule="!('*' in self.managementPolicies || 'Create' in self.managementPolicies || 'Update' in self.managementPolicies) || has(self.forProvider.visibility) || (has(self.initProvider) && has(self.initProvider.visibility))",message="spec.forProvider.visibility is a required parameter"
	Spec   OrganizationActionsSecretSpec   `json:"spec"`
	Status OrganizationActionsSecretStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// OrganizationActionsSecretList contains a list of OrganizationActionsSecrets
type OrganizationActionsSecretList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OrganizationActionsSecret `json:"items"`
}

// Repository type metadata.
var (
	OrganizationActionsSecret_Kind             = "OrganizationActionsSecret"
	OrganizationActionsSecret_GroupKind        = schema.GroupKind{Group: CRDGroup, Kind: OrganizationActionsSecret_Kind}.String()
	OrganizationActionsSecret_KindAPIVersion   = OrganizationActionsSecret_Kind + "." + CRDGroupVersion.String()
	OrganizationActionsSecret_GroupVersionKind = CRDGroupVersion.WithKind(OrganizationActionsSecret_Kind)
)

func init() {
	SchemeBuilder.Register(&OrganizationActionsSecret{}, &OrganizationActionsSecretList{})
}