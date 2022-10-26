// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"context"

	"github.com/veraison/services/config"
	"github.com/veraison/services/log"
	"github.com/veraison/services/verification/api"
	"github.com/veraison/services/verification/sessionmanager"
	"github.com/veraison/services/verification/verifier"
	"github.com/veraison/services/vtsclient"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	DefaultListenAddr = "localhost:8080"
)

type cfg struct {
	ListenAddr string `mapstructure:"listen-addr" valid:"dialstring"`
}

func main() {
	v, err := config.ReadRawConfig("", true)
	if err != nil {
		log.Fatalf("Could not read config: %v", err)
	}

	subs, err := config.GetSubs(v, "*vts", "*verifier", "*verification", "*logging")
	if err != nil {
		log.Fatalf("Could not read config: %v", err)
	}

	classifiers := map[string]interface{}{"service": "verification"}
	if err := log.Init(subs["logging"], classifiers); err != nil {
		log.Fatalf("could not configure logging: %v", err)
	}
	log.InitGinWriter() // route gin output to our logger.

	log.Infow("Initializing Verification Service", "version", config.Version)

	sessionManager := sessionmanager.NewSessionManagerTTLCache()

	log.Info("initializing VTS client")
	vtsClient := vtsclient.NewGRPC()
	if err := vtsClient.Init(subs["vts"]); err != nil {
		log.Fatalf("Could not initialize VTS client: %v", err)
	}

	vtsState, err := vtsClient.GetServiceState(context.TODO(), &emptypb.Empty{})
	if err == nil {
		log.Infow("vts connection established", "server-version", vtsState.ServerVersion)
	} else {
		log.Warnw("Could not connect to VTS server. If you do not expect the server to be running yet, this is probably OK, otherwise it may indicate an issue with vts.server-addr in your settings",
			"error", err)
	}

	log.Info("initializing verifier")
	verifier := verifier.New(subs["verifier"], vtsClient)

	apiHandler := api.NewHandler(sessionManager, verifier)

	cfg := cfg{ListenAddr: DefaultListenAddr}
	loader := config.NewLoader(&cfg)
	if err := loader.LoadFromViper(subs["verification"]); err != nil {
		log.Fatalf("Could not load verfication config: %v", err)

	}

	log.Infow("initializing verification API service", "address", cfg.ListenAddr)
	apiServer(apiHandler, cfg.ListenAddr)
}

func apiServer(apiHandler api.IHandler, listenAddr string) {
	if err := api.NewRouter(apiHandler).Run(listenAddr); err != nil {
		log.Fatalf("Gin engine failed: %v", err)
	}
}
