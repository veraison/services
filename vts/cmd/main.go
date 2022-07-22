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
	PluginManagerCfg = config.Store{
		"backend":          "go-plugin",
		"go-plugin.folder": "../plugins/bin/",
	}
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

	pluginManager := pluginmanager.New(PluginManagerCfg)
	if err := pluginManager.Init(); err != nil {
		log.Fatalf("plugin manager initialization failed: %v", err)
	}

	// from this point onwards taStore, enStore and pluginManager are owned by vts
	vts := trustedservices.NewGRPC(VtsGrpcConfig, taStore, enStore, pluginManager)

	err = vts.Init()
	if err != nil {
		log.Fatalf("VTS initialisation failed: %v", err)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan bool, 1)

	go vtsRun(vts, done)
	go sigWaiter(sigs, done)

	<-done

	log.Println("stopping VTS")
	if err := vts.Close(); err != nil {
		log.Println("VTS termination failed:", err)
	}
	log.Println("bye!")
}

func vtsRun(vts trustedservices.ITrustedServices, done chan bool) {
	if err := vts.Run(); err != nil {
		log.Println("VTS failed:", err)
	}

	done <- true
}

func sigWaiter(sigs chan os.Signal, done chan bool) {
	sig := <-sigs

	log.Println(sig, "received, exiting")

	done <- true
}
