// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package pluginmanager

type IPluginManager interface {
	Init(dir string) error
	Close() error
	// TODO
}
