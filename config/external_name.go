/*
Copyright 2022 Upbound Inc.
*/

package config

import (
	"context"

	"github.com/Azure/terraform-provider-azapi/xpprovider"
	"github.com/pkg/errors"

	"github.com/crossplane/upjet/v2/pkg/config"
)

// ExternalNameConfigs contains all external name configurations for this
// provider.
var ExternalNameConfigs = map[string]config.ExternalName{
	"azapi_data_plane_resource": dataPlaneResource(),
	"azapi_resource":            azapiResource(),
	"azapi_resource_action":     azapiResourceAction(),
	"azapi_update_resource":     azapiUpdateResource(),
}

func dataPlaneResource() config.ExternalName {
	e := config.IdentifierFromProvider
	e.GetIDFn = func(ctx context.Context, externalName string, parameters map[string]any, terraformProviderConfig map[string]any) (string, error) {
		name, ok := parameters["name"].(string)
		if !ok {
			return "", errors.New("parameter `name` is required")
		}
		parentId, ok := parameters["parent_id"].(string)
		if !ok {
			return "", errors.New("parameter `parent_id` is required")
		}
		resourceType, ok := parameters["type"].(string)
		if !ok {
			return "", errors.New("parameter `type` is required")
		}
		return xpprovider.DataPlaneResourceId(name, parentId, resourceType)
	}
	return e
}

func azapiResource() config.ExternalName {
	e := config.IdentifierFromProvider
	e.GetIDFn = func(ctx context.Context, externalName string, parameters map[string]any, terraformProviderConfig map[string]any) (string, error) {
		name, ok := parameters["name"].(string)
		if !ok {
			return "", errors.New("parameter `name` is required")
		}
		parentId, ok := parameters["parent_id"].(string)
		if !ok {
			return "", errors.New("parameter `parent_id` is required")
		}
		resourceType, ok := parameters["type"].(string)
		if !ok {
			return "", errors.New("parameter `type` is required")
		}
		return xpprovider.NewResourceID(name, parentId, resourceType)
	}
	// Override GetExternalNameFn to handle resources where Azure API doesn't return
	// the standard 'id' field (e.g., Storage Tables). Falls back to constructing
	// the ID from parent_id and name when id is missing.
	// See: https://github.com/upbound/provider-upjet-azapi/issues/XXX
	e.GetExternalNameFn = func(tfstate map[string]any) (string, error) {
		// First try standard id field
		if id, ok := tfstate["id"].(string); ok && id != "" {
			return id, nil
		}
		// Fallback: construct from parent_id + name
		// This handles Azure resources where the API response doesn't include
		// the full ARM resource ID (e.g., Storage Tables, some nested resources)
		parentId, hasParentId := tfstate["parent_id"].(string)
		name, hasName := tfstate["name"].(string)
		if hasParentId && hasName && parentId != "" && name != "" {
			// Construct the resource ID in the format: {parent_id}/{name}
			return parentId + "/" + name, nil
		}
		return "", errors.New("cannot determine resource identity: 'id' field is empty and fallback from 'parent_id'/'name' failed")
	}
	return e
}

func azapiUpdateResource() config.ExternalName {
	e := config.IdentifierFromProvider
	e.GetIDFn = func(ctx context.Context, externalName string, parameters map[string]any, terraformProviderConfig map[string]any) (string, error) {
		resourceType, ok := parameters["type"].(string)
		if !ok {
			return "", errors.New("parameter `type` is required")
		}
		resourceId, ok := parameters["resource_id"].(string)
		if ok && resourceId != "" {
			return xpprovider.ResourceIDWithResourceType(resourceId, resourceType)
		}

		name, ok := parameters["name"].(string)
		if !ok {
			return "", errors.New("parameter `name` is required")
		}
		parentId, ok := parameters["parent_id"].(string)
		if !ok {
			return "", errors.New("parameter `parent_id` is required")
		}
		return xpprovider.NewResourceID(name, parentId, resourceType)
	}
	return e
}

func azapiResourceAction() config.ExternalName {
	e := config.IdentifierFromProvider
	e.GetIDFn = func(ctx context.Context, externalName string, parameters map[string]any, terraformProviderConfig map[string]any) (string, error) {
		resourceType, ok := parameters["resource_type"].(string)
		if !ok {
			return "", errors.New("parameter `type` is required")
		}
		resourceId, ok := parameters["resource_id"].(string)
		if !ok {
			return "", errors.New("parameter `resource_id` is required")
		}
		return xpprovider.ResourceIDWithResourceType(resourceId, resourceType)
	}
	return e
}

// ExternalNameConfigurations applies all external name configs listed in the
// table ExternalNameConfigs and sets the version of those resources to v1beta1
// assuming they will be tested.
func ExternalNameConfigurations() config.ResourceOption {
	return func(r *config.Resource) {
		if e, ok := ExternalNameConfigs[r.Name]; ok {
			r.ExternalName = e
			r.Version = "v1beta1"
		}
	}
}

// cliReconciledExternalNameConfigs contains all external name configurations
// belonging to Terraform resources to be reconciled under the CLI-based
// architecture for this provider.
var cliReconciledExternalNameConfigs = map[string]config.ExternalName{}

// resourceConfigurator applies all external name configs
// listed in the table NoForkExternalNameConfigs and
// cliReconciledExternalNameConfigs and sets the version
// of those resources to v1beta1. For those resource in
// noForkExternalNameConfigs, it also sets
// config.Resource.UseNoForkClient to `true`.
func resourceConfigurator() config.ResourceOption {
	return func(r *config.Resource) {
		// if configured both for the no-fork and CLI based architectures,
		// no-fork configuration prevails
		e, configured := ExternalNameConfigs[r.Name]
		if !configured {
			e, configured = cliReconciledExternalNameConfigs[r.Name]
		}
		if !configured {
			return
		}
		r.Version = "v1beta1"
		r.ExternalName = e
	}
}
