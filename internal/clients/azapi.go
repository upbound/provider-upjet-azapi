// SPDX-FileCopyrightText: 2025 Upbound Inc. <https://upbound.io>
//
// SPDX-License-Identifier: Apache-2.0

package clients

import (
	"context"
	"encoding/json"

	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/upjet/pkg/terraform"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tfsdk "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/upbound/provider-azapi/apis/v1beta1"
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
)

// TerraformSetupBuilder builds Terraform a terraform.SetupFn function which
// returns Terraform provider setup configuration
func TerraformSetupBuilder(tfProvider *schema.Provider) terraform.SetupFn {
	return func(ctx context.Context, client client.Client, mg resource.Managed) (terraform.Setup, error) {
		ps := terraform.Setup{}

		configRef := mg.GetProviderConfigReference()
		if configRef == nil {
			return ps, errors.New(errNoProviderConfig)
		}
		pc := &v1beta1.ProviderConfig{}
		if err := client.Get(ctx, types.NamespacedName{Name: configRef.Name}, pc); err != nil {
			return ps, errors.Wrap(err, errGetProviderConfig)
		}

		t := resource.NewProviderConfigUsageTracker(client, &v1beta1.ProviderConfigUsage{})
		if err := t.Track(ctx, mg); err != nil {
			return ps, errors.Wrap(err, errTrackUsage)
		}

		data, err := resource.CommonCredentialExtractor(ctx, pc.Spec.Credentials.Source, client, pc.Spec.Credentials.CommonCredentialSelectors)
		if err != nil {
			return ps, errors.Wrap(err, errExtractCredentials)
		}
		creds := map[string]string{}
		if err := json.Unmarshal(data, &creds); err != nil {
			return ps, errors.Wrap(err, errUnmarshalCredentials)
		}

		// set provider configuration
		ps.Configuration = map[string]any{}
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
		return ps, errors.Wrap(configureTFProviderMeta(ctx, &ps, *tfProvider), "failed to configure the Terraform AzureAD provider meta")
	}
}

func configureTFProviderMeta(ctx context.Context, ps *terraform.Setup, p schema.Provider) error {
	// Please be aware that this implementation relies on the schema.Provider
	// parameter `p` being a non-pointer. This is because normally
	// the Terraform plugin SDK normally configures the provider
	// only once and using a pointer argument here will cause
	// race conditions between resources referring to different
	// ProviderConfigs.
	diag := p.Configure(context.WithoutCancel(ctx), &tfsdk.ResourceConfig{
		Config: ps.Configuration,
	})
	if diag != nil && diag.HasError() {
		return errors.Errorf("failed to configure the provider: %v", diag)
	}
	ps.Meta = p.Meta()
	return nil
}
