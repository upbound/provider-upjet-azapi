// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package apis

import (
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

var s = runtime.NewScheme()

// BuildScheme builds the module's type registry using the specified
// runtime.SchemeBuilder.
func BuildScheme(sb runtime.SchemeBuilder) error {
	return errors.Wrap(sb.AddToScheme(s), "failed to register the GVKs with the runtime scheme")
}
