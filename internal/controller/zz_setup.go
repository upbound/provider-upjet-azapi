// SPDX-FileCopyrightText: 2025 Upbound Inc. <https://upbound.io>
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/upjet/pkg/controller"

	providerconfig "github.com/upbound/provider-azapi/internal/controller/providerconfig"
	dataplaneresource "github.com/upbound/provider-azapi/internal/controller/resources/dataplaneresource"
	resource "github.com/upbound/provider-azapi/internal/controller/resources/resource"
	resourceaction "github.com/upbound/provider-azapi/internal/controller/resources/resourceaction"
	updateresource "github.com/upbound/provider-azapi/internal/controller/resources/updateresource"
)

// Setup creates all controllers with the supplied logger and adds them to
// the supplied manager.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		providerconfig.Setup,
		dataplaneresource.Setup,
		resource.Setup,
		resourceaction.Setup,
		updateresource.Setup,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}
