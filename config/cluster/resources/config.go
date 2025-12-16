// SPDX-FileCopyrightText: 2025 Upbound Inc. <https://upbound.io>
//
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	"encoding/json"

	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/upjet/v2/pkg/config"
	"github.com/crossplane/upjet/v2/pkg/config/conversion"
	"github.com/upbound/provider-azapi/v2/apis/cluster/resources/v1beta1"
	"github.com/upbound/provider-azapi/v2/apis/cluster/resources/v1beta2"
)

const (
	group          = "resources"
	versionV1Beta1 = "v1beta1"
	versionV1Beta2 = "v1beta2"
)

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("azapi_data_plane_resource", func(r *config.Resource) {
		r.Kind = "DataPlaneResource"
		r.ShortGroup = group
		r.Version = versionV1Beta2
		r.PreviousVersions = []string{versionV1Beta1}
		r.ControllerReconcileVersion = versionV1Beta2
		r.SetCRDStorageVersion(versionV1Beta1)

		r.Conversions = r.Conversions[1:]
		typeChangingPaths := []string{"body", "output", "responseExportValues"}
		r.Conversions = append(r.Conversions,
			conversion.NewIdentityConversionExpandPaths(conversion.AllVersions, conversion.AllVersions, conversion.DefaultPathPrefixes(), typeChangingPaths...),
			conversion.NewCustomConverter(versionV1Beta1, versionV1Beta2, dataPlaneResourceConverterFromv1beta1Tov1beta2),
			conversion.NewCustomConverter(versionV1Beta2, versionV1Beta1, dataPlaneResourceConverterFromv1beta2Tov1beta1),
		)
	})
	p.AddResourceConfigurator("azapi_resource", func(r *config.Resource) {
		r.Kind = "Resource"
		r.ShortGroup = group

		r.Version = versionV1Beta2
		r.PreviousVersions = []string{versionV1Beta1}
		r.ControllerReconcileVersion = versionV1Beta2
		r.SetCRDStorageVersion(versionV1Beta1)
		r.Conversions = r.Conversions[1:]
		typeChangingPaths := []string{"body", "output", "responseExportValues"}
		r.Conversions = append(r.Conversions,
			conversion.NewIdentityConversionExpandPaths(conversion.AllVersions, conversion.AllVersions, conversion.DefaultPathPrefixes(), typeChangingPaths...),
			conversion.NewCustomConverter(versionV1Beta1, versionV1Beta2, azapiResourceConverterFromv1beta1Tov1beta2),
			conversion.NewCustomConverter(versionV1Beta2, versionV1Beta1, azapiResourceConverterFromv1beta2Tov1beta1),
		)
	})
	p.AddResourceConfigurator("azapi_resource_action", func(r *config.Resource) {
		r.Kind = "ResourceAction"
		r.ShortGroup = group

		r.Version = versionV1Beta2
		r.PreviousVersions = []string{versionV1Beta1}
		r.ControllerReconcileVersion = versionV1Beta2
		r.SetCRDStorageVersion(versionV1Beta1)
		r.Conversions = r.Conversions[1:]
		typeChangingPaths := []string{"body", "output", "responseExportValues"}
		r.Conversions = append(r.Conversions,
			conversion.NewIdentityConversionExpandPaths(conversion.AllVersions, conversion.AllVersions, conversion.DefaultPathPrefixes(), typeChangingPaths...),
			conversion.NewCustomConverter(versionV1Beta1, versionV1Beta2, resourceActionConverterFromv1beta1Tov1beta2),
			conversion.NewCustomConverter(versionV1Beta2, versionV1Beta1, resourceActionConverterFromv1beta2Tov1beta1),
		)

	})
	p.AddResourceConfigurator("azapi_update_resource", func(r *config.Resource) {
		r.Kind = "UpdateResource"
		r.ShortGroup = group
		r.LateInitializer = config.LateInitializer{
			IgnoredFields: []string{"name", "parent_id"},
		}
		r.Version = versionV1Beta2
		r.PreviousVersions = []string{versionV1Beta1}
		r.ControllerReconcileVersion = versionV1Beta2
		r.SetCRDStorageVersion(versionV1Beta1)
		r.Conversions = r.Conversions[1:]
		typeChangingPaths := []string{"body", "output", "responseExportValues"}
		r.Conversions = append(r.Conversions,
			conversion.NewIdentityConversionExpandPaths(conversion.AllVersions, conversion.AllVersions, conversion.DefaultPathPrefixes(), typeChangingPaths...),
			conversion.NewCustomConverter(versionV1Beta1, versionV1Beta2, updateResourceConverterFromv1beta1Tov1beta2),
			conversion.NewCustomConverter(versionV1Beta2, versionV1Beta1, updateResourceConverterFromv1beta2Tov1beta1),
		)

	})
}

