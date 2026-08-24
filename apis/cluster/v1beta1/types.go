// SPDX-FileCopyrightText: 2025 Upbound Inc. <https://upbound.io>
//
// SPDX-License-Identifier: Apache-2.0

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpv2 "github.com/crossplane/crossplane/apis/v2/core/v2"
)

// A ProviderConfigSpec defines the desired state of a ProviderConfig.
type ProviderConfigSpec struct {
	// Credentials required to authenticate to this provider.
	Credentials ProviderCredentials `json:"credentials"`

	// ClientID is the client ID of the Azure AD application to use.
	// Required if Credentials.Source is OIDCTokenFile.
	// +kubebuilder:validation:Optional
	ClientID *string `json:"clientID,omitempty"`

	// TenantID is the Azure AD tenant ID to use.
	// Required if Credentials.Source is OIDCTokenFile.
	// +kubebuilder:validation:Optional
	TenantID *string `json:"tenantID,omitempty"`

	// Environment is the Azure Cloud Environment to use. One of "public",
	// "usgovernment", "china", "custom". Defaults to "public".
	// Note: "custom" additionally requires ARM_ACTIVE_DIRECTORY_AUTHORITY_HOST,
	// ARM_RESOURCE_MANAGER_ENDPOINT, and ARM_RESOURCE_MANAGER_AUDIENCE
	// environment variables on the provider deployment, which this API does
	// not expose directly.
	// +kubebuilder:validation:Optional
	Environment *string `json:"environment,omitempty"`

	// OidcTokenFilePath is the path to the file containing the OIDC/federated
	// identity token, projected into the pod by the Workload Identity
	// mutating webhook. Only used if Credentials.Source is OIDCTokenFile.
	// If unset, the AZURE_FEDERATED_TOKEN_FILE environment variable is used,
	// which the Azure Workload Identity webhook sets to the path it projected
	// the token to. If that is also unset, it defaults to
	// "/var/run/secrets/azure/tokens/azure-identity-token".
	// +kubebuilder:validation:Optional
	OidcTokenFilePath *string `json:"oidcTokenFilePath,omitempty"`

	// SubscriptionID is the Azure subscription ID to use. If unset, and
	// Credentials.Source is Secret/Environment/Filesystem, the subscription ID
	// from the extracted credentials is used instead. Required for
	// OIDCTokenFile if the underlying resources need it and no default
	// subscription can be resolved another way.
	// +kubebuilder:validation:Optional
	SubscriptionID *string `json:"subscriptionID,omitempty"`
}

// ProviderCredentials required to authenticate.
type ProviderCredentials struct {
	// Source of the provider credentials.
	// +kubebuilder:validation:Enum=None;Secret;InjectedIdentity;Environment;Filesystem;OIDCTokenFile
	Source xpv2.CredentialsSource `json:"source"`

	xpv2.CommonCredentialSelectors `json:",inline"`
}

// A ProviderConfigStatus reflects the observed state of a ProviderConfig.
type ProviderConfigStatus struct {
	xpv2.ProviderConfigStatus `json:",inline"`
}

// +kubebuilder:object:root=true

// A ProviderConfig configures a AzAPI provider.
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="SECRET-NAME",type="string",JSONPath=".spec.credentials.secretRef.name",priority=1
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:resource:scope=Cluster,categories={crossplane,provider,azapi}
type ProviderConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProviderConfigSpec   `json:"spec"`
	Status ProviderConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ProviderConfigList contains a list of ProviderConfig.
type ProviderConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ProviderConfig `json:"items"`
}

// +kubebuilder:object:root=true

// A ProviderConfigUsage indicates that a resource is using a ProviderConfig.
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="CONFIG-NAME",type="string",JSONPath=".providerConfigRef.name"
// +kubebuilder:printcolumn:name="RESOURCE-KIND",type="string",JSONPath=".resourceRef.kind"
// +kubebuilder:printcolumn:name="RESOURCE-NAME",type="string",JSONPath=".resourceRef.name"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,provider,azapi}
type ProviderConfigUsage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	xpv2.ProviderConfigUsage `json:",inline"`
}

// +kubebuilder:object:root=true

// ProviderConfigUsageList contains a list of ProviderConfigUsage
type ProviderConfigUsageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ProviderConfigUsage `json:"items"`
}
