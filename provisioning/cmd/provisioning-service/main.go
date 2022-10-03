// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/setrofim/viper"
	"github.com/veraison/services/provisioning/api"
	"github.com/veraison/services/provisioning/decoder"
	"github.com/veraison/services/vtsclient"
)

// TODO(tho) make these configurable
var (
	DefaultPluginDir  = "../plugins/bin/"
	DefaultListenAddr = "localhost:8888"
)

func initConfig() (*viper.Viper, error) {
	v := viper.New()

	v.SetConfigType("yaml")
	v.SetConfigName("config")

	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	v.AddConfigPath(wd)

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	cfg := v.Sub("provisioning")
	if cfg == nil {
		return nil, errors.New(`"provisioning" section not found in config`)
	}

	cfg.SetDefault("plugin-dir", DefaultPluginDir)
	cfg.SetDefault("list-addr", DefaultListenAddr)

	return cfg, nil
}

func main() {
	cfg, err := initConfig()
	if err != nil {
		log.Fatalf("could not read config: %v", err)
	}

	pluginDir := cfg.GetString("plugin-dir")
	listenAddr := cfg.GetString("list-addr")

	pluginManager := NewGoPluginManager(pluginDir)
	vtsClient := vtsclient.NewGRPC(cfg.Sub("vts-grpc"))
	apiHandler := api.NewHandler(pluginManager, vtsClient)
	go apiServer(apiHandler, listenAddr)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan bool, 1)
	go terminator(sigs, done, pluginManager)
	<-done
	log.Println("bye!")
}

func terminator(
	sigs chan os.Signal,
	done chan bool,
	pluginManager decoder.IDecoderManager,
) {
	sig := <-sigs

	log.Println(sig, "received, exiting")

	log.Println("stopping the plugin manager")
	if err := pluginManager.Close(); err != nil {
		log.Println("plugin manager termination failed:", err)
	}

	done <- true
}

func apiServer(apiHandler api.IHandler, listenAddr string) {
	if err := api.NewRouter(apiHandler).Run(listenAddr); err != nil {
		log.Fatalf("Gin engine failed: %v", err)
	}
}

func NewGoPluginManager(dir string) decoder.IDecoderManager {
	mgr := &decoder.GoPluginDecoderManager{}
	err := mgr.Init(dir)
	if err != nil {
		log.Fatalf("plugin initialisation failed: %v", err)
	}

	return mgr
}
