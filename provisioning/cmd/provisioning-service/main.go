// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/veraison/services/config"
	"github.com/veraison/services/log"
	"github.com/veraison/services/provisioning/api"
	"github.com/veraison/services/provisioning/decoder"
	"github.com/veraison/services/vtsclient"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	DefaultPluginDir  = "../../plugins/bin/"
	DefaultListenAddr = "localhost:8888"
)

type cfg struct {
	PluginDir  string `mapstructure:"plugin-dir"`
	ListenAddr string `mapstructure:"listen-addr" valid:"dialstring"`
}

func (o cfg) Validate() error {
	if _, err := os.Stat(o.PluginDir); err != nil {
		return fmt.Errorf("could not stat PluginDir: %w", err)
	}

	return nil
}

func main() {
	v, err := config.ReadRawConfig("", false)
	if err != nil {
		log.Fatalf("Could not read config sources: %v", err)
	}

	cfg := cfg{
		PluginDir:  DefaultPluginDir,
		ListenAddr: DefaultListenAddr,
	}

	subs, err := config.GetSubs(v, "provisioning", "*vts", "*logging")
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

	log.Info("loading plugins")
	pluginManager := NewGoPluginManager(cfg.PluginDir, log.Named("plugin"))

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
	pluginManager decoder.IDecoderManager,
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

func NewGoPluginManager(dir string, logger *zap.SugaredLogger) decoder.IDecoderManager {
	mgr := &decoder.GoPluginDecoderManager{}
	err := mgr.Init(dir, logger)
	if err != nil {
		logger.Fatalf("plugin initialisation failed: %v", err)
	}

	return mgr
}
