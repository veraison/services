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
	ListenAddr  = "localhost:8080"
	VerifierCfg = config.Store{
		// placeholder, empty for now
	}
	VTSClientCfg = config.Store{
		"vts-server.addr": "dns:127.0.0.1:50051",
	}
)

func main() {
	sessionManager := sessionmanager.NewSessionManagerTTLCache()
	vtsClient := vtsclient.NewGRPC(VTSClientCfg)
	verifier := verifier.New(VerifierCfg, vtsClient)
	apiHandler := api.NewHandler(sessionManager, verifier)
	apiServer(apiHandler, ListenAddr)
}

func apiServer(apiHandler api.IHandler, listenAddr string) {
	if err := api.NewRouter(apiHandler).Run(listenAddr); err != nil {
		log.Fatalf("Gin engine failed: %v", err)
	}
}
