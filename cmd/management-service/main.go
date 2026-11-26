// Copyright 2023-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/mattn/go-sqlite3"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/veraison/services/pkg/api/common"
	"github.com/veraison/services/pkg/api/common/auth"
	"github.com/veraison/services/pkg/api/management"
	"github.com/veraison/services/pkg/config"
	"github.com/veraison/services/pkg/log"
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
	if err := execute(); err != nil {
		log.Fatal(err)
	}
}

func execute() error {
	log.Infow("initializing Management Service", "version", config.Version)
	config.CmdLine()

	v, err := config.ReadRawConfigFromPath(config.File(), false)
	if err != nil {
		return fmt.Errorf("could not read config: %w", err)
	}

	subs, err := config.GetSubs(v, "*management", "*logging", "*auth")
	if err != nil {
		return fmt.Errorf("could not parse config: %w", err)
	}

	if err := initLogging(subs["logging"]); err != nil {
		return err
	}
	defer log.Info("bye!")

	log.Info("initializing policy manager")
	policyManager, err := management.CreatePolicyManagerFromConfig(v, "policy")
	if err != nil {
		return fmt.Errorf("could not init policy manager: %w", err)
	}
	defer policyManager.Close()

	cfg := cfg{
		ListenAddr:     DefaultListenAddr,
		Protocol:       "https",
		Cert:           "[unset]",
		CertKey:        "[unset]",
		MaxPayloadSize: "1MB",
	}
	loader := config.NewLoader(&cfg)
	if err := loader.LoadFromViper(subs["management"]); err != nil {
		return fmt.Errorf("could not load verification config: %w", err)
	}

	maxPayloadSize, err := config.ParseBytesSize(cfg.MaxPayloadSize)
	if err != nil {
		return fmt.Errorf("could not parse max-payload-size: %w", err)
	}

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

	apiHandler := management.NewHandler(policyManager, log.Named("api"), maxPayloadSize)

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

	classifiers := map[string]interface{}{"service": "management"}
	if err := log.Init(v, classifiers); err != nil {
		return fmt.Errorf("could not configure logging: %w", err)
	}

	// route gin output to our logger.
	log.InitGinWriter()

	return nil
}

func serve(
	apiHandler *management.Handler,
	authorizer auth.IAuthorizer,
	listenAddr string,
	errChan chan error,
) {
	log.Infof("management API HTTP service listening on %s", listenAddr)
	router := common.NewRouter()
	management.SetupRouter(router, apiHandler, authorizer)
	errChan <- router.Run(listenAddr)
}

func serveTLS(
	apiHandler *management.Handler,
	authorizer auth.IAuthorizer,
	listenAddr, certFile, keyFile string,
	errChan chan error,
) {
	log.Infof("management API HTTPS service listening on %s", listenAddr)
	router := common.NewRouter()
	management.SetupRouter(router, apiHandler, authorizer)
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
