// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/mattn/go-sqlite3"

	"github.com/veraison/services/config"
	"github.com/veraison/services/kvstore"
	"github.com/veraison/services/vts/pluginmanager"
	"github.com/veraison/services/vts/trustedservices"
)

var (
	PluginDir  = "../plugins/bin/"
	TaStoreCfg = config.Store{
		"backend":        "sql",
		"sql.driver":     "sqlite3",
		"sql.datasource": "ta-store.sql",
	}
	EnStoreCfg = config.Store{
		"backend":        "sql",
		"sql.driver":     "sqlite3",
		"sql.datasource": "en-store.sql",
	}
	VtsGrpcConfig = config.Store{
		"server.addr": "127.0.0.1:50051",
	}
)

func main() {
	taStore, err := kvstore.New(TaStoreCfg)
	if err != nil {
		log.Fatalf("trust anchor store initialisation failed: %v", err)
	}

	enStore, err := kvstore.New(EnStoreCfg)
	if err != nil {
		log.Fatalf("endorsement store initialization failed: %v", err)
	}

	pluginManager := NewGoPluginManager(PluginDir)

	vts := trustedservices.NewGRPC(VtsGrpcConfig, taStore, enStore, pluginManager)

	go vts.Run()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan bool, 1)
	go terminator(sigs, done, vts)
	<-done
	log.Println("bye!")
}

func NewGoPluginManager(dir string) *pluginmanager.GoPluginManager {
	mgr := &pluginmanager.GoPluginManager{}
	err := mgr.Init(dir)
	if err != nil {
		log.Fatalf("plugin initialisation failed: %v", err)
	}

	return mgr
}

func terminator(sigs chan os.Signal, done chan bool, vts trustedservices.ITrustedServices) {
	sig := <-sigs

	log.Println(sig, "received, exiting")

	log.Println("stopping VTS")
	if err := vts.Close(); err != nil {
		log.Println("VTS termination failed:", err)
	}

	done <- true
}
