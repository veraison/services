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

type cfg struct {
	ListenAddr string `mapstructure:"listen-addr" valid:"dialstring"`
}

func main() {
	v, err := config.ReadRawConfig("", true)
	if err != nil {
		log.Fatalf("Could not read config: %v", err)
	}

	subs, err := config.GetSubs(v, "*vts", "*verifier", "*verification")
	if err != nil {
		log.Fatalf("Could not read config: %v", err)
	}

	sessionManager := sessionmanager.NewSessionManagerTTLCache()

	vtsClient := vtsclient.NewGRPC()
	if err := vtsClient.Init(subs["vts"]); err != nil {
		log.Fatalf("Could not initialize VTS client: %v", err)
	}

	verifier := verifier.New(subs["verifier"], vtsClient)
	apiHandler := api.NewHandler(sessionManager, verifier)

	cfg := cfg{ListenAddr: DefaultListenAddr}
	loader := config.NewLoader(&cfg)
	if err := loader.LoadFromViper(subs["verification"]); err != nil {
		log.Fatalf("Could not load verfication config: %v", err)

	}

	apiServer(apiHandler, cfg.ListenAddr)
}

func apiServer(apiHandler api.IHandler, listenAddr string) {
	if err := api.NewRouter(apiHandler).Run(listenAddr); err != nil {
		log.Fatalf("Gin engine failed: %v", err)
	}
}
