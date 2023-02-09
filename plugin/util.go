// Copyright 2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package plugin

import "reflect"

func GetTypeName[I IPluggable]() string {
	return reflect.TypeOf((*I)(nil)).Elem().Name()
}
