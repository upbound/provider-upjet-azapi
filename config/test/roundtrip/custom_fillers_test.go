// SPDX-FileCopyrightText: 2026 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package roundtrip

import (
	"encoding/json"

	"github.com/crossplane/upjet/v2/pkg/apitesting/roundtrip"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/randfill"

	resourcesv1beta1 "github.com/upbound/provider-azapi/v2/apis/cluster/resources/v1beta1"
	resourcesv1beta2 "github.com/upbound/provider-azapi/v2/apis/cluster/resources/v1beta2"
)

// validJSONObjects are pre-built valid JSON object payloads resembling AzAPI request bodies.
// Keys at every nesting level must be in alphabetical order so that roundtripping through
// Kubernetes unstructured conversion (which re-marshals via map[string]interface{} and
// sorts keys) produces identical bytes on the way back.
// Additional constraints:
//   - Avoid numbers that lose precision as float64 (json.Unmarshal decodes all numbers to float64
//     for interface{} targets, so very large integers change value).
//   - Avoid trailing-zero floats like 1.0 — json.Marshal(float64(1)) emits "1", not "1.0".
//   - Unicode escape sequences like A decode to the character "A" and re-encode as "A",
//     so use literal characters rather than escape sequences.
var validJSONObjects = []json.RawMessage{
	// minimal / zero-value body
	json.RawMessage(`{}`),
	// single-key flat object
	json.RawMessage(`{"location":"eastus"}`),
	// nested object (sku is common in ARM)
	json.RawMessage(`{"sku":{"capacity":1,"name":"Standard"}}`),
	// flat object with two scalar types
	json.RawMessage(`{"properties":{"enabled":true,"tier":"Standard"}}`),
	// realistic storage-account body: multiple top-level keys, nested properties
	// (alphabetical at every level: kind < location < properties < sku;
	//  inside properties: accessTier < allowBlobPublicAccess < minimumTlsVersion < supportsHttpsTrafficOnly)
	json.RawMessage(`{"kind":"StorageV2","location":"eastus","properties":{"accessTier":"Hot","allowBlobPublicAccess":false,"minimumTlsVersion":"TLS1_2","supportsHttpsTrafficOnly":true},"sku":{"name":"Standard_LRS"}}`),
	// body with an array-valued property and string tags map
	// (alphabetical: location < locks < tags; inside tags: environment < owner)
	json.RawMessage(`{"location":"westus","locks":["/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet"],"tags":{"environment":"prod","owner":"platform"}}`),
	// body with a null-valued property (null passes through unstructured as nil → "null")
	// (alphabetical: location < properties; inside properties: extendedLocation < hardwareProfile < osProfile;
	//  inside hardwareProfile: vmSize; inside osProfile: adminUsername < computerName)
	json.RawMessage(`{"location":"eastus","properties":{"extendedLocation":null,"hardwareProfile":{"vmSize":"Standard_D2s_v3"},"osProfile":{"adminUsername":"azureuser","computerName":"vm01"}}}`),
}

// validJSONStringArrays are pre-built JSON arrays of strings.
// responseExportValues in v1beta2 is *v1.JSON but roundtrips through []*string in v1beta1,
// so the JSON must be a valid string array (not an arbitrary JSON object).
// The arrays are compared after json.Marshal([]*string{...}), which produces compact JSON with
// no extra whitespace, so entries here must not contain spaces either.
var validJSONStringArrays = []json.RawMessage{
	// empty — callers that treat nil and empty the same still pass
	json.RawMessage(`[]`),
	// wildcard — export the full response body
	json.RawMessage(`["*"]`),
	// single top-level property
	json.RawMessage(`["properties.loginServer"]`),
	// two nested paths (common for ACR output)
	json.RawMessage(`["properties.loginServer","properties.policies.quarantinePolicy.status"]`),
	// VM-style paths including bracket-indexed sub-resource
	json.RawMessage(`["properties.networkProfile.networkInterfaces[0].id","properties.storageProfile.osDisk.name"]`),
	// storage account paths — three sibling properties
	json.RawMessage(`["properties.accessTier","properties.minimumTlsVersion","properties.supportsHttpsTrafficOnly"]`),
}

// validJSONStrings are valid JSON-encoded strings usable for v1beta1 *string body/output fields.
// These go through the custom converter (string ↔ v1.JSON raw bytes), not identity conversion,
// so key ordering does not need to be alphabetical.
var validJSONStrings = []string{
	`{}`,
	`{"location":"eastus"}`,
	`{"sku":{"name":"Standard","capacity":1}}`,
	`{"properties":{"enabled":true}}`,
	`{"id":"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.Storage/storageAccounts/sa","properties":{"primaryEndpoints":{"blob":"https://sa.blob.core.windows.net/"},"provisioningState":"Succeeded"}}`,
}

func randJSONStringArray(c randfill.Continue) *apiextv1.JSON {
	return &apiextv1.JSON{Raw: validJSONStringArrays[c.Intn(len(validJSONStringArrays))]}
}

func randJSONString(c randfill.Continue) *string {
	s := validJSONStrings[c.Intn(len(validJSONStrings))]
	return &s
}

// fuzzStringJSON replaces a non-nil *string with valid JSON (preserving nil).
func fuzzStringJSON(sp **string, c randfill.Continue) {
	if *sp != nil {
		*sp = randJSONString(c)
	}
}

// fuzzJSONStringArray replaces a non-nil *apiextv1.JSON with a JSON string array (preserving nil).
// Used for responseExportValues which must survive json.Unmarshal into []*string in v1beta1.
func fuzzJSONStringArray(jp **apiextv1.JSON, c randfill.Continue) {
	if *jp != nil {
		*jp = randJSONStringArray(c)
	}
}

