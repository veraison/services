// Copyright 2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	_ "github.com/mattn/go-sqlite3"
	"github.com/veraison/services/config"
	"github.com/veraison/services/log"
	"github.com/veraison/services/management"
	"github.com/veraison/services/management/api"
)

var (
	DefaultListenAddr = "localhost:8088"
)

type cfg struct {
	ListenAddr string `mapstructure:"listen-addr" valid:"dialstring"`
}

func main() {
	config.CmdLine()

	v, err := config.ReadRawConfig(*config.File, false)
	if err != nil {
		log.Fatalf("Could not read config: %v", err)
	}

	subs, err := config.GetSubs(v, "*management", "*logging")
	if err != nil {
		log.Fatalf("Could not parse config: %v", err)
	}

	classifiers := map[string]interface{}{"service": "management"}
	if err := log.Init(subs["logging"], classifiers); err != nil {
		log.Fatalf("could not configure logging: %v", err)
	}
	log.InitGinWriter() // route gin output to our logger.

	log.Infow("Initializing Management Service", "version", config.Version)

	log.Info("initializing policy manager")
	pm, err := management.CreatePolicyManagerFromConfig(v, "policy")
	if err != nil {
		log.Fatalf("could not init policy manager: %v", err)
	}

	cfg := cfg{ListenAddr: DefaultListenAddr}
	loader := config.NewLoader(&cfg)
	if err := loader.LoadFromViper(subs["management"]); err != nil {
		log.Fatalf("Could not load verfication config: %v", err)

	}

	log.Infow("initializing management API service", "address", cfg.ListenAddr)
	handler := api.NewHandler(pm, log.Named("api"))
	if err := api.NewRouter(handler).Run(cfg.ListenAddr); err != nil {
		log.Fatalf("Gin engine failed: %v", err)
	}
}
