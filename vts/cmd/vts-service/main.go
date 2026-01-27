// Copyright 2022-2026 Contributors to the Veraison project.
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
	"github.com/veraison/services/log"
	"github.com/veraison/services/plugin"
	"github.com/veraison/services/policy"
	"github.com/veraison/services/vts/coservsigner"
	"github.com/veraison/services/vts/earsigner"
	"github.com/veraison/services/vts/policymanager"
	"github.com/veraison/services/vts/store"
	"github.com/veraison/services/vts/trustedservices"
)

func main() {
	config.CmdLine()

	v, err := config.ReadRawConfig(*config.File, false)
	if err != nil {
		log.Fatalf("could not read config: %v", err)
	}

	subs, err := config.GetSubs(v, "store", "po-store",
		"*po-agent", "plugin", "*vts", "ear-signer", "*coserv-signer", "*logging")
	if err != nil {
		log.Fatal(err)
	}

	classifiers := map[string]any{"service": "vts"}
	if err := log.Init(subs["logging"], classifiers); err != nil {
		log.Fatalf("could not configure logging: %v", err)
	}

	log.Info("initializing stores")
	enStore, err := store.New(subs["store"], log.Named("store"))
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
	var schemePluginManager plugin.IManager[handler.ISchemeHandler]
	var coservProxyPluginManager plugin.IManager[handler.ICoservProxyHandler]

	psubs, err := config.GetSubs(subs["plugin"], "*go-plugin", "*builtin")
	if err != nil {
		log.Fatalf("could not get subs: %v", err)
	}

	switch config.SchemeLoader {
	case "plugins":
		loader, err := plugin.CreateGoPluginLoader(
			psubs["go-plugin"].AllSettings(),
			log.Named("plugin"))
		if err != nil {
			log.Fatalf("could not create plugin loader: %v", err)
		}

		schemePluginManager, err = plugin.CreateGoPluginManagerWithLoader(
			loader,
			"scheme-handler",
			log.Named("plugin"),
			handler.SchemeHandlerRPC)
		if err != nil {
			log.Fatalf("could not create store PluginManagerWithLoader: %v", err)
		}
		coservProxyPluginManager, err = plugin.CreateGoPluginManagerWithLoader(
			loader,
			"coserv-proxy-handler",
			log.Named("plugin"),
			handler.CoservProxyHandlerRPC)
		if err != nil {
			log.Fatalf("could not create coserv PluginManagerWithLoader: %v", err)
		}
	case "builtin":
		loader, err := builtin.CreateBuiltinLoader(
			psubs["builtin"].AllSettings(),
			log.Named("builtin"))
		if err != nil {
			log.Fatalf("could not create builtin loader: %v", err)
		}
		schemePluginManager, err = builtin.CreateBuiltinManagerWithLoader[handler.ISchemeHandler](
			loader, log.Named("builtin"),
			"scheme-handler")
		if err != nil {
			log.Fatalf("could not create store BuiltinManagerWithLoader: %v", err)
		}
		coservProxyPluginManager, err = builtin.CreateBuiltinManagerWithLoader[handler.ICoservProxyHandler](
			loader, log.Named("builtin"),
			"coserv-handler")
		if err != nil {
			log.Fatalf("could not create coserv BuiltinManagerWithLoader: %v", err)
		}
	default:
		log.Panicw("invalid SchemeLoader value", "SchemeLoader", config.SchemeLoader)
	}

	log.Info("Provisioning media types:")
	for _, mt := range schemePluginManager.GetRegisteredMediaTypesByCategory("provisioning") {
		log.Info("\t", mt)
	}

	log.Info("Verification media types:")
	for _, mt := range schemePluginManager.GetRegisteredMediaTypesByCategory("verification") {
		log.Info("\t", mt)
	}

	log.Info("CoSERV Proxy media types:")
	for _, mt := range coservProxyPluginManager.GetRegisteredMediaTypes() {
		log.Info("\t", mt)
	}

	log.Info("loading EAR signer")
	earSigner, err := earsigner.New(subs["ear-signer"], afero.NewOsFs())
	if err != nil {
		log.Fatalf("EAR signer initialization failed: %v", err)
	}

	var coservSigner coservsigner.ICoservSigner

	if subs["coserv-signer"].GetBool("use") {
		log.Info("loading CoSERV signer")
		coservSigner, err = coservsigner.New(subs["coserv-signer"], afero.NewOsFs())
		if err != nil {
			log.Fatalf("CoSERV signer initialization failed: %v", err)
		}

		// CoSERV media types.
		log.Info("TODO CoSERV profile types:")
	}

	log.Info("initializing service")
	// from this point onwards taStore, enStore, evPluginManager,
	// endPluginManager, storePluginManager, coservProxyPluginManager,
	// policyManager and earSigner are owned by vts
	vts := trustedservices.NewGRPC(enStore,
		schemePluginManager, coservProxyPluginManager,
		policyManager, earSigner, coservSigner, log.Named("vts"))

	if err = vts.Init(subs["vts"]); err != nil {
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
