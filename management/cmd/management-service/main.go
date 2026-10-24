// Copyright 2023-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	_ "github.com/mattn/go-sqlite3"
	"github.com/veraison/services/auth"
	"github.com/veraison/services/config"
	"github.com/veraison/services/log"
	"github.com/veraison/services/management"
	"github.com/veraison/services/management/api"
)

var (
	DefaultListenAddr = "localhost:10443"
)

type cfg struct {
	ListenAddr     string `mapstructure:"listen-addr" valid:"dialstring"`
	Protocol       string `mapstructure:"protocol" valid:"in(http|https)"`
	Cert           string `mapstructure:"cert"`
	CertKey        string `mapstructure:"cert-key"`
	MaxPayloadSize string `mapstructure:"max-payload-size"`
}

func main() {
	config.CmdLine()

	v, err := config.ReadRawConfig(*config.File, false)
	if err != nil {
		log.Fatalf("Could not read config: %v", err)
	}

	subs, err := config.GetSubs(v, "*management", "*logging", "*auth")
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

	cfg := cfg{
		ListenAddr:     DefaultListenAddr,
		Protocol:       "https",
		Cert:           "[unset]",
		CertKey:        "[unset]",
		MaxPayloadSize: "1MB",
	}
	loader := config.NewLoader(&cfg)
	if err := loader.LoadFromViper(subs["management"]); err != nil {
		log.Fatalf("Could not load verification config: %v", err)

	}

	maxPayloadSize, err := config.ParseBytesSize(cfg.MaxPayloadSize)
	if err != nil {
		log.Fatalf("Could not parse max-payload-size: %v", err)
	}

	authorizer, err := auth.NewAuthorizer(subs["auth"], log.Named("auth"))
	if err != nil {
		log.Fatalf("could not init authorizer: %v", err)
	}
	defer func() {
		err := authorizer.Close()
		if err != nil {
			log.Errorf("Could not close authorizer: %v", err)
		}
	}()

	handler := api.NewHandler(pm, log.Named("api"), maxPayloadSize)

	if cfg.Protocol == "https" {
		apiServerTLS(handler, authorizer, cfg.ListenAddr, cfg.Cert, cfg.CertKey)
	} else {
		apiServer(handler, authorizer, cfg.ListenAddr)
	}
}

func apiServer(apiHandler api.Handler, authorizer auth.IAuthorizer, listenAddr string) {
	log.Infow("initializing management API HTTP service", "address", listenAddr)
	if err := api.NewRouter(apiHandler, authorizer).Run(listenAddr); err != nil {
		log.Fatalf("Gin engine failed: %v", err)
	}
}

func apiServerTLS(apiHandler api.Handler, authorizer auth.IAuthorizer, listenAddr, certFile, keyFile string) {
	log.Infow("initializing management API HTTPS service", "address", listenAddr)

	if err := api.NewRouter(apiHandler, authorizer).RunTLS(listenAddr, certFile, keyFile); err != nil {
		log.Fatalf("Gin engine failed: %v", err)
	}
}
