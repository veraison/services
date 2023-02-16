// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/veraison/services/builtin"
	"github.com/veraison/services/config"
	"github.com/veraison/services/handler"
	"github.com/veraison/services/log"
	"github.com/veraison/services/plugin"
	"github.com/veraison/services/provisioning/api"
	"github.com/veraison/services/trustedservices"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	DefaultListenAddr = "localhost:8888"
)

type cfg struct {
	ListenAddr string `mapstructure:"listen-addr" valid:"dialstring"`
}

func main() {
	v, err := config.ReadRawConfig("", false)
	if err != nil {
		log.Fatalf("Could not read config sources: %v", err)
	}

	cfg := cfg{
		ListenAddr: DefaultListenAddr,
	}

	subs, err := config.GetSubs(v, "provisioning", "plugin", "*vts", "*logging")
	if err != nil {
		log.Fatal(err)
	}

	classifiers := map[string]interface{}{"service": "provisioning"}
	if err := log.Init(subs["logging"], classifiers); err != nil {
		log.Fatalf("could not configure logging: %v", err)
	}
	log.InitGinWriter() // route gin output to our logger.

	log.Infow("Initializing Provisioning Service", "version", config.Version)

	loader := config.NewLoader(&cfg)
	if err = loader.LoadFromViper(subs["provisioning"]); err != nil {
		log.Fatalf("Could not load config: %v", err)
	}

	log.Info("loading attestation schemes")
	var pluginManager plugin.IManager[handler.IEndorsementHandler]

	if config.SchemeLoader == "plugins" { // nolint:gocritic
		pluginManager, err = plugin.CreateGoPluginManager(
			subs["plugin"], log.Named("plugin"),
			"endorsement-handler", handler.EndorsementHandlerRPC)
		if err != nil {
			log.Fatalf("plugin manager initialization failed: %v", err)
		}
	} else if config.SchemeLoader == "builtin" {
		pluginManager, err = builtin.CreateBuiltinManager[handler.IEndorsementHandler](
			subs["plugin"], log.Named("builtin"), "endorsement-handler")
		if err != nil {
			log.Fatalf("scheme manager initialization failed: %v", err)
		}
	} else {
		log.Panicw("invalid SchemeLoader value", "SchemeLoader", config.SchemeLoader)
	}

	log.Info("Registered media types:")
	for _, mt := range pluginManager.GetRegisteredMediaTypes() {
		log.Info("\t", mt)
	}

	log.Info("initializing VTS client")
	vtsClient := trustedservices.NewGRPCClient()
	if err := vtsClient.Init(subs["vts"]); err != nil {
		log.Fatalf("Could not initilize VTS client: %v", err)
	}

	vtsState, err := vtsClient.GetServiceState(context.TODO(), &emptypb.Empty{})
	if err == nil {
		log.Infow("vts connection established", "server-version", vtsState.ServerVersion)
	} else {
		log.Warnw("Could not connect to VTS server. If you do not expect the server to be running yet, this is probably OK, otherwise it may indicate an issue with your vts.server-addr in your settings",
			"error", err)
	}

	log.Infow("initializing provisioning API service", "address", cfg.ListenAddr)
	apiHandler := api.NewHandler(pluginManager, vtsClient, log.Named("api"))
	go apiServer(apiHandler, cfg.ListenAddr)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan bool, 1)
	go terminator(sigs, done, pluginManager)
	<-done
	log.Info("bye!")
}

func terminator(
	sigs chan os.Signal,
	done chan bool,
	pluginManager plugin.IManager[handler.IEndorsementHandler],
) {
	sig := <-sigs

	log.Info(sig, "received, exiting")

	log.Info("stopping the plugin manager")
	if err := pluginManager.Close(); err != nil {
		log.Error("plugin manager termination failed:", err)
	}

	done <- true
}

func apiServer(apiHandler api.IHandler, listenAddr string) {
	if err := api.NewRouter(apiHandler).Run(listenAddr); err != nil {
		log.Fatalf("Gin engine failed: %v", err)
	}
}