func jsonFieldToStringPtr(jf *apiextv1.JSON, sp **string) error {
	if jf == nil {
		return nil
	}
	jBytes, err := jf.MarshalJSON()
	if err != nil {
		return err
	}
	jStr := string(jBytes)
	*sp = &jStr
	return nil
}

func jsonFieldToStringSlice(jf *apiextv1.JSON, s *[]*string) error {
	if jf == nil {
		return nil
	}
	return json.Unmarshal(jf.Raw, s)
}

func dataPlaneResourceConverterFromv1beta2Tov1beta1(src resource.Managed, target resource.Managed) error { //nolint:gocyclo // easier to follow as a unit
	srcTyped := src.(*v1beta2.DataPlaneResource)
	targetTyped := target.(*v1beta1.DataPlaneResource)

	if err := jsonFieldToStringPtr(srcTyped.Spec.ForProvider.Body, &targetTyped.Spec.ForProvider.Body); err != nil {
		return err
	}
	if err := jsonFieldToStringPtr(srcTyped.Spec.InitProvider.Body, &targetTyped.Spec.InitProvider.Body); err != nil {
		return err
	}
	if err := jsonFieldToStringPtr(srcTyped.Status.AtProvider.Body, &targetTyped.Status.AtProvider.Body); err != nil {
		return err
	}

	if err := jsonFieldToStringPtr(srcTyped.Status.AtProvider.Output, &targetTyped.Status.AtProvider.Output); err != nil {
		return err
	}

	if err := jsonFieldToStringSlice(srcTyped.Spec.ForProvider.ResponseExportValues, &targetTyped.Spec.ForProvider.ResponseExportValues); err != nil {
		return err
	}
	if err := jsonFieldToStringSlice(srcTyped.Spec.InitProvider.ResponseExportValues, &targetTyped.Spec.InitProvider.ResponseExportValues); err != nil {
		return err
	}
	if err := jsonFieldToStringSlice(srcTyped.Status.AtProvider.ResponseExportValues, &targetTyped.Status.AtProvider.ResponseExportValues); err != nil {
		return err
	}

	return nil
}

func dataPlaneResourceConverterFromv1beta1Tov1beta2(src resource.Managed, target resource.Managed) error { //nolint:gocyclo // easier to follow as a unit
	srcTyped := src.(*v1beta1.DataPlaneResource)
	targetTyped := target.(*v1beta2.DataPlaneResource)

	if srcTyped.Spec.ForProvider.Body != nil {
		targetTyped.Spec.ForProvider.Body = &apiextv1.JSON{Raw: []byte(*srcTyped.Spec.ForProvider.Body)}
	}
	if srcTyped.Spec.InitProvider.Body != nil {
		targetTyped.Spec.InitProvider.Body = &apiextv1.JSON{Raw: []byte(*srcTyped.Spec.InitProvider.Body)}
	}
	if srcTyped.Status.AtProvider.Body != nil {
		targetTyped.Status.AtProvider.Body = &apiextv1.JSON{Raw: []byte(*srcTyped.Status.AtProvider.Body)}
	}

	if srcTyped.Spec.ForProvider.ResponseExportValues != nil {
		bytes, err := json.Marshal(srcTyped.Spec.ForProvider.ResponseExportValues)
		if err != nil {
			return err
		}
		targetTyped.Spec.ForProvider.ResponseExportValues = &apiextv1.JSON{Raw: bytes}
	}
	if srcTyped.Spec.InitProvider.ResponseExportValues != nil {
		bytes, err := json.Marshal(srcTyped.Spec.InitProvider.ResponseExportValues)
		if err != nil {
			return err
		}
		targetTyped.Spec.InitProvider.ResponseExportValues = &apiextv1.JSON{Raw: bytes}
	}
	if srcTyped.Status.AtProvider.ResponseExportValues != nil {
		bytes, err := json.Marshal(srcTyped.Status.AtProvider.ResponseExportValues)
		if err != nil {
			return err
		}
		targetTyped.Status.AtProvider.ResponseExportValues = &apiextv1.JSON{Raw: bytes}
	}

	if srcTyped.Status.AtProvider.Output != nil {
		targetTyped.Status.AtProvider.Output = &apiextv1.JSON{Raw: []byte(*srcTyped.Status.AtProvider.Output)}
	}
	return nil
}

