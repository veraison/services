// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/veraison/services/auth"
	"github.com/veraison/services/config"
	"github.com/veraison/services/log"
	"github.com/veraison/services/provisioning/api"
	"github.com/veraison/services/provisioning/provisioner"
	"github.com/veraison/services/vtsclient"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	DefaultListenAddr = "localhost:8888"
)

type cfg struct {
	ListenAddr string `mapstructure:"listen-addr" valid:"dialstring"`
}

func main() {
	config.CmdLine()

	v, err := config.ReadRawConfig(*config.File, false)
	if err != nil {
		log.Fatalf("Could not read config sources: %v", err)
	}

	cfg := cfg{
		ListenAddr: DefaultListenAddr,
	}

	subs, err := config.GetSubs(v, "provisioning", "vts", "*logging", "*auth")
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

	log.Info("initializing VTS client")
	vtsClient := vtsclient.NewGRPC()
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

	log.Info("initializing provisioner")
	provisioner := provisioner.New(vtsClient)

	log.Infow("initializing provisioning API service", "address", cfg.ListenAddr)
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

	apiHandler := api.NewHandler(provisioner, log.Named("api"))
	go apiServer(apiHandler, authorizer, cfg.ListenAddr)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan bool, 1)
	go terminator(sigs, done)
	<-done
	log.Info("bye!")
}

func terminator(
	sigs chan os.Signal,
	done chan bool,
) {
	sig := <-sigs

	log.Info(sig, "received, exiting")

	done <- true
}

func apiServer(apiHandler api.IHandler, authorizer auth.IAuthorizer, listenAddr string) {
	if err := api.NewRouter(apiHandler, authorizer).Run(listenAddr); err != nil {
		log.Fatalf("Gin engine failed: %v", err)
	}
}
