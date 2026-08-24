// SPDX-FileCopyrightText: 2025 Upbound Inc. <https://upbound.io>
//
// SPDX-License-Identifier: Apache-2.0

package clients

import (
	"context"
	"encoding/json"
	"os"

	"github.com/Azure/terraform-provider-azapi/xpprovider"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	xpv2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	"github.com/crossplane/upjet/v2/pkg/terraform"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clusterv1beta1 "github.com/upbound/provider-azapi/v2/apis/cluster/v1beta1"
	namespacedv1beta1 "github.com/upbound/provider-azapi/v2/apis/namespaced/v1beta1"
)

const (
	// error messages
	errNoProviderConfig        = "no providerConfigRef provided"
	errGetProviderConfig       = "cannot get referenced ProviderConfig"
	errTrackUsage              = "cannot track ProviderConfig usage"
	errExtractCredentials      = "cannot extract credentials"
	errUnmarshalCredentials    = "cannot unmarshal azapi credentials as JSON"
	keySubscriptionID          = "subscriptionId"
	keyClientID                = "clientId"
	keyClientSecret            = "clientSecret"
	keyTenantID                = "tenantId"
	keyTerraformSubscriptionID = "subscription_id"
	keyTerraformClientID       = "client_id"
	keyTerraformClientSecret   = "client_secret"
	keyTerraformTenantID       = "tenant_id"

	// error messages and config keys specific to OIDC / Workload Identity
	// Federation auth
	errTenantIDNotSet    = "tenant ID must be set in ProviderConfig when credentials source is OIDCTokenFile"
	errClientIDNotSet    = "client ID must be set in ProviderConfig when credentials source is OIDCTokenFile"
	keyUseOIDC           = "use_oidc"
	keyEnvironment       = "environment"
	keyOidcTokenFilePath = "oidc_token_file_path"
	// defaultOidcTokenFilePath is the path the Workload Identity mutating
	// webhook projects the federated token to by default.
	defaultOidcTokenFilePath = "/var/run/secrets/azure/tokens/azure-identity-token"
	// envAzureFederatedTokenFile is the environment variable the Azure
	// Workload Identity mutating webhook sets on injected pods to the path it
	// actually projected the federated token to. That path has changed
	// between webhook releases, so it is a more reliable source than
	// defaultOidcTokenFilePath.
	envAzureFederatedTokenFile = "AZURE_FEDERATED_TOKEN_FILE"
)

const (
	credentialsSourceOIDCTokenFile xpv2.CredentialsSource = "OIDCTokenFile"
)

// TerraformSetupBuilder builds Terraform a terraform.SetupFn function which
// returns Terraform provider setup configuration
func TerraformSetupBuilder() terraform.SetupFn {
	return func(ctx context.Context, client client.Client, mgx resource.Managed) (terraform.Setup, error) {
		ps := terraform.Setup{Configuration: map[string]any{}}

		pcSpec, err := resolveProviderConfig(ctx, client, mgx)
		if err != nil {
			return terraform.Setup{}, err
		}

		switch pcSpec.Credentials.Source { //nolint:exhaustive
		case credentialsSourceOIDCTokenFile:
			err = oidcAuth(pcSpec, &ps)
		default:
			err = spAuth(ctx, pcSpec, &ps, client)
		}
		if err != nil {
			return terraform.Setup{}, err
		}

		applyCommonOverrides(pcSpec, &ps)

		ps.FrameworkProvider, err = xpprovider.FrameworkProvider(ctx)
		if err != nil {
			return terraform.Setup{}, errors.Wrap(err, "error initializing the framework provider")
		}
		return ps, nil
	}
}

// applyCommonOverrides applies ProviderConfigSpec fields that are common to
// every credentials source. It is called once, after the credentials-source
// switch in TerraformSetupBuilder, so that Environment and SubscriptionID are
// honored regardless of which auth path populated ps.Configuration (e.g. so a
// Secret-sourced setup can still opt into "usgovernment", and so
// SubscriptionID can override the value extracted from a Secret if the user
// explicitly sets it).
func applyCommonOverrides(pcSpec *namespacedv1beta1.ProviderConfigSpec, ps *terraform.Setup) {
	if pcSpec.Environment != nil && *pcSpec.Environment != "" {
		ps.Configuration[keyEnvironment] = *pcSpec.Environment
	}
	if pcSpec.SubscriptionID != nil && *pcSpec.SubscriptionID != "" {
		ps.Configuration[keyTerraformSubscriptionID] = *pcSpec.SubscriptionID
	}
}

// spAuth populates ps.Configuration from the Secret/Environment/Filesystem/
// None credential sources, using the CommonCredentialExtractor.
func spAuth(ctx context.Context, pcSpec *namespacedv1beta1.ProviderConfigSpec, ps *terraform.Setup, crClient client.Client) error {
	data, err := resource.CommonCredentialExtractor(ctx, pcSpec.Credentials.Source, crClient, pcSpec.Credentials.CommonCredentialSelectors)
	if err != nil {
		return errors.Wrap(err, errExtractCredentials)
	}
	creds := map[string]string{}
	if err := json.Unmarshal(data, &creds); err != nil {
		return errors.Wrap(err, errUnmarshalCredentials)
	}

	if v, ok := creds[keySubscriptionID]; ok {
		ps.Configuration[keyTerraformSubscriptionID] = v
	}
	if v, ok := creds[keyClientID]; ok {
		ps.Configuration[keyTerraformClientID] = v
	}
	if v, ok := creds[keyClientSecret]; ok {
		ps.Configuration[keyTerraformClientSecret] = v
	}
	if v, ok := creds[keyTenantID]; ok {
		ps.Configuration[keyTerraformTenantID] = v
	}
	return nil
}