func azapiResourceConverterFromv1beta2Tov1beta1(src resource.Managed, target resource.Managed) error { //nolint:gocyclo // easier to follow as a unit
	srcTyped := src.(*v1beta2.Resource)
	targetTyped := target.(*v1beta1.Resource)

	if err := jsonFieldToStringPtr(srcTyped.Spec.ForProvider.Body, &targetTyped.Spec.ForProvider.Body); err != nil {
		return err
	}
	if err := jsonFieldToStringPtr(srcTyped.Spec.InitProvider.Body, &targetTyped.Spec.InitProvider.Body); err != nil {
		return err
	}
	if err := jsonFieldToStringPtr(srcTyped.Status.AtProvider.Body, &targetTyped.Status.AtProvider.Body); err != nil {
		return err
	}

	if err := jsonFieldToStringPtr(srcTyped.Status.AtProvider.Output, &targetTyped.Status.AtProvider.Output); err != nil {
		return err
	}

	if err := jsonFieldToStringSlice(srcTyped.Spec.ForProvider.ResponseExportValues, &targetTyped.Spec.ForProvider.ResponseExportValues); err != nil {
		return err
	}
	if err := jsonFieldToStringSlice(srcTyped.Spec.InitProvider.ResponseExportValues, &targetTyped.Spec.InitProvider.ResponseExportValues); err != nil {
		return err
	}
	if err := jsonFieldToStringSlice(srcTyped.Status.AtProvider.ResponseExportValues, &targetTyped.Status.AtProvider.ResponseExportValues); err != nil {
		return err
	}

	return nil
}

func azapiResourceConverterFromv1beta1Tov1beta2(src resource.Managed, target resource.Managed) error { //nolint:gocyclo // easier to follow as a unit
	srcTyped := src.(*v1beta1.Resource)
	targetTyped := target.(*v1beta2.Resource)

	if srcTyped.Spec.ForProvider.Body != nil {
		targetTyped.Spec.ForProvider.Body = &apiextv1.JSON{Raw: []byte(*srcTyped.Spec.ForProvider.Body)}
	}
	if srcTyped.Spec.InitProvider.Body != nil {
		targetTyped.Spec.InitProvider.Body = &apiextv1.JSON{Raw: []byte(*srcTyped.Spec.InitProvider.Body)}
	}
	if srcTyped.Status.AtProvider.Body != nil {
		targetTyped.Status.AtProvider.Body = &apiextv1.JSON{Raw: []byte(*srcTyped.Status.AtProvider.Body)}
	}

	if srcTyped.Spec.ForProvider.ResponseExportValues != nil {
		bytes, err := json.Marshal(srcTyped.Spec.ForProvider.ResponseExportValues)
		if err != nil {
			return err
		}
		targetTyped.Spec.ForProvider.ResponseExportValues = &apiextv1.JSON{Raw: bytes}
	}
	if srcTyped.Spec.InitProvider.ResponseExportValues != nil {
		bytes, err := json.Marshal(srcTyped.Spec.InitProvider.ResponseExportValues)
		if err != nil {
			return err
		}
		targetTyped.Spec.InitProvider.ResponseExportValues = &apiextv1.JSON{Raw: bytes}
	}
	if srcTyped.Status.AtProvider.ResponseExportValues != nil {
		bytes, err := json.Marshal(srcTyped.Status.AtProvider.ResponseExportValues)
		if err != nil {
			return err
		}
		targetTyped.Status.AtProvider.ResponseExportValues = &apiextv1.JSON{Raw: bytes}
	}
	if srcTyped.Status.AtProvider.Output != nil {
		targetTyped.Status.AtProvider.Output = &apiextv1.JSON{Raw: []byte(*srcTyped.Status.AtProvider.Output)}
	}
	return nil
}

