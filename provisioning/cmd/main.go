// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/veraison/services/config"
	"github.com/veraison/services/provisioning/api"
	"github.com/veraison/services/provisioning/decoder"
	"github.com/veraison/services/vtsclient"
)

// TODO(tho) make these configurable
var (
	DefaultPluginDir  = "../plugins/bin/"
	DefaultListenAddr = "localhost:8888"
)

func main() {
	cfg := config.NewYAMLReader()
	if _, err := cfg.ReadFile("config.yaml"); err != nil {
		log.Fatalf("counl not read config file: %v", err)
	}

	provCfg := cfg.MustGetStore("provisioning")

	pluginDir, err := config.GetString(provCfg, "plugin-dir", &DefaultPluginDir)
	if err != nil {
		log.Fatalf("could not get plugin-dir: %v", err)
	}

	listenAddr, err := config.GetString(provCfg, "list-addr", &DefaultListenAddr)
	if err != nil {
		log.Fatalf("could not get listen-addr: %v", err)
	}

	pluginManager := NewGoPluginManager(pluginDir)
	vtsClient := vtsclient.NewGRPC(cfg.MustGetStore("vts-grpc"))
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
