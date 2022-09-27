// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package pluginmanager

import (
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/go-plugin"
	"github.com/spf13/viper"
	"github.com/veraison/services/proto"
	"github.com/veraison/services/scheme"
)

type GoPluginManager struct {
	Config        *viper.Viper
	DispatchTable map[string]*scheme.SchemeGoPlugin
}

func New(v *viper.Viper) *GoPluginManager {
	return &GoPluginManager{
		Config: v,
	}
}

// variables read from the config store:
//   * "go-plugin.folder"
func (o *GoPluginManager) Init() error {
	defaultBackend := "go-plugin"
	o.Config.SetDefault("backend", defaultBackend)

	backend := o.Config.GetString("backend")
	if backend != defaultBackend {
		return fmt.Errorf("want backend %s, got %s", defaultBackend, backend)
	}

	dir := o.Config.GetString("go-plugin.folder")
	if dir == "" {
		return fmt.Errorf(`"go-pluing.folder" not specified`)
	}

	pPaths, err := plugin.Discover("*", dir)
	if err != nil {
		return err
	}

	tbl := make(map[string]*scheme.SchemeGoPlugin)

	for _, p := range pPaths {
		ctx, err := scheme.NewSchemeGoPlugin(p)
		if err != nil {
			return err
		}

		for _, mt := range ctx.SupportedMediaTypes {
			// TODO(tho) check if this same media type has been already
			// advertised by another plugin.  Should raise fatal error if this
			// is the case.
			tbl[mt] = ctx
		}
	}

	o.DispatchTable = tbl

	return nil
}
func (o *GoPluginManager) Close() error {
	for _, v := range o.DispatchTable {
		if v.Client != nil {
			log.Printf("killing client %s", v.Name)
			v.Client.Kill()
		}
	}
	return nil
}

// GetPlugin returns the handle of the IScheme implementation
func (o *GoPluginManager) LookupByMediaType(mediaType string) (scheme.IScheme, error) {
	ctx, ok := o.DispatchTable[mediaType]
	if !ok || ctx.Handle == nil {
		return nil, fmt.Errorf("no active plugin found for media type %s", mediaType)
	}

	return ctx.Handle, nil
}

func (o *GoPluginManager) LookupByAttestationFormat(format proto.AttestationFormat) (scheme.IScheme, error) {
	// XXX this is obviously sub-optimal, however since this interface will go away once
	// XXX the provisioning plugins are migrated under VTS, we don't bother
	for _, v := range o.DispatchTable {
		if v.Handle.GetFormat() == format {
			return v.Handle, nil
		}
	}

	return nil, fmt.Errorf("no active plugin found for format %s", format.String())
}

func (o *GoPluginManager) SupportedVerificationMediaTypes() ([]string, error) {
	var a []string

	if o.DispatchTable == nil {
		return nil, errors.New("nil dispatch table")
	}

	for k := range o.DispatchTable {
		a = append(a, k)
	}

	return a, nil
}
