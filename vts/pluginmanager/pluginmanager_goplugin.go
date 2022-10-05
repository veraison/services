// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package pluginmanager

import (
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/hashicorp/go-plugin"
	"github.com/setrofim/viper"
	"github.com/veraison/services/config"
	"github.com/veraison/services/proto"
	"github.com/veraison/services/scheme"
)

type cfg struct {
	Backend        string
	BackendConfigs map[string]interface{} `mapstructure:",remain"`
}

func (o cfg) Validate() error {
	supportedBackends := map[string]bool{
		"go-plugin": true,
	}

	var unexpected []string
	for k := range o.BackendConfigs {
		if _, ok := supportedBackends[k]; !ok {
			unexpected = append(unexpected, k)
		}
	}

	if len(unexpected) > 0 {
		sort.Strings(unexpected)
		return fmt.Errorf("unexpected directives: %s", strings.Join(unexpected, ", "))
	}

	return nil
}

type backendCfg struct {
	Folder string
}

type GoPluginManager struct {
	Backend       string
	DispatchTable map[string]*scheme.SchemeGoPlugin
}

func New() *GoPluginManager {
	return &GoPluginManager{}
}

// variables read from the config store:
//   * "go-plugin.folder"
func (o *GoPluginManager) Init(v *viper.Viper) error {
	cfg := cfg{Backend: "go-plugin"}
	loader := config.NewLoader(&cfg)
	if err := loader.LoadFromViper(v); err != nil {
		return err
	}

	subs, err := config.GetSubs(v, "go-plugin")
	if err != nil {
		return err
	}

	var backendCfg backendCfg
	loader = config.NewLoader(&backendCfg)
	if err := loader.LoadFromViper(subs["go-plugin"]); err != nil {
		return err
	}

	pPaths, err := plugin.Discover("*", backendCfg.Folder)
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
