// Copyright 2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package plugin

// IManager defines the interface for managing plugins of a particular type
// (specified by the type parameter).
type IManager[I IPluggable] interface {
	// Init initializes the manager, configuring the RPC channel that will
	// be used to communicate with plugins, and performing plugin
	// discovery.
	Init(name string, ch *RPCChannel[I]) error

	// Close terminates the manager, shutting down any current plugin
	// connections.
	Close() error

	// IsRegisteredMediaType returns true iff the provided mediaType has
	// been registered with the manager as handled by one of the
	// discovered plugins.
	IsRegisteredMediaType(mediaType string) bool

	// GetRegisteredMediaTypes returns a []string of media types that have
	// been registered with the manager by discovered plugins.
	GetRegisteredMediaTypes() []string

	// GetRegisteredAttestationSchemes returns a []string of names for
	// schemes that have been registered with the manager by discovered
	// plugins.
	GetRegisteredAttestationSchemes() []string

	// LookupByMediaType returns a handle (implementation of the managed
	// interface) to the plugin that handles the specified mediaType. If
	// the mediaType is not handled by any of the registered plugins, an
	// error is returned.
	LookupByMediaType(mediaType string) (I, error)

	// LookupByName returns a handle (implementation of the managed
	// interface) to the plugin with the specified name. If there is no
	// such plugin, an error is returned.
	LookupByName(name string) (I, error)

	// LookupByScheme returns a handle (implementation of the managed
	// interface) to the plugin that implements the attestation scheme with
	// the specified name. If there is no such plugin, an error is
	// returned.
	LookupByAttestationScheme(name string) (I, error)
}
