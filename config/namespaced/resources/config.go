// SPDX-FileCopyrightText: 2025 Upbound Inc. <https://upbound.io>
//
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	"github.com/crossplane/upjet/v2/pkg/config"
)

const group = "resources"

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("azapi_data_plane_resource", func(r *config.Resource) {
		r.Kind = "DataPlaneResource"
		r.ShortGroup = group
	})
	p.AddResourceConfigurator("azapi_resource", func(r *config.Resource) {
		r.Kind = "Resource"
		r.ShortGroup = group

	})
	p.AddResourceConfigurator("azapi_resource_action", func(r *config.Resource) {
		r.Kind = "ResourceAction"
		r.ShortGroup = group
	})
	p.AddResourceConfigurator("azapi_update_resource", func(r *config.Resource) {
		r.Kind = "UpdateResource"
		r.ShortGroup = group
		r.LateInitializer = config.LateInitializer{
			IgnoredFields: []string{"name", "parent_id"},
		}
	})
}
