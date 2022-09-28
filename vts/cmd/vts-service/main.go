// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/viper"

	"github.com/veraison/services/kvstore"
	"github.com/veraison/services/policy"
	"github.com/veraison/services/vts/pluginmanager"
	"github.com/veraison/services/vts/policymanager"
	"github.com/veraison/services/vts/trustedservices"
)

func main() {
	v := viper.New()

	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	v.AddConfigPath(wd)
	v.SetConfigType("yaml")
	v.SetConfigName("config")

	if err := v.ReadInConfig(); err != nil {
		log.Fatalf("could not read config: %v", err)
	}

	taStore, err := kvstore.New(v.Sub("ta-store"))
	if err != nil {
		log.Fatalf("trust anchor store initialisation failed: %v", err)
	}

	enStore, err := kvstore.New(v.Sub("en-store"))
	if err != nil {
		log.Fatalf("endorsement store initialization failed: %v", err)
	}

	poStore, err := policy.NewStore(v.Sub("po-store"))
	if err != nil {
		log.Fatalf("policy store initialization failed: %v", err)
	}

	policyManager, err := policymanager.New(v.Sub("po-agent"), poStore)
	if err != nil {
		log.Fatalf("policy manager initialization failed: %v", err)
	}

	pluginManager := pluginmanager.New(v.Sub("plugin"))
	if err := pluginManager.Init(); err != nil {
		log.Fatalf("plugin manager initialization failed: %v", err)
	}

	// from this point onwards taStore, enStore and pluginManager are owned by vts
	vts := trustedservices.NewGRPC(v.Sub("vts-grpc"), taStore, enStore,
		pluginManager, policyManager)

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
