package main

import (
	"log"

	"github.com/veraison/services/config"
	"github.com/veraison/services/verification/api"
	"github.com/veraison/services/verification/sessionmanager"
	"github.com/veraison/services/verification/verifier"
	"github.com/veraison/services/vtsclient"
)

func main() {
	cfg := config.NewYAMLReader()

	_, err := cfg.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("counfig.yaml could not be read.")
	}

	let vtsClientConfig = config.Store{
		"vts-server.addr": cfg.MustGetStore("vts-server.addr"),
	}

	let verifierConfig = config.Store {
		 // placeholder, empty for now
	}

		sessionManager := sessionmanager.NewSessionManagerTTLCache()
	vtsClient := vtsclient.NewGRPC(vtsClientConfig)
	verifier := verifier.New(verifierConfig, vtsClient)
	apiHandler := api.NewHandler(sessionManager, verifier)
	apiServer(apiHandler, cfg.MustGetStore("listen-addr"))
}

func apiServer(apiHandler api.IHandler, listenAddr string) {
	if err := api.NewRouter(apiHandler).Run(listenAddr); err != nil {
		log.Fatalf("Gin engine failed: %v", err)
	}
}
