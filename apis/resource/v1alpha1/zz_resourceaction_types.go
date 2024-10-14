// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// Code generated by upjet. DO NOT EDIT.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	v1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

type ResourceActionInitParameters struct {

	// The name of the resource action. It's also possible to make Http requests towards the resource ID if leave this field empty.
	Action *string `json:"action,omitempty" tf:"action,omitempty"`

	// A JSON object that contains the request body.
	Body *string `json:"body,omitempty" tf:"body,omitempty"`

	// A list of ARM resource IDs which are used to avoid modify azapi resources at the same time.
	Locks []*string `json:"locks,omitempty" tf:"locks,omitempty"`

	// Specifies the Http method of the azure resource action. Allowed values are POST, PATCH, PUT and DELETE. Defaults to POST.
	Method *string `json:"method,omitempty" tf:"method,omitempty"`

	// The ID of an existing azure source.
	ResourceID *string `json:"resourceId,omitempty" tf:"resource_id,omitempty"`

	// A list of path that needs to be exported from response body.
	// Setting it to ["*"] will export the full response body.
	// Here's an example. If it sets to ["keys"], it will set the following json to computed property output.
	ResponseExportValues []*string `json:"responseExportValues,omitempty" tf:"response_export_values,omitempty"`

	// It is in a format like <resource-type>@<api-version>. <resource-type> is the Azure resource type, for example, Microsoft.Storage/storageAccounts.
	// <api-version> is version of the API used to manage this azure resource.
	Type *string `json:"type,omitempty" tf:"type,omitempty"`

	// When to perform the action, value must be one of: apply, destroy. Default is apply.
	// When to perform the action, value must be one of: 'apply', 'destroy'. Default is 'apply'.
	When *string `json:"when,omitempty" tf:"when,omitempty"`
}

type ResourceActionObservation struct {

	// The name of the resource action. It's also possible to make Http requests towards the resource ID if leave this field empty.
	Action *string `json:"action,omitempty" tf:"action,omitempty"`

	// A JSON object that contains the request body.
	Body *string `json:"body,omitempty" tf:"body,omitempty"`

	// The ID of the azure resource action.
	ID *string `json:"id,omitempty" tf:"id,omitempty"`

	// A list of ARM resource IDs which are used to avoid modify azapi resources at the same time.
	Locks []*string `json:"locks,omitempty" tf:"locks,omitempty"`

	// Specifies the Http method of the azure resource action. Allowed values are POST, PATCH, PUT and DELETE. Defaults to POST.
	Method *string `json:"method,omitempty" tf:"method,omitempty"`

	// The output json containing the properties specified in response_export_values. Here are some examples to decode json and extract the value.
	Output *string `json:"output,omitempty" tf:"output,omitempty"`

	// The ID of an existing azure source.
	ResourceID *string `json:"resourceId,omitempty" tf:"resource_id,omitempty"`

	// A list of path that needs to be exported from response body.
	// Setting it to ["*"] will export the full response body.
	// Here's an example. If it sets to ["keys"], it will set the following json to computed property output.
	ResponseExportValues []*string `json:"responseExportValues,omitempty" tf:"response_export_values,omitempty"`

	// It is in a format like <resource-type>@<api-version>. <resource-type> is the Azure resource type, for example, Microsoft.Storage/storageAccounts.
	// <api-version> is version of the API used to manage this azure resource.
	Type *string `json:"type,omitempty" tf:"type,omitempty"`

	// When to perform the action, value must be one of: apply, destroy. Default is apply.
	// When to perform the action, value must be one of: 'apply', 'destroy'. Default is 'apply'.
	When *string `json:"when,omitempty" tf:"when,omitempty"`
}

type ResourceActionParameters struct {

	// The name of the resource action. It's also possible to make Http requests towards the resource ID if leave this field empty.
	// +kubebuilder:validation:Optional
	Action *string `json:"action,omitempty" tf:"action,omitempty"`

	// A JSON object that contains the request body.
	// +kubebuilder:validation:Optional
	Body *string `json:"body,omitempty" tf:"body,omitempty"`

	// A list of ARM resource IDs which are used to avoid modify azapi resources at the same time.
	// +kubebuilder:validation:Optional
	Locks []*string `json:"locks,omitempty" tf:"locks,omitempty"`

	// Specifies the Http method of the azure resource action. Allowed values are POST, PATCH, PUT and DELETE. Defaults to POST.
	// +kubebuilder:validation:Optional
	Method *string `json:"method,omitempty" tf:"method,omitempty"`

	// The ID of an existing azure source.
	// +kubebuilder:validation:Optional
	ResourceID *string `json:"resourceId,omitempty" tf:"resource_id,omitempty"`

	// A list of path that needs to be exported from response body.
	// Setting it to ["*"] will export the full response body.
	// Here's an example. If it sets to ["keys"], it will set the following json to computed property output.
	// +kubebuilder:validation:Optional
	ResponseExportValues []*string `json:"responseExportValues,omitempty" tf:"response_export_values,omitempty"`

	// It is in a format like <resource-type>@<api-version>. <resource-type> is the Azure resource type, for example, Microsoft.Storage/storageAccounts.
	// <api-version> is version of the API used to manage this azure resource.
	// +kubebuilder:validation:Optional
	Type *string `json:"type,omitempty" tf:"type,omitempty"`

	// When to perform the action, value must be one of: apply, destroy. Default is apply.
	// When to perform the action, value must be one of: 'apply', 'destroy'. Default is 'apply'.
	// +kubebuilder:validation:Optional
	When *string `json:"when,omitempty" tf:"when,omitempty"`
}

// ResourceActionSpec defines the desired state of ResourceAction
type ResourceActionSpec struct {
	v1.ResourceSpec `json:",inline"`
	ForProvider     ResourceActionParameters `json:"forProvider"`
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
	InitProvider ResourceActionInitParameters `json:"initProvider,omitempty"`
}

// ResourceActionStatus defines the observed state of ResourceAction.
type ResourceActionStatus struct {
	v1.ResourceStatus `json:",inline"`
	AtProvider        ResourceActionObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// ResourceAction is the Schema for the ResourceActions API. Perform resource action which changes an existing resource's state
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,upjet-azapi}
type ResourceAction struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// +kubebuilder:validation:XValidation:rule="!('*' in self.managementPolicies || 'Create' in self.managementPolicies || 'Update' in self.managementPolicies) || has(self.forProvider.resourceId) || (has(self.initProvider) && has(self.initProvider.resourceId))",message="spec.forProvider.resourceId is a required parameter"
	// +kubebuilder:validation:XValidation:rule="!('*' in self.managementPolicies || 'Create' in self.managementPolicies || 'Update' in self.managementPolicies) || has(self.forProvider.type) || (has(self.initProvider) && has(self.initProvider.type))",message="spec.forProvider.type is a required parameter"
	Spec   ResourceActionSpec   `json:"spec"`
	Status ResourceActionStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ResourceActionList contains a list of ResourceActions
type ResourceActionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ResourceAction `json:"items"`
}

// Repository type metadata.
var (
	ResourceAction_Kind             = "ResourceAction"
	ResourceAction_GroupKind        = schema.GroupKind{Group: CRDGroup, Kind: ResourceAction_Kind}.String()
	ResourceAction_KindAPIVersion   = ResourceAction_Kind + "." + CRDGroupVersion.String()
	ResourceAction_GroupVersionKind = CRDGroupVersion.WithKind(ResourceAction_Kind)
)

func init() {
	SchemeBuilder.Register(&ResourceAction{}, &ResourceActionList{})
}
