// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/veraison/services/config"
	"github.com/veraison/services/provisioning/api"
	"github.com/veraison/services/provisioning/decoder"
	"github.com/veraison/services/vtsclient"
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

	subs, err := config.GetSubs(v, "provisioning", "*vts")
	if err != nil {
		log.Fatal(err)
	}

	loader := config.NewLoader(&cfg)
	if err = loader.LoadFromViper(subs["provisioning"]); err != nil {
		log.Fatalf("Could not load config: %v", err)
	}

	pluginManager := NewGoPluginManager(cfg.PluginDir)

	vtsClient := vtsclient.NewGRPC()
	if err := vtsClient.Init(subs["vts"]); err != nil {
		log.Fatalf("Could not initilize VTS client: %v", err)
	}

	apiHandler := api.NewHandler(pluginManager, vtsClient)
	go apiServer(apiHandler, cfg.ListenAddr)

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