var azapiCustomFuzzers = []roundtrip.FuzzFunc{
	// Type-level fuzzer: whenever randfill needs to fill an *apiextv1.JSON value,
	// produce a valid JSON object instead of random bytes.
	// This covers body, output, sensitiveBody and sensitiveResponseExportValues fields
	// which are *v1.JSON in both v1beta1 and v1beta2.
	apiExtJSONFuzzer,
	fuzzDataplaneResourceV1Beta1,
	fuzzResourceV1Beta1,
	fuzzUpdateResourceV1Beta1,
	fuzzResourceActionV1Beta1,
	fuzzDataplaneResourceV1Beta2,
	fuzzResourceV1Beta2,
	fuzzUpdateResourceV1Beta2,
	fuzzResourceActionV1Beta2,
}

// apiExtJSONFuzzer fills *apiextv1.JSON with a valid JSON object.
func apiExtJSONFuzzer(j *apiextv1.JSON, c randfill.Continue) {
	j.Raw = validJSONObjects[c.Intn(len(validJSONObjects))]
}

// --- v1beta1 fuzzers ---
// In v1beta1: body and output are *string (need valid JSON to cleanly roundtrip
// through *v1.JSON in v1beta2). sensitiveBody is *apiextv1.JSON and is handled
// by apiExtJSONFuzzer. responseExportValues is []*string and roundtrips freely.

func fuzzDataplaneResourceV1Beta1(s *resourcesv1beta1.DataPlaneResource, c randfill.Continue) {
	c.FillNoCustom(s)
	fuzzStringJSON(&s.Spec.ForProvider.Body, c)
	fuzzStringJSON(&s.Spec.InitProvider.Body, c)
	fuzzStringJSON(&s.Status.AtProvider.Body, c)
	fuzzStringJSON(&s.Status.AtProvider.Output, c)
}

func fuzzResourceV1Beta1(s *resourcesv1beta1.Resource, c randfill.Continue) {
	c.FillNoCustom(s)
	fuzzStringJSON(&s.Spec.ForProvider.Body, c)
	fuzzStringJSON(&s.Spec.InitProvider.Body, c)
	fuzzStringJSON(&s.Status.AtProvider.Body, c)
	fuzzStringJSON(&s.Status.AtProvider.Output, c)
}

func fuzzUpdateResourceV1Beta1(s *resourcesv1beta1.UpdateResource, c randfill.Continue) {
	c.FillNoCustom(s)
	fuzzStringJSON(&s.Spec.ForProvider.Body, c)
	fuzzStringJSON(&s.Spec.InitProvider.Body, c)
	fuzzStringJSON(&s.Status.AtProvider.Body, c)
	fuzzStringJSON(&s.Status.AtProvider.Output, c)
}

func fuzzResourceActionV1Beta1(s *resourcesv1beta1.ResourceAction, c randfill.Continue) {
	c.FillNoCustom(s)
	fuzzStringJSON(&s.Spec.ForProvider.Body, c)
	fuzzStringJSON(&s.Spec.InitProvider.Body, c)
	fuzzStringJSON(&s.Status.AtProvider.Body, c)
	fuzzStringJSON(&s.Status.AtProvider.Output, c)
}

// --- v1beta2 fuzzers ---
// In v1beta2: body, output, sensitiveBody, and sensitiveResponseExportValues are *v1.JSON
// and are handled by apiExtJSONFuzzer (any valid JSON works since they roundtrip as strings
// in v1beta1). responseExportValues is *v1.JSON in v1beta2 but []*string in v1beta1, so the
// JSON must be a valid string array.

func fuzzDataplaneResourceV1Beta2(s *resourcesv1beta2.DataPlaneResource, c randfill.Continue) {
	c.FillNoCustom(s)
	fuzzJSONStringArray(&s.Spec.ForProvider.ResponseExportValues, c)
	fuzzJSONStringArray(&s.Spec.InitProvider.ResponseExportValues, c)
	fuzzJSONStringArray(&s.Status.AtProvider.ResponseExportValues, c)
}

func fuzzResourceV1Beta2(s *resourcesv1beta2.Resource, c randfill.Continue) {
	c.FillNoCustom(s)
	fuzzJSONStringArray(&s.Spec.ForProvider.ResponseExportValues, c)
	fuzzJSONStringArray(&s.Spec.InitProvider.ResponseExportValues, c)
	fuzzJSONStringArray(&s.Status.AtProvider.ResponseExportValues, c)
}

func fuzzUpdateResourceV1Beta2(s *resourcesv1beta2.UpdateResource, c randfill.Continue) {
	c.FillNoCustom(s)
	fuzzJSONStringArray(&s.Spec.ForProvider.ResponseExportValues, c)
	fuzzJSONStringArray(&s.Spec.InitProvider.ResponseExportValues, c)
	fuzzJSONStringArray(&s.Status.AtProvider.ResponseExportValues, c)
}

func fuzzResourceActionV1Beta2(s *resourcesv1beta2.ResourceAction, c randfill.Continue) {
	c.FillNoCustom(s)
	fuzzJSONStringArray(&s.Spec.ForProvider.ResponseExportValues, c)
	fuzzJSONStringArray(&s.Spec.InitProvider.ResponseExportValues, c)
	fuzzJSONStringArray(&s.Status.AtProvider.ResponseExportValues, c)
	// SensitiveResponseExportValues is *v1.JSON in both versions (identity conversion),
	// handled by apiExtJSONFuzzer — no string-array override needed.
}