// oidcAuth populates ps.Configuration for the OIDCTokenFile credential
// source (Workload Identity Federation via a projected service-account
// token file).
func oidcAuth(pcSpec *namespacedv1beta1.ProviderConfigSpec, ps *terraform.Setup) error {
	if pcSpec.TenantID == nil || len(*pcSpec.TenantID) == 0 {
		return errors.New(errTenantIDNotSet)
	}
	if pcSpec.ClientID == nil || len(*pcSpec.ClientID) == 0 {
		return errors.New(errClientIDNotSet)
	}
	ps.Configuration[keyUseOIDC] = true
	// An explicit spec.oidcTokenFilePath always wins. Otherwise prefer
	// AZURE_FEDERATED_TOKEN_FILE, which the Azure Workload Identity webhook
	// sets to wherever it actually projected the token, falling back to the
	// historical hardcoded default only when that variable is unset.
	tokenPath := defaultOidcTokenFilePath
	if v := os.Getenv(envAzureFederatedTokenFile); len(v) > 0 {
		tokenPath = v
	}
	if pcSpec.OidcTokenFilePath != nil && len(*pcSpec.OidcTokenFilePath) > 0 {
		tokenPath = *pcSpec.OidcTokenFilePath
	}
	ps.Configuration[keyOidcTokenFilePath] = tokenPath
	ps.Configuration[keyTerraformTenantID] = *pcSpec.TenantID
	ps.Configuration[keyTerraformClientID] = *pcSpec.ClientID
	return nil
}

func legacyToModernProviderConfigSpec(pc *clusterv1beta1.ProviderConfig) (*namespacedv1beta1.ProviderConfigSpec, error) {
	if pc == nil {
		return nil, nil
	}
	data, err := json.Marshal(pc.Spec)
	if err != nil {
		return nil, err
	}

	var mSpec namespacedv1beta1.ProviderConfigSpec
	err = json.Unmarshal(data, &mSpec)
	return &mSpec, err
}

func enrichLocalSecretRefs(pc *namespacedv1beta1.ProviderConfig, mg resource.Managed) {
	if pc != nil && pc.Spec.Credentials.SecretRef != nil {
		pc.Spec.Credentials.SecretRef.Namespace = mg.GetNamespace()
	}
}

func resolveProviderConfig(ctx context.Context, crClient client.Client, mg resource.Managed) (*namespacedv1beta1.ProviderConfigSpec, error) {
	switch managed := mg.(type) {
	case resource.LegacyManaged: //nolint:staticcheck // still handling the cluster-scoped MRs
		return resolveProviderConfigLegacy(ctx, crClient, managed)
	case resource.ModernManaged:
		return resolveProviderConfigModern(ctx, crClient, managed)
	default:
		return nil, errors.New("resource is not a managed")
	}
}

func resolveProviderConfigLegacy(ctx context.Context, client client.Client, mg resource.LegacyManaged) (*namespacedv1beta1.ProviderConfigSpec, error) { //nolint:staticcheck // still handling the cluster-scoped MRs
	configRef := mg.GetProviderConfigReference()
	if configRef == nil {
		return nil, errors.New(errNoProviderConfig)
	}
	pc := &clusterv1beta1.ProviderConfig{}
	if err := client.Get(ctx, types.NamespacedName{Name: configRef.Name}, pc); err != nil {
		return nil, errors.Wrap(err, errGetProviderConfig)
	}

	t := resource.NewLegacyProviderConfigUsageTracker(client, &clusterv1beta1.ProviderConfigUsage{})
	if err := t.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, errTrackUsage)
	}

	return legacyToModernProviderConfigSpec(pc)
}

func resolveProviderConfigModern(ctx context.Context, crClient client.Client, mg resource.ModernManaged) (*namespacedv1beta1.ProviderConfigSpec, error) {
	configRef := mg.GetProviderConfigReference()
	if configRef == nil {
		return nil, errors.New(errNoProviderConfig)
	}

	pcRuntimeObj, err := crClient.Scheme().New(namespacedv1beta1.SchemeGroupVersion.WithKind(configRef.Kind))
	if err != nil {
		return nil, errors.Wrapf(err, "referenced provider config kind %q is invalid for %s/%s", configRef.Kind, mg.GetNamespace(), mg.GetName())
	}
	pcObj, ok := pcRuntimeObj.(resource.ProviderConfig)
	if !ok {
		return nil, errors.Errorf("referenced provider config kind %q is not a provider config type %s/%s", configRef.Kind, mg.GetNamespace(), mg.GetName())
	}

	// Namespace will be ignored if the PC is a cluster-scoped type
	if err := crClient.Get(ctx, types.NamespacedName{Name: configRef.Name, Namespace: mg.GetNamespace()}, pcObj); err != nil {
		return nil, errors.Wrap(err, errGetProviderConfig)
	}

	var pcSpec namespacedv1beta1.ProviderConfigSpec
	switch pc := pcObj.(type) {
	case *namespacedv1beta1.ProviderConfig:
		enrichLocalSecretRefs(pc, mg)
		pcSpec = pc.Spec
	case *namespacedv1beta1.ClusterProviderConfig:
		pcSpec = pc.Spec
	default:
		return nil, errors.New("unknown provider config kind")
	}
	t := resource.NewProviderConfigUsageTracker(crClient, &namespacedv1beta1.ProviderConfigUsage{})
	if err := t.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, errTrackUsage)
	}
	return &pcSpec, nil
}
