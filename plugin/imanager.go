// Copyright 2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package plugin

type IManager[I IPluggable] interface {
	Init(name string, ch RPCChannel[I]) error
	Close() error

	IsRegisteredMediaType(mediaType string) bool
	GetRegisteredMediaTypes() []string
	LookupByMediaType(mediaType string) (I, error)
	LookupByName(name string) (I, error)
}