func resourceActionConverterFromv1beta2Tov1beta1(src resource.Managed, target resource.Managed) error { //nolint:gocyclo // easier to follow as a unit
	srcTyped := src.(*v1beta2.ResourceAction)
	targetTyped := target.(*v1beta1.ResourceAction)

	if err := jsonFieldToStringPtr(srcTyped.Spec.ForProvider.Body, &targetTyped.Spec.ForProvider.Body); err != nil {
		return err
	}
	if err := jsonFieldToStringPtr(srcTyped.Spec.InitProvider.Body, &targetTyped.Spec.InitProvider.Body); err != nil {
		return err
	}
	if err := jsonFieldToStringPtr(srcTyped.Status.AtProvider.Body, &targetTyped.Status.AtProvider.Body); err != nil {
		return err
	}

	if err := jsonFieldToStringPtr(srcTyped.Status.AtProvider.Output, &targetTyped.Status.AtProvider.Output); err != nil {
		return err
	}

	if err := jsonFieldToStringSlice(srcTyped.Spec.ForProvider.ResponseExportValues, &targetTyped.Spec.ForProvider.ResponseExportValues); err != nil {
		return err
	}
	if err := jsonFieldToStringSlice(srcTyped.Spec.InitProvider.ResponseExportValues, &targetTyped.Spec.InitProvider.ResponseExportValues); err != nil {
		return err
	}
	if err := jsonFieldToStringSlice(srcTyped.Status.AtProvider.ResponseExportValues, &targetTyped.Status.AtProvider.ResponseExportValues); err != nil {
		return err
	}

	return nil
}

func resourceActionConverterFromv1beta1Tov1beta2(src resource.Managed, target resource.Managed) error { //nolint:gocyclo // easier to follow as a unit
	srcTyped := src.(*v1beta1.ResourceAction)
	targetTyped := target.(*v1beta2.ResourceAction)

	if srcTyped.Spec.ForProvider.Body != nil {
		targetTyped.Spec.ForProvider.Body = &apiextv1.JSON{Raw: []byte(*srcTyped.Spec.ForProvider.Body)}
	}
	if srcTyped.Spec.InitProvider.Body != nil {
		targetTyped.Spec.InitProvider.Body = &apiextv1.JSON{Raw: []byte(*srcTyped.Spec.InitProvider.Body)}
	}
	if srcTyped.Status.AtProvider.Body != nil {
		targetTyped.Status.AtProvider.Body = &apiextv1.JSON{Raw: []byte(*srcTyped.Status.AtProvider.Body)}
	}

	if srcTyped.Spec.ForProvider.ResponseExportValues != nil {
		bytes, err := json.Marshal(srcTyped.Spec.ForProvider.ResponseExportValues)
		if err != nil {
			return err
		}
		targetTyped.Spec.ForProvider.ResponseExportValues = &apiextv1.JSON{Raw: bytes}
	}
	if srcTyped.Spec.InitProvider.ResponseExportValues != nil {
		bytes, err := json.Marshal(srcTyped.Spec.InitProvider.ResponseExportValues)
		if err != nil {
			return err
		}
		targetTyped.Spec.InitProvider.ResponseExportValues = &apiextv1.JSON{Raw: bytes}
	}
	if srcTyped.Status.AtProvider.ResponseExportValues != nil {
		bytes, err := json.Marshal(srcTyped.Status.AtProvider.ResponseExportValues)
		if err != nil {
			return err
		}
		targetTyped.Status.AtProvider.ResponseExportValues = &apiextv1.JSON{Raw: bytes}
	}
	if srcTyped.Status.AtProvider.Output != nil {
		targetTyped.Status.AtProvider.Output = &apiextv1.JSON{Raw: []byte(*srcTyped.Status.AtProvider.Output)}
	}
	return nil
}

