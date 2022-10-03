package main

import (
	"errors"
	"log"
	"os"

	"github.com/setrofim/viper"
	"github.com/veraison/services/verification/api"
	"github.com/veraison/services/verification/sessionmanager"
	"github.com/veraison/services/verification/verifier"
	"github.com/veraison/services/vtsclient"
)

var (
	ListenAddr = "localhost:8080"
)

func main() {

	VTSClientCfg := viper.New()
	VTSClientCfg.SetDefault("vts-server.addr", "dns:127.0.0.1:50051")

	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	VTSClientCfg.AddConfigPath(wd)
	VTSClientCfg.SetConfigType("yaml")
	VTSClientCfg.SetConfigName("config")

	err = VTSClientCfg.ReadInConfig()
	if errors.As(err, &viper.ConfigFileNotFoundError{}) {
		// If there is no config file, use the defaults set above.
		err = nil
	}

	if err != nil {
		log.Fatal(err)
	}

	sessionManager := sessionmanager.NewSessionManagerTTLCache()
	vtsClient := vtsclient.NewGRPC(VTSClientCfg)
	verifier := verifier.New(viper.New(), vtsClient)
	apiHandler := api.NewHandler(sessionManager, verifier)
	apiServer(apiHandler, ListenAddr)
}

func apiServer(apiHandler api.IHandler, listenAddr string) {
	if err := api.NewRouter(apiHandler).Run(listenAddr); err != nil {
		log.Fatalf("Gin engine failed: %v", err)
	}
}
