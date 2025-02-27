// Copyright 2022-2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"context"
	"errors"

	"github.com/veraison/services/config"
	"github.com/veraison/services/log"
	"github.com/veraison/services/proto"
	"github.com/veraison/services/verification/api"
	"github.com/veraison/services/verification/sessionmanager"
	"github.com/veraison/services/verification/verifier"
	"github.com/veraison/services/vtsclient"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	DefaultListenAddr = "localhost:8443"
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
		log.Fatalf("Could not read config: %v", err)
	}
	cfg := cfg{
		ListenAddr: DefaultListenAddr,
		Protocol: "https",
	}

	subs, err := config.GetSubs(v, "*vts", "*verifier", "*verification", "*logging",
				"*sessionmanager")
	if err != nil {
		log.Fatalf("Could not read config: %v", err)
	}

	classifiers := map[string]interface{}{"service": "verification"}
	if err := log.Init(subs["logging"], classifiers); err != nil {
		log.Fatalf("could not configure logging: %v", err)
	}
	log.InitGinWriter() // route gin output to our logger.

	log.Infow("Initializing Verification Service", "version", config.Version)

	loader := config.NewLoader(&cfg)
	if err := loader.LoadFromViper(subs["verification"]); err != nil {
		log.Fatalf("Could not load verification config: %v", err)
	}

	log.Info("initializing session manager")
	sessionManager, err := sessionmanager.New(subs["sessionmanager"])
	if err != nil {
		log.Fatalf("Could not create session manager: %v", err)
	}

	log.Info("initializing VTS client")
	vtsClient := vtsclient.NewGRPC()
	if err := vtsClient.Init(subs["vts"], cfg.Cert, cfg.CertKey); err != nil {
		log.Fatalf("Could not initialize VTS client: %v", err)
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
		log.Warnw("Could not connect to VTS server. If you do not expect the server to be running yet, this is probably OK, otherwise it may indicate an issue with vts.server-addr in your settings",
			"error", err)
	}

	log.Info("initializing verifier")
	verifier := verifier.New(subs["verifier"], vtsClient)

	apiHandler := api.NewHandler(sessionManager, verifier)

	if cfg.Protocol == "https" {
		apiServerTLS(apiHandler, cfg.ListenAddr, cfg.Cert, cfg.CertKey)
	} else {
		apiServer(apiHandler, cfg.ListenAddr)
	}
}

func apiServer(apiHandler api.IHandler, listenAddr string) {
	log.Infow("initializing verification API HTTP service", "address", listenAddr)

	if err := api.NewRouter(apiHandler).Run(listenAddr); err != nil {
		log.Fatalf("Gin engine failed: %v", err)
	}
}

func apiServerTLS(apiHandler api.IHandler, listenAddr, certFile, keyFile string) {
	log.Infow("initializing verification API HTTPS service", "address", listenAddr)

	if err := api.NewRouter(apiHandler).RunTLS(listenAddr, certFile, keyFile); err != nil {
		log.Fatalf("Gin engine failed: %v", err)
	}
}
