// Copyright 2022-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/afero"
	"github.com/spf13/viper"

	"github.com/veraison/services/pkg/config"
	"github.com/veraison/services/pkg/endorsement"
	"github.com/veraison/services/pkg/log"
	"github.com/veraison/services/pkg/plugin"
	"github.com/veraison/services/pkg/plugin/loader"
	"github.com/veraison/services/pkg/policy"
	"github.com/veraison/services/pkg/scheme"
	"github.com/veraison/services/pkg/vts"
	"github.com/veraison/services/pkg/vts/service"
)

func main() {
	if err := execute(); err != nil {
		log.Fatal(err)
	}
}

func execute() error {
	log.Infow("initializing Veraison Trusted Services", "version", config.Version)
	config.CmdLine()

	log.Infof("reading configuration from %s", config.File())
	subs, err := config.ReadSubs(
		"policy",
		"endorsement-store",
		"ear-signer",
		"*plugin",
		"*vts",
		"*logging",
		"*scheme",
	)
	if err != nil {
		return err
	}

	err = initLogging(subs["logging"])
	if err != nil {
		return err
	}
	defer log.Info("bye!")

	earSigner, err := initSigner(subs["ear-signer"])
	if err != nil {
		return err
	}

	pluginLoader, err := initPluginLoader(subs["plugin"])
	if err != nil {
		return err
	}
	defer pluginLoader.Close()

	endorsementStore, err := initEndorsementStore(subs["endorsement-store"], pluginLoader)
	if err != nil {
		return err
	}
	defer endorsementStore.Close()

	schemeManager, err := initAttestationSchemes(subs["scheme"], pluginLoader)
	if err != nil {
		return err
	}
	defer schemeManager.Close()

	log.Info("")
	log.Info("CoRIM profiles:")
	for _, mt := range schemeManager.GetRegisteredCorimProfiles() {
		log.Info("\t", mt)
	}
	log.Info("")

	log.Info("Evidence media types:")
	for _, mt := range schemeManager.GetRegisteredEvidenceMediaTypes() {
		log.Info("\t", mt)
	}
	log.Info("")

	policyManager, err := initPolicyManager(subs["policy"])
	if err != nil {
		return err
	}
	defer policyManager.Close()

	server, err := initServer(subs["vts"], endorsementStore, schemeManager, policyManager, earSigner)
	if err != nil {
		return err
	}
	defer server.Close()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan bool, 1)

	go serve(server, done)
	go trapSignals(sigs, done)
	<-done

	log.Info("stopping service")
	return nil
}

func initLogging(v *viper.Viper) error {
	classifiers := map[string]any{"service": "vts"}
	return log.Init(v, classifiers)
}

func initPluginLoader(v *viper.Viper) (plugin.ILoader, error) {
	log.Info("discovering plugins")
	type pluginConfig struct {
		Locations []string `mapstructure:"dirs"`
	}
	cfg := pluginConfig{Locations: []string{"."}}

	configLoader := config.NewLoader(&cfg)
	if err := configLoader.LoadFromViper(v); err != nil {
		return nil, fmt.Errorf("could not load plugin config: %w", err)
	}

	pluginLoader := loader.NewPluginLoader(log.Named("plugin"))
	for _, location := range cfg.Locations {
		if err := pluginLoader.LoadFrom(location); err != nil {
			return nil, fmt.Errorf("could not load plugins from %s: %w", location, err)
		}
	}

	return pluginLoader, nil
}

func initEndorsementStore(v *viper.Viper, pluginLoader plugin.ILoader) (*endorsement.Store, error) {
	log.Info("initializing endorsement store")
	cfg, err := endorsement.StoreConfigFromViper(v)
	if err != nil {
		return nil, fmt.Errorf("could not load endorsement store configuration: %w", err)
	}

	endorsementStore, err := endorsement.NewStore(cfg, pluginLoader, log.Named("endorsement-store"))
	if err != nil {
		return nil, fmt.Errorf("could not create endorsement store: %w", err)
	}

	return endorsementStore, nil
}

func initPolicyManager(v *viper.Viper) (*policy.Manager, error) {
	log.Info("initializing policy manager")
	return policy.CreateManager(v, log.Named("policy"))
}

func initAttestationSchemes(v *viper.Viper, pluginLoader plugin.ILoader) (*scheme.Manager, error) {
	log.Info("initializing attestation schemes")

	log.Debug("loading scheme configuration")
	params, err := plugin.ParametersMapFromViper(v, scheme.PluginNameFromScheme)
	if err != nil {
		return nil, fmt.Errorf("could not load scheme configuration: %w", err)
	}

	log.Debug("creating scheme manager")
	manager, err := scheme.CreateManager(pluginLoader, params, log.Named("scheme"))
	if err != nil {
		return nil, fmt.Errorf("could not create scheme manager: %w", err)
	}

	return manager, nil
}

func initSigner(v *viper.Viper) (service.IEarSigner, error) {
	log.Info("loading EAR signer")
	earSigner, err := service.NewEarSigner(v, afero.NewOsFs())
	if err != nil {
		return nil, fmt.Errorf("EAR signer initialization failed: %w", err)
	}

	return earSigner, nil
}

func initServer(
	v *viper.Viper,
	endorsementStore *endorsement.Store,
	schemeManager *scheme.Manager,
	policyManager *policy.Manager,
	earSigner service.IEarSigner,
) (*service.Server, error) {
	log.Info("initializing trusted service")
	logger := log.Named("vts")

	log.Debug("loading trusted service configuration")
	cfg, err := vts.CreateServerConfigFromViper(v, logger)
	if err != nil {
		return nil, fmt.Errorf("could not load VTS configuration: %w", err)
	}

	log.Debug("creating GRPC server")
	server, err := service.CreateServer(
		cfg,
		endorsementStore,
		schemeManager,
		policyManager,
		earSigner,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("VTS initialisation failed: %w", err)
	}

	return server, nil
}

func serve(server *service.Server, done chan bool) {
	if err := server.Run(); err != nil {
		log.Error("VTS failed:", err)
	}

	done <- true
}

func trapSignals(sigs chan os.Signal, done chan bool) {
	sig := <-sigs

	log.Info(sig, " received, exiting")

	done <- true
}
