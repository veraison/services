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

var (
	DefaultPluginDir  = "../../plugins/bin/"
	DefaultListenAddr = "localhost:8888"
)

type config struct {
	Provisioning struct {
		PluginDir  string `mapstructure:"plugin-dir"`
		ListenAddr string `mapstructure:"listen-addr"`
	}
	VtsGRPC *viper.Viper
}

func initConfig() (*config, error) {
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

	v.SetDefault("provisioning.plugin-dir", DefaultPluginDir)
	v.SetDefault("provisioning.listen-addr", DefaultListenAddr)

	var cfg config

	if err = v.UnmarshalKey("provisioning", &cfg.Provisioning); err != nil {
		return nil, err
	}

	if cfg.VtsGRPC = v.Sub("vts-grpc"); cfg.VtsGRPC == nil {
		return nil, errors.New(`"vts-grpc" section not found in config.`)
	}

	return &cfg, nil
}

func main() {
	cfg, err := initConfig()
	if err != nil {
		log.Fatalf("could not load config: %v", err)
	}

	pluginDir := cfg.Provisioning.PluginDir
	listenAddr := cfg.Provisioning.ListenAddr

	pluginManager := NewGoPluginManager(pluginDir)
	vtsClient := vtsclient.NewGRPC(cfg.VtsGRPC)
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
