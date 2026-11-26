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
	"github.com/veraison/services/pkg/api/verification"
	"github.com/veraison/services/pkg/api/verification/session"
	"github.com/veraison/services/pkg/config"
	"github.com/veraison/services/pkg/log"
	"github.com/veraison/services/pkg/vts/client"
)

const DefaultPort = 8443

func main() {
	if err := execute(); err != nil {
		log.Fatal(err)
	}
}

func execute() error {
	log.Infow("initializing Verification Service", "version", config.Version)
	config.CmdLine()

	log.Infof("reading configuration from %s", config.File())
	subs, err := config.ReadSubs("*verification", "vts", "*sessionmanager", "*logging", "*auth")
	if err != nil {
		return err
	}

	if err := initLogging(subs["logging"]); err != nil {
		return err
	}
	defer log.Info("bye!")

	cfg, err := common.CreateConfig(subs["verification"], subs["vts"], DefaultPort)
	if err != nil {
		return fmt.Errorf("could not create configuration: %w", err)
	}
	log.Debugf("service config: %+v", cfg)

	log.Info("initializing session manager")
	sessionManager, err := session.CreateManager(subs["sessionmanager"])
	if err != nil {
		return fmt.Errorf("could not create session manager: %w", err)
	}

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

	log.Info("initializing verifier")
	verifier := verification.NewVerifier(ctx, vtsClient)

	apiHandler := verification.NewHandler(
		sessionManager,
		verifier,
		cfg.DiscoveryMaxAge,
		cfg.MaxPayloadSize,
	)

	gin.SetMode(gin.ReleaseMode)

	done := make(chan error, 1)
	if cfg.Protocol == "https" {
		go serveTLS(apiHandler, cfg.ListenAddr, cfg.Cert, cfg.CertKey, done)
	} else {
		go serve(apiHandler, cfg.ListenAddr, done)
	}

	go trapSignals(done)
	return <-done
}

func initLogging(v *viper.Viper) error {
	log.Info("initializing logging")

	classifiers := map[string]interface{}{"service": "verification"}
	if err := log.Init(v, classifiers); err != nil {
		return fmt.Errorf("could not configure logging: %w", err)
	}

	// route gin output to our logger.
	log.InitGinWriter()

	return nil
}

func serve(
	apiHandler verification.IHandler,
	listenAddr string,
	errChan chan error,
) {
	log.Infof("verification API HTTP service listening on %s", listenAddr)
	router := common.NewRouter()
	verification.SetupRouter(router, apiHandler)
	errChan <- router.Run(listenAddr)
}

func serveTLS(
	apiHandler verification.IHandler,
	listenAddr, certFile, keyFile string,
	errChan chan error,
) {
	log.Infof("verification API HTTPS service listening on %s", listenAddr)
	router := common.NewRouter()
	verification.SetupRouter(router, apiHandler)
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
