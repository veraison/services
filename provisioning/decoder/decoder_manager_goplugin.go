// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package decoder

import (
	"fmt"
	"os/exec"

	"github.com/hashicorp/go-plugin"
	"github.com/veraison/services/log"
	"go.uber.org/zap"
)

var handshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "VERAISON_PROVISIONING_DECODER_PLUGIN",
	MagicCookieValue: "VERAISON",
}

// pluginMap is the map of plugins we can dispense.
var pluginMap = map[string]plugin.Plugin{
	"decoder": &Plugin{},
}

type GoPluginDecoderManager struct {
	DispatchTable map[string]*GoPluginDecoderContext

	logger *zap.SugaredLogger
}

func (o *GoPluginDecoderManager) Init(dir string, logger *zap.SugaredLogger) error {
	o.logger = logger

	// TODO(tho) might want to define a naming convention for endorsement
	// decoder plugins
	o.logger.Debugw("discovering plugins", "location", dir)
	pPaths, err := plugin.Discover("*", dir)
	if err != nil {
		return err
	}

	o.DispatchTable = make(map[string]*GoPluginDecoderContext)

	for _, p := range pPaths {

		ctx, err := NewGoPluginDecoderContext(p, o.logger)
		if err != nil {
			return err
		}

		for _, mt := range ctx.supportedMediaTypes {
			// TODO(tho) check if this same media type has been already
			// advertised by another plugin.  Should raise fatal error if this
			// is the case.
			o.DispatchTable[mt] = ctx
			o.logger.Infow("media type registred", "media-type", mt)
		}
	}

	if len(o.DispatchTable) > 0 {
		o.logger.Infof("found decoder plugins for %d media types", len(o.DispatchTable))
	} else {
		o.logger.Warn("did not find any decoder plugins")
	}

	return nil
}

func (o GoPluginDecoderManager) Close() error {
	for _, v := range o.DispatchTable {
		if v.client != nil {
			o.logger.Debugw("killing client", "client", v.name)
			v.client.Kill()
		}
	}
	return nil
}

func (o GoPluginDecoderManager) Dispatch(
	mediaType string,
	data []byte,
) (*EndorsementDecoderResponse, error) {
	ctx, ok := o.DispatchTable[mediaType]
	if !ok || ctx.handle == nil {
		return nil, fmt.Errorf("no active plugin found for media type %s", mediaType)
	}

	return ctx.handle.Decode(data)
}

func (o GoPluginDecoderManager) IsSupportedMediaType(mediaType string) bool {
	_, ok := o.DispatchTable[mediaType]

	return ok
}

func (o GoPluginDecoderManager) GetSupportedMediaTypes() []string {
	var a []string

	for k := range o.DispatchTable {
		a = append(a, k)
	}

	return a
}

type GoPluginDecoderContext struct {
	path                string
	name                string
	supportedMediaTypes []string
	handle              IDecoder
	client              *plugin.Client
}

func NewGoPluginDecoderContext(
	path string,
	logger *zap.SugaredLogger,
) (*GoPluginDecoderContext, error) {
	client := plugin.NewClient(
		&plugin.ClientConfig{
			HandshakeConfig: handshakeConfig,
			Plugins:         pluginMap,
			Cmd:             exec.Command(path),
			Logger:          log.NewLogger(logger),
		},
	)

	rpcClient, err := client.Client()
	if err != nil {
		client.Kill()
		return nil, fmt.Errorf(
			"unable to create the RPC client for %s: %w",
			path, err,
		)
	}

	protocolClient, err := rpcClient.Dispense("decoder")
	if err != nil {
		client.Kill()
		return nil, fmt.Errorf(
			"unable to create a new instance of plugin %s: %w",
			path, err,
		)
	}

	handle, ok := protocolClient.(IDecoder)
	if !ok {
		client.Kill()
		return nil, fmt.Errorf(
			"plugin %s does not provide an implementation of the endorsement decoder interface",
			path,
		)
	}

	return &GoPluginDecoderContext{
		path:                path,
		name:                handle.GetName(),
		supportedMediaTypes: handle.GetSupportedMediaTypes(),
		handle:              handle,
		client:              client,
	}, nil
}
