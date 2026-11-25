// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package scheme

import (
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/veraison/services/pkg/fault"
	"github.com/veraison/services/pkg/plugin"
	"github.com/veraison/services/pkg/util"
	"go.uber.org/zap"
)

type Manager struct {
	params                     plugin.ParametersMap
	logger                     *zap.SugaredLogger
	handlerByAttestationScheme map[string]IHandler
	handlerByEvidenceMediaType map[string]IHandler
	handlerByCorimProfile      map[string]IHandler
}

func CreateManager(
	loader plugin.ILoader,
	params plugin.ParametersMap,
	logger *zap.SugaredLogger,
) (*Manager, error) {
	manager := NewManager(params, logger)
	if err := manager.UpdateFromLoader(loader); err != nil {
		return nil, err
	}

	return manager, nil
}

func NewManager(params plugin.ParametersMap, logger *zap.SugaredLogger) *Manager {
	return &Manager{
		params:                     params,
		logger:                     logger,
		handlerByAttestationScheme: make(map[string]IHandler),
		handlerByEvidenceMediaType: make(map[string]IHandler),
		handlerByCorimProfile:      make(map[string]IHandler),
	}
}

func (o *Manager) UpdateFromLoader(loader plugin.ILoader) error {
	var err error
	for _, loadedPlugin := range loader.Plugins() {
		pluggable, ok := loadedPlugin.Pluggables()[PluginType]
		if !ok {
			continue
		}

		handlerName := pluggable.GetName()
		handler, ok := pluggable.(IHandler)
		if !ok {
			panic(fmt.Sprintf(
				"%s in %s does not implement scheme.IHander",
				PluginType,
				loadedPlugin.Source(),
			))
		}

		params, ok := o.params[handlerName]
		if !ok {
			params = plugin.NewParameters()
		}

		if err := handler.Init(params); err != nil {
			if errors.Is(err, fault.ErrSkippedPlugin) {
				o.logger.Debugf("skipping %s", handlerName)
				if err := loader.Unload(loadedPlugin, PluginType); err != nil {
					// will not get here since we know the loadedPlugin came
					// from this loader and it implements PluginType
					panic(err)
				}

				continue
			}

			return fmt.Errorf(
				"could not initialize %s: %w",
				handlerName,
				err,
			)
		}
		o.logger.Debugf("initialized %s", handlerName)

		schemeName := handler.GetAttestationScheme()
		if other, ok := o.handlerByAttestationScheme[schemeName]; ok {
			return fmt.Errorf(
				"%s and %s both provide attestation scheme %s",
				handlerName,
				other.GetName(),
				schemeName,
			)
		}
		o.handlerByAttestationScheme[schemeName] = handler

		for _, mediaType := range handler.GetSupportedEvidenceMediaTypes() {
			mediaType, err = util.NormalizeMediaType(mediaType)
			if err != nil {
				return fmt.Errorf(
					"%s contains an invalid media type %q",
					handlerName,
					mediaType,
				)
			}

			if other, ok := o.handlerByEvidenceMediaType[schemeName]; ok {
				return fmt.Errorf(
					"%s and %s both handle evidence media type %s",
					handlerName,
					other.GetName(),
					mediaType,
				)
			}
			o.handlerByEvidenceMediaType[mediaType] = handler
		}

		for _, profile := range handler.GetSupportedCorimProfiles() {
			if other, ok := o.handlerByCorimProfile[schemeName]; ok {
				return fmt.Errorf(
					"%s and %s both handle CoRIM profile %s",
					handlerName,
					other.GetName(),
					profile,
				)
			}
			o.handlerByCorimProfile[profile] = handler
		}
	}

	return nil
}

func (o *Manager) GetRegisteredAttestationSchemes() []string {
	return slices.Collect(maps.Keys(o.handlerByAttestationScheme))
}

func (o *Manager) GetRegisteredCorimProfiles() []string {
	return slices.Collect(maps.Keys(o.handlerByCorimProfile))
}

func (o *Manager) GetRegisteredEvidenceMediaTypes() []string {
	return slices.Collect(maps.Keys(o.handlerByEvidenceMediaType))
}

func (o *Manager) LookupByEvidenceMediaType(mediaType string) (IHandler, error) {
	handler, ok := o.handlerByEvidenceMediaType[mediaType]
	if !ok {
		return nil, fault.NotFound("handler for evidence media type %q", mediaType)
	}

	return handler, nil
}

func (o *Manager) LookupByCorimProfile(mediaType string) (IHandler, error) {
	handler, ok := o.handlerByCorimProfile[mediaType]
	if !ok {
		return nil, fault.NotFound("handler for CoRIM profile %q", mediaType)
	}

	return handler, nil
}

func (o *Manager) Close() {
	for _, handler := range o.handlerByAttestationScheme {
		if err := handler.Fini(); err != nil {
			o.logger.Errorf("closing %q: %s", handler.GetName(), err.Error())
		}
	}
}
