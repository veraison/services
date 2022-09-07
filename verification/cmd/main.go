package main

import (
	"log"

	"github.com/veraison/services/config"
	"github.com/veraison/services/verification/api"
	"github.com/veraison/services/verification/sessionmanager"
	"github.com/veraison/services/verification/verifier"
	"github.com/veraison/services/vtsclient"
)

var (
	DefaultListenAddr = "localhost:8080"
)

func main() {
	cfg := config.NewYAMLReader()

	_, err := cfg.ReadFile("./config.yaml")
	if err != nil {
		log.Fatalf("config.yaml could not be read: %v", err)
	}

	vtsClientConfig := cfg.MustGetStore("vts-client")

	verifierConfig := config.Store{
		// placeholder, empty for now
	}

	apiServerConfig := cfg.MustGetStore("api-server")
	listenAddr, err := config.GetString(apiServerConfig, "listen-addr", &DefaultListenAddr)
	if err != nil {
		log.Fatalf("loading api-server configuration: %v", err)
	}

	sessionManager := sessionmanager.NewSessionManagerTTLCache()
	vtsClient := vtsclient.NewGRPC(vtsClientConfig)
	verifier := verifier.New(verifierConfig, vtsClient)
	apiHandler := api.NewHandler(sessionManager, verifier)
	apiServer(apiHandler, listenAddr)
}

func apiServer(apiHandler api.IHandler, listenAddr string) {
	if err := api.NewRouter(apiHandler).Run(listenAddr); err != nil {
		log.Fatalf("Gin engine failed: %v", err)
	}
}
