// SPDX-FileCopyrightText: 2025 Upbound Inc. <https://upbound.io>
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/upjet/v2/pkg/controller"

	providerconfig "github.com/upbound/provider-azapi/internal/controller/cluster/providerconfig"
	dataplaneresource "github.com/upbound/provider-azapi/internal/controller/cluster/resources/dataplaneresource"
	resource "github.com/upbound/provider-azapi/internal/controller/cluster/resources/resource"
	resourceaction "github.com/upbound/provider-azapi/internal/controller/cluster/resources/resourceaction"
	updateresource "github.com/upbound/provider-azapi/internal/controller/cluster/resources/updateresource"
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

// SetupGated creates all controllers with the supplied logger and adds them to
// the supplied manager gated.
func SetupGated(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		providerconfig.SetupGated,
		dataplaneresource.SetupGated,
		resource.SetupGated,
		resourceaction.SetupGated,
		updateresource.SetupGated,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}
