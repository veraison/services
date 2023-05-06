// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"os"
	"os/signal"
	"syscall"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/afero"

	"github.com/veraison/services/builtin"
	"github.com/veraison/services/config"
	"github.com/veraison/services/handler"
	"github.com/veraison/services/kvstore"
	"github.com/veraison/services/log"
	"github.com/veraison/services/plugin"
	"github.com/veraison/services/policy"
	"github.com/veraison/services/vts/earsigner"
	"github.com/veraison/services/vts/policymanager"
	"github.com/veraison/services/vts/trustedservices"
)

func main() {
	config.CmdLine()

	v, err := config.ReadRawConfig(*config.File, false)
	if err != nil {
		log.Fatalf("could not read config: %v", err)
	}

	subs, err := config.GetSubs(v, "ta-store", "en-store", "po-store",
		"*po-agent", "plugin", "*vts", "ear-signer", "*logging")
	if err != nil {
		log.Fatal(err)
	}

	classifiers := map[string]interface{}{"service": "vts"}
	if err := log.Init(subs["logging"], classifiers); err != nil {
		log.Fatalf("could not configure logging: %v", err)
	}

	log.Info("initializing stores")
	taStore, err := kvstore.New(subs["ta-store"], log.Named("ta-store"))
	if err != nil {
		log.Fatalf("trust anchor store initialisation failed: %v", err)
	}

	enStore, err := kvstore.New(subs["en-store"], log.Named("en-store"))
	if err != nil {
		log.Fatalf("endorsement store initialization failed: %v", err)
	}

	poStore, err := policy.NewStore(subs["po-store"], log.Named("po-store"))
	if err != nil {
		log.Fatalf("policy store initialization failed: %v", err)
	}

	log.Info("initializing policy manager")
	policyManager, err := policymanager.New(subs["po-agent"], poStore, log.Named("policy"))
	if err != nil {
		log.Fatalf("policy manager initialization failed: %v", err)
	}

	log.Info("loading attestation schemes")
	var pluginManager plugin.IManager[handler.IEvidenceHandler]

	if config.SchemeLoader == "plugins" { // nolint:gocritic
		pluginManager, err = plugin.CreateGoPluginManager(
			subs["plugin"], log.Named("plugin"),
			"evidence-handler", handler.EvidenceHandlerRPC)
		if err != nil {
			log.Fatalf("plugin manager initialization failed: %v", err)
		}
	} else if config.SchemeLoader == "builtin" {
		pluginManager, err = builtin.CreateBuiltinManager[handler.IEvidenceHandler](
			subs["plugin"], log.Named("builtin"), "evidence-handler")
		if err != nil {
			log.Fatalf("scheme manager initialization failed: %v", err)
		}
	} else {
		log.Panicw("invalid SchemeLoader value", "SchemeLoader", config.SchemeLoader)
	}

	log.Info("Registered media types:")
	for _, mt := range pluginManager.GetRegisteredMediaTypes() {
		log.Info("\t", mt)
	}

	log.Info("loading EAR signer")
	earSigner, err := earsigner.New(subs["ear-signer"], afero.NewOsFs())
	if err != nil {
		log.Fatalf("EAR signer initialization failed: %v", err)
	}

	log.Info("initializing service")
	// from this point onwards taStore, enStore, pluginManager, policyManager
	// and earSigner are owned by vts
	vts := trustedservices.NewGRPC(taStore, enStore,
		pluginManager, policyManager, earSigner, log.Named("vts"))

	if err = vts.Init(subs["vts"], pluginManager); err != nil {
		log.Fatalf("VTS initialisation failed: %v", err)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan bool, 1)

	go vtsRun(vts, done)
	go sigWaiter(sigs, done)

	<-done

	log.Info("stopping service")
	if err := vts.Close(); err != nil {
		log.Error("service termination failed:", err)
	}
	log.Info("bye!")
}

func vtsRun(vts trustedservices.ITrustedServices, done chan bool) {
	if err := vts.Run(); err != nil {
		log.Error("VTS failed:", err)
	}

	done <- true
}

func sigWaiter(sigs chan os.Signal, done chan bool) {
	sig := <-sigs

	log.Info(sig, " received, exiting")

	done <- true
}
