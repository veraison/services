// Copyright 2022-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/veraison/services/pkg/api/common"
	"github.com/veraison/services/pkg/api/common/auth"
	"github.com/veraison/services/pkg/api/provisioning"
	"github.com/veraison/services/pkg/config"
	"github.com/veraison/services/pkg/log"
	"github.com/veraison/services/pkg/vts/client"
)

const DefaultPort = 9443

func main() {
	if err := execute(); err != nil {
		log.Fatal(err)
	}
}

func execute() error {
	log.Infow("initializing Provisioning Service", "version", config.Version)
	config.CmdLine()

	log.Infof("reading configuration from %s", config.File())
	subs, err := config.ReadSubs("*provisioning", "vts", "*logging", "*auth")
	if err != nil {
		return err
	}

	if err := initLogging(subs["logging"]); err != nil {
		return err
	}
	defer log.Info("bye!")

	cfg, err := common.CreateConfig(subs["provisioning"], subs["vts"], DefaultPort)
	if err != nil {
		return fmt.Errorf("could not create configuration: %w", err)
	}
	log.Debugf("service config: %+v", cfg)

	ctx := context.Background()

	log.Info("initializing VTS client")
	vtsClient := client.NewTrustedService(cfg.VTS(), log.Named("vts"))

	vtsState, err := vtsClient.GetServiceState(ctx)
	if err == nil {
		log.Infow("vts connection established", "server-version", vtsState.Version)
	} else {
		log.Warn("could not connect to VTS server")
		log.Warnf("(%v)", err)
		log.Warn("if you do not expect the server to be running yet, this is probably OK")
		log.Warn("otherwise it may indicate an issue with vts.server-addr in your settings")
	}

	log.Info("initializing provisioner")
	provisioner := provisioning.NewProvisioner(ctx, vtsClient)

	log.Infow("initializing provisioning API service", "address", cfg.ListenAddr)
	authorizer, err := auth.NewAuthorizer(subs["auth"], log.Named("auth"))
	if err != nil {
		return fmt.Errorf("could not init authorizer: %w", err)
	}
	defer func() {
		err := authorizer.Close()
		if err != nil {
			log.Errorf("Could not close authorizer: %v", err)
		}
	}()

	apiHandler := provisioning.NewHandler(
		provisioner,
		log.Named("api"),
		cfg.DiscoveryMaxAge,
		cfg.MaxPayloadSize,
	)

	gin.SetMode(gin.ReleaseMode)

	done := make(chan error, 1)
	if cfg.Protocol == "https" {
		go serveTLS(apiHandler, authorizer, cfg.ListenAddr, cfg.Cert, cfg.CertKey, done)
	} else {
		go serve(apiHandler, authorizer, cfg.ListenAddr, done)
	}

	go trapSignals(done)
	return <-done
}

func initLogging(v *viper.Viper) error {
	log.Info("initializing logging")

	classifiers := map[string]any{"service": "provisioning"}
	if err := log.Init(v, classifiers); err != nil {
		return fmt.Errorf("could not configure logging: %w", err)
	}

	// route gin output to our logger.
	log.InitGinWriter()

	return nil
}

func serve(
	apiHandler provisioning.IHandler,
	authorizer auth.IAuthorizer,
	listenAddr string,
	errChan chan error,
) {
	log.Infof("provisioning API HTTP service listening on %s", listenAddr)
	router := common.NewRouter()
	provisioning.SetupRouter(router, apiHandler, authorizer)
	errChan <- router.Run(listenAddr)
}

func serveTLS(
	apiHandler provisioning.IHandler,
	authorizer auth.IAuthorizer,
	listenAddr, certFile, keyFile string,
	errChan chan error,
) {
	log.Infof("provisioning API HTTPS service listening on %s", listenAddr)
	router := common.NewRouter()
	provisioning.SetupRouter(router, apiHandler, authorizer)
	errChan <- router.RunTLS(listenAddr, certFile, keyFile)
}

func trapSignals(
	done chan error,
) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigs
	log.Info(sig, " received, exiting")
	done <- nil
}
