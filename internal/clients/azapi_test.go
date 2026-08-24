// SPDX-FileCopyrightText: 2026 Upbound Inc. <https://upbound.io>
//
// SPDX-License-Identifier: Apache-2.0

package clients

import (
	"context"
	"strings"
	"testing"

	xpv2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	"github.com/crossplane/upjet/v2/pkg/terraform"
	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	namespacedv1beta1 "github.com/upbound/provider-azapi/v2/apis/namespaced/v1beta1"
)

func strPtr(s string) *string { return &s }

func TestOidcAuth(t *testing.T) {
	cases := map[string]struct {
		spec *namespacedv1beta1.ProviderConfigSpec
		// env is the value of AZURE_FEDERATED_TOKEN_FILE for the case. It is
		// always applied via t.Setenv so the cases stay hermetic regardless of
		// the environment the tests run in.
		env     string
		want    map[string]any
		wantErr string
	}{
		"DefaultTokenPath": {
			spec: &namespacedv1beta1.ProviderConfigSpec{
				TenantID: strPtr("tenant-1"),
				ClientID: strPtr("client-1"),
			},
			want: map[string]any{
				keyUseOIDC:           true,
				keyOidcTokenFilePath: defaultOidcTokenFilePath,
				keyTerraformTenantID: "tenant-1",
				keyTerraformClientID: "client-1",
			},
		},
		"CustomTokenPath": {
			spec: &namespacedv1beta1.ProviderConfigSpec{
				TenantID:          strPtr("tenant-1"),
				ClientID:          strPtr("client-1"),
				OidcTokenFilePath: strPtr("/custom/path/token"),
			},
			want: map[string]any{
				keyUseOIDC:           true,
				keyOidcTokenFilePath: "/custom/path/token",
				keyTerraformTenantID: "tenant-1",
				keyTerraformClientID: "client-1",
			},
		},
		"EmptyTokenPathFallsBackToDefault": {
			// An empty-string OidcTokenFilePath must be treated the same as
			// unset, falling back to defaultOidcTokenFilePath, rather than
			// writing oidc_token_file_path: "".
			spec: &namespacedv1beta1.ProviderConfigSpec{
				TenantID:          strPtr("tenant-1"),
				ClientID:          strPtr("client-1"),
				OidcTokenFilePath: strPtr(""),
			},
			want: map[string]any{
				keyUseOIDC:           true,
				keyOidcTokenFilePath: defaultOidcTokenFilePath,
				keyTerraformTenantID: "tenant-1",
				keyTerraformClientID: "client-1",
			},
		},
		"FederatedTokenFileEnvPreferredOverDefault": {
			// The Workload Identity webhook sets
			// AZURE_FEDERATED_TOKEN_FILE to the path it actually projected
			// the token to, which has moved between webhook releases, so it
			// must win over the hardcoded default.
			spec: &namespacedv1beta1.ProviderConfigSpec{
				TenantID: strPtr("tenant-1"),
				ClientID: strPtr("client-1"),
			},
			env: "/var/run/secrets/azure/tokens/injected-token",
			want: map[string]any{
				keyUseOIDC:           true,
				keyOidcTokenFilePath: "/var/run/secrets/azure/tokens/injected-token",
				keyTerraformTenantID: "tenant-1",
				keyTerraformClientID: "client-1",
			},
		},
		"ExplicitTokenPathWinsOverFederatedTokenFileEnv": {
			spec: &namespacedv1beta1.ProviderConfigSpec{
				TenantID:          strPtr("tenant-1"),
				ClientID:          strPtr("client-1"),
				OidcTokenFilePath: strPtr("/custom/path/token"),
			},
			env: "/var/run/secrets/azure/tokens/injected-token",
			want: map[string]any{
				keyUseOIDC:           true,
				keyOidcTokenFilePath: "/custom/path/token",
				keyTerraformTenantID: "tenant-1",
				keyTerraformClientID: "client-1",
			},
		},
		"EmptyTokenPathFallsBackToFederatedTokenFileEnv": {
			// An empty-string OidcTokenFilePath is treated as unset, so the
			// env var still takes precedence over the hardcoded default.
			spec: &namespacedv1beta1.ProviderConfigSpec{
				TenantID:          strPtr("tenant-1"),
				ClientID:          strPtr("client-1"),
				OidcTokenFilePath: strPtr(""),
			},
			env: "/var/run/secrets/azure/tokens/injected-token",
			want: map[string]any{
				keyUseOIDC:           true,
				keyOidcTokenFilePath: "/var/run/secrets/azure/tokens/injected-token",
				keyTerraformTenantID: "tenant-1",
				keyTerraformClientID: "client-1",
			},
		},
		"SubscriptionIDNotAppliedByOidcAuthDirectly": {
			// SubscriptionID is applied centrally by applyCommonOverrides in
			// TerraformSetupBuilder, not by oidcAuth itself.
			spec: &namespacedv1beta1.ProviderConfigSpec{
				TenantID:       strPtr("tenant-1"),
				ClientID:       strPtr("client-1"),
				SubscriptionID: strPtr("sub-1"),
			},
			want: map[string]any{
				keyUseOIDC:           true,
				keyOidcTokenFilePath: defaultOidcTokenFilePath,
				keyTerraformTenantID: "tenant-1",
				keyTerraformClientID: "client-1",
			},
		},
		"MissingTenantID": {
			spec:    &namespacedv1beta1.ProviderConfigSpec{ClientID: strPtr("client-1")},
			wantErr: errTenantIDNotSet,
		},
		"MissingClientID": {
			spec:    &namespacedv1beta1.ProviderConfigSpec{TenantID: strPtr("tenant-1")},
			wantErr: errClientIDNotSet,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// Always set the variable, empty included, so a value leaking in
			// from the ambient environment cannot change the outcome.
			t.Setenv(envAzureFederatedTokenFile, tc.env)
			ps := &terraform.Setup{Configuration: map[string]any{}}
			err := oidcAuth(tc.spec, ps)
			if tc.wantErr != "" {
				if err == nil || err.Error() != tc.wantErr {
					t.Fatalf("oidcAuth() error = %v, want %q", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("oidcAuth() unexpected error = %v", err)
			}
			if diff := cmp.Diff(tc.want, map[string]any(ps.Configuration)); diff != "" {
				t.Errorf("oidcAuth() Configuration mismatch (-want +got):\n%s", diff)
			}
			if useOIDC, ok := ps.Configuration[keyUseOIDC]; ok {
				if _, isBool := useOIDC.(bool); !isBool {
					t.Errorf("Configuration[%q] = %#v (%T), want a bool", keyUseOIDC, useOIDC, useOIDC)
				}
			}
		})
	}
}

// TestApplyCommonOverrides covers the hoisted Environment/SubscriptionID
// application logic used by TerraformSetupBuilder after the credentials-
// source switch. This is what makes Environment/SubscriptionID apply
// uniformly across every credentials source, including the default
// (spAuth/Secret) path — see items 2 and 3 of the final review fix brief.
func TestApplyCommonOverrides(t *testing.T) {
	cases := map[string]struct {
		spec    *namespacedv1beta1.ProviderConfigSpec
		initial map[string]any
		want    map[string]any
	}{
		"EnvironmentAppliedForDefaultSpAuthPath": {
			// Simulates the state ps.Configuration would be in after spAuth
			// has run for a Secret-sourced ProviderConfig: no environment key
			// present yet because spAuth never sets it.
			spec: &namespacedv1beta1.ProviderConfigSpec{
				Environment: strPtr("usgovernment"),
			},
			initial: map[string]any{
				keyTerraformSubscriptionID: "sub-from-secret",
				keyTerraformClientID:       "client-from-secret",
				keyTerraformClientSecret:   "secret-from-secret",
				keyTerraformTenantID:       "tenant-from-secret",
			},
			want: map[string]any{
				keyTerraformSubscriptionID: "sub-from-secret",
				keyTerraformClientID:       "client-from-secret",
				keyTerraformClientSecret:   "secret-from-secret",
				keyTerraformTenantID:       "tenant-from-secret",
				keyEnvironment:             "usgovernment",
			},
		},
		"NilEnvironmentLeavesConfigurationUnchanged": {
			spec:    &namespacedv1beta1.ProviderConfigSpec{},
			initial: map[string]any{},
			want:    map[string]any{},
		},
		"EmptyEnvironmentLeavesConfigurationUnchanged": {
			spec:    &namespacedv1beta1.ProviderConfigSpec{Environment: strPtr("")},
			initial: map[string]any{},
			want:    map[string]any{},
		},
		"SubscriptionIDAppliedWhenUnset": {
			spec: &namespacedv1beta1.ProviderConfigSpec{
				SubscriptionID: strPtr("sub-1"),
			},
			initial: map[string]any{},
			want: map[string]any{
				keyTerraformSubscriptionID: "sub-1",
			},
		},
		"SubscriptionIDOverridesSecretDerivedValue": {
			// Explicitly setting SubscriptionID must override the value
			// extracted from a Secret by spAuth.
			spec: &namespacedv1beta1.ProviderConfigSpec{
				SubscriptionID: strPtr("sub-override"),
			},
			initial: map[string]any{
				keyTerraformSubscriptionID: "sub-from-secret",
			},
			want: map[string]any{
				keyTerraformSubscriptionID: "sub-override",
			},
		},
		"EmptySubscriptionIDLeavesConfigurationUnchanged": {
			spec: &namespacedv1beta1.ProviderConfigSpec{
				SubscriptionID: strPtr(""),
			},
			initial: map[string]any{
				keyTerraformSubscriptionID: "sub-from-secret",
			},
			want: map[string]any{
				keyTerraformSubscriptionID: "sub-from-secret",
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			ps := &terraform.Setup{Configuration: tc.initial}
			applyCommonOverrides(tc.spec, ps)
			if diff := cmp.Diff(tc.want, map[string]any(ps.Configuration)); diff != "" {
				t.Errorf("applyCommonOverrides() Configuration mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// TestSpAuth is the regression test for the Secret/Environment/Filesystem/
// None credential sources (the design doc called for one explicitly). spAuth
// was refactored (moved, re-parameterized with an explicit crClient
// argument) in this branch, so it warrants direct coverage.
func TestSpAuth(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add corev1 to scheme: %v", err)
	}

	const (
		secretNamespace = "crossplane-system"
		secretName      = "azapi-creds"
		secretKey       = "credentials"
	)

	validSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: secretName, Namespace: secretNamespace},
		Data: map[string][]byte{
			secretKey: []byte(`{"subscriptionId":"sub-1","clientId":"client-1","clientSecret":"secret-1","tenantId":"tenant-1"}`),
		},
	}
	nonJSONSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: secretName, Namespace: secretNamespace},
		Data: map[string][]byte{
			secretKey: []byte(`not-json`),
		},
	}

	specWithSecretRef := func() *namespacedv1beta1.ProviderConfigSpec {
		return &namespacedv1beta1.ProviderConfigSpec{
			Credentials: namespacedv1beta1.ProviderCredentials{
				Source: xpv2.CredentialsSourceSecret,
				CommonCredentialSelectors: xpv2.CommonCredentialSelectors{
					SecretRef: &xpv2.SecretKeySelector{
						SecretReference: xpv2.SecretReference{
							Name:      secretName,
							Namespace: secretNamespace,
						},
						Key: secretKey,
					},
				},
			},
		}
	}

	cases := map[string]struct {
		objects []runtime.Object
		spec    *namespacedv1beta1.ProviderConfigSpec
		want    map[string]any
		wantErr string
	}{
		"SecretSourcePopulatesConfiguration": {
			objects: []runtime.Object{validSecret},
			spec:    specWithSecretRef(),
			want: map[string]any{
				keyTerraformSubscriptionID: "sub-1",
				keyTerraformClientID:       "client-1",
				keyTerraformClientSecret:   "secret-1",
				keyTerraformTenantID:       "tenant-1",
			},
		},
		"SecretNotFound": {
			objects: nil,
			spec:    specWithSecretRef(),
			wantErr: errExtractCredentials,
		},
		"NonJSONSecretData": {
			objects: []runtime.Object{nonJSONSecret},
			spec:    specWithSecretRef(),
			wantErr: errUnmarshalCredentials,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			crClient := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(tc.objects...).Build()
			ps := &terraform.Setup{Configuration: map[string]any{}}
			err := spAuth(context.Background(), tc.spec, ps, crClient)
			if tc.wantErr != "" {
				if err == nil {
					t.Fatalf("spAuth() error = nil, want error containing %q", tc.wantErr)
				}
				if !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("spAuth() error = %q, want it to contain %q", err.Error(), tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("spAuth() unexpected error = %v", err)
			}
			if diff := cmp.Diff(tc.want, map[string]any(ps.Configuration)); diff != "" {
				t.Errorf("spAuth() Configuration mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
