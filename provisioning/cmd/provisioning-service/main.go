// Copyright 2022-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/veraison/services/auth"
	"github.com/veraison/services/config"
	"github.com/veraison/services/log"
	"github.com/veraison/services/proto"
	"github.com/veraison/services/provisioning/api"
	"github.com/veraison/services/provisioning/provisioner"
	"github.com/veraison/services/vtsclient"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	DefaultListenAddr = "localhost:9443"
)

type cfg struct {
	ListenAddr string `mapstructure:"listen-addr" valid:"dialstring"`
	Protocol   string `mapstructure:"protocol" valid:"in(http|https)"`
	Cert       string `mapstructure:"cert" config:"zerodefault"`
	CertKey    string `mapstructure:"cert-key" config:"zerodefault"`
}

func (o cfg) Validate() error {
	if o.Protocol == "https" && (o.Cert == "" || o.CertKey == "") {
		return errors.New(`both cert and cert-key must be specified when protocol is "https"`)
	}

	return nil
}

func main() {
	config.CmdLine()

	v, err := config.ReadRawConfig(*config.File, false)
	if err != nil {
		log.Fatalf("Could not read config sources: %v", err)
	}

	cfg := cfg{
		ListenAddr: DefaultListenAddr,
		Protocol:   "https",
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
	if err := vtsClient.Init(subs["vts"], cfg.Cert, cfg.CertKey); err != nil {
		log.Fatalf("Could not initilize VTS client: %v", err)
	}

	vtsState, err := vtsClient.GetServiceState(context.TODO(), &emptypb.Empty{})
	if err == nil {
		if vtsState.Status == proto.ServiceStatus_SERVICE_STATUS_READY {
			log.Infow("vts connection established", "server-version",
				vtsState.ServerVersion)
		} else {
			log.Warnw("VTS server not ready. If you do not expect the server to be running yet, this is probably OK, otherwise it may indicate an issue with your vts.server-addr in your settings",
				"server-state", vtsState.Status.String())
		}
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

	if cfg.Protocol == "https" {
		go apiServerTLS(apiHandler, authorizer, cfg.ListenAddr, cfg.Cert, cfg.CertKey)
	} else {
		go apiServer(apiHandler, authorizer, cfg.ListenAddr)
	}

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
	log.Infow("initializing provisioning API HTTP service", "address", listenAddr)

	if err := api.NewRouter(apiHandler, authorizer).Run(listenAddr); err != nil {
		log.Fatalf("Gin engine failed: %v", err)
	}
}

func apiServerTLS(
	apiHandler api.IHandler,
	authorizer auth.IAuthorizer,
	listenAddr, certFile, keyFile string,
) {
	log.Infow("initializing provisioning API HTTPS service", "address", listenAddr)

	err := api.NewRouter(apiHandler, authorizer).RunTLS(listenAddr, certFile, keyFile)
	if err != nil {
		log.Fatalf("Gin engine failed: %v", err)
	}
}