func updateResourceConverterFromv1beta2Tov1beta1(src resource.Managed, target resource.Managed) error { //nolint:gocyclo // easier to follow as a unit
	srcTyped := src.(*v1beta2.UpdateResource)
	targetTyped := target.(*v1beta1.UpdateResource)

	if err := jsonFieldToStringPtr(srcTyped.Spec.ForProvider.Body, &targetTyped.Spec.ForProvider.Body); err != nil {
		return err
	}
	if err := jsonFieldToStringPtr(srcTyped.Spec.InitProvider.Body, &targetTyped.Spec.InitProvider.Body); err != nil {
		return err
	}
	if err := jsonFieldToStringPtr(srcTyped.Status.AtProvider.Body, &targetTyped.Status.AtProvider.Body); err != nil {
		return err
	}

	if err := jsonFieldToStringPtr(srcTyped.Status.AtProvider.Output, &targetTyped.Status.AtProvider.Output); err != nil {
		return err
	}

	if err := jsonFieldToStringSlice(srcTyped.Spec.ForProvider.ResponseExportValues, &targetTyped.Spec.ForProvider.ResponseExportValues); err != nil {
		return err
	}
	if err := jsonFieldToStringSlice(srcTyped.Spec.InitProvider.ResponseExportValues, &targetTyped.Spec.InitProvider.ResponseExportValues); err != nil {
		return err
	}
	if err := jsonFieldToStringSlice(srcTyped.Status.AtProvider.ResponseExportValues, &targetTyped.Status.AtProvider.ResponseExportValues); err != nil {
		return err
	}

	return nil
}

func updateResourceConverterFromv1beta1Tov1beta2(src resource.Managed, target resource.Managed) error { //nolint:gocyclo // easier to follow as a unit
	srcTyped := src.(*v1beta1.UpdateResource)
	targetTyped := target.(*v1beta2.UpdateResource)

	if srcTyped.Spec.ForProvider.Body != nil {
		targetTyped.Spec.ForProvider.Body = &apiextv1.JSON{Raw: []byte(*srcTyped.Spec.ForProvider.Body)}
	}
	if srcTyped.Spec.InitProvider.Body != nil {
		targetTyped.Spec.InitProvider.Body = &apiextv1.JSON{Raw: []byte(*srcTyped.Spec.InitProvider.Body)}
	}
	if srcTyped.Status.AtProvider.Body != nil {
		targetTyped.Status.AtProvider.Body = &apiextv1.JSON{Raw: []byte(*srcTyped.Status.AtProvider.Body)}
	}

	if srcTyped.Spec.ForProvider.ResponseExportValues != nil {
		bytes, err := json.Marshal(srcTyped.Spec.ForProvider.ResponseExportValues)
		if err != nil {
			return err
		}
		targetTyped.Spec.ForProvider.ResponseExportValues = &apiextv1.JSON{Raw: bytes}
	}
	if srcTyped.Spec.InitProvider.ResponseExportValues != nil {
		bytes, err := json.Marshal(srcTyped.Spec.InitProvider.ResponseExportValues)
		if err != nil {
			return err
		}
		targetTyped.Spec.InitProvider.ResponseExportValues = &apiextv1.JSON{Raw: bytes}
	}
	if srcTyped.Status.AtProvider.ResponseExportValues != nil {
		bytes, err := json.Marshal(srcTyped.Status.AtProvider.ResponseExportValues)
		if err != nil {
			return err
		}
		targetTyped.Status.AtProvider.ResponseExportValues = &apiextv1.JSON{Raw: bytes}
	}

	if srcTyped.Status.AtProvider.Output != nil {
		targetTyped.Status.AtProvider.Output = &apiextv1.JSON{Raw: []byte(*srcTyped.Status.AtProvider.Output)}
	}
	return nil
}
