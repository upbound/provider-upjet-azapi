// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: CC0-1.0

package resource

import (
	"github.com/crossplane/upjet/pkg/config"
)

// Configure configures resource group
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("azapi_data_plane_resource", func(r *config.Resource) {
		r.Kind = "DataPlaneResource"
		r.ShortGroup = "resource"
	})

	p.AddResourceConfigurator("azapi_resource", func(r *config.Resource) {
		r.Kind = "Resource"
		r.ShortGroup = "resource"
	})

	p.AddResourceConfigurator("azapi_resource_action", func(r *config.Resource) {
		r.Kind = "ResourceAction"
		r.ShortGroup = "resource"
	})

	p.AddResourceConfigurator("azapi_update_resource", func(r *config.Resource) {
		r.Kind = "UpdateResource"
		r.ShortGroup = "resource"
	})
}
