// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/viper"
	"github.com/veraison/services/pkg/config"
	"github.com/veraison/services/pkg/util"
	"github.com/veraison/services/pkg/vts"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	DefaultProtocol        = "https"
	DefaultMaxPayloadSize  = "1MB"
	DefaultDiscoveryMaxAge = "1m"
)

type rawFrontendConfig struct {
	ListenAddr      string `mapstructure:"listen-addr" valid:"dialstring"`
	Protocol        string `mapstructure:"protocol" valid:"in(http|https)"`
	Cert            string `mapstructure:"cert" config:"zerodefault"`
	CertKey         string `mapstructure:"cert-key" config:"zerodefault"`
	DiscoveryMaxAge string `mapstructure:"discovery-max-age" config:"zerodefault"`
	MaxPayloadSize  string `mapstructure:"max-payload-size"`
}

type rawVtsConfig struct {
	ServerAddress string   `mapstructure:"server-addr" valid:"dialstring"`
	UseTLS        bool     `mapstructure:"tls" config:"zerodefault"`
	TLSCACerts    []string `mapstructure:"tls-ca-certs" config:"zerodefault"`
}

type Config struct {
	ListenAddr      string                           `json:"listen-addr"`
	Protocol        string                           `json:"protocol"`
	MaxPayloadSize  int64                            `json:"max-payload-size"`
	DiscoveryMaxAge time.Duration                    `json:"discovery-max-age"`
	VTSAddress      string                           `json:"vts-address"`
	Cert            string                           `json:"cert"`
	CertKey         string                           `json:"cert-key"`
	TLSCACerts      []string                         `json:"tls-ca-certs"`
	TLSCreds        credentials.TransportCredentials `json:"-"`
}

func (o *Config) String() string {
	encoded, _ := json.Marshal(o)
	return string(encoded)
}

func (o Config) VTS() *vts.ClientConfig {
	return &vts.ClientConfig{
		ServerAddress: o.VTSAddress,
		TLSCreds:      o.TLSCreds,
	}
}

func CreateConfig(frontendViper, vtsViper *viper.Viper, defaultPort int) (*Config, error) {
	rawFrontendCfg, err := readRawFrontendConfig(frontendViper, defaultPort)
	if err != nil {
		return nil, err
	}

	rawVtsCfg, err := readRawVtsConfig(vtsViper)
	if err != nil {
		return nil, err
	}

	maxPayloadSize, err := config.ParseBytesSize(rawFrontendCfg.MaxPayloadSize)
	if err != nil {
		return nil, fmt.Errorf("could not parse max-payload-size: %w", err)
	}

	discoverMaxAge, err := time.ParseDuration(rawFrontendCfg.DiscoveryMaxAge)
	if err != nil {
		return nil, fmt.Errorf("could not parse discovery-max-age: %w", err)
	}

	if (rawFrontendCfg.Protocol == "https" || rawVtsCfg.UseTLS) &&
		(rawFrontendCfg.Cert == "" || rawFrontendCfg.CertKey == "") {
		return nil, errors.New(`both cert and cert-key must be specified when protocol is "https" ` +
			`or TLS is enabled for VTS`)
	}

	var creds credentials.TransportCredentials
	if rawVtsCfg.UseTLS {
		creds, err = util.LoadTLSCreds(
			rawFrontendCfg.Cert,
			rawFrontendCfg.CertKey,
			rawVtsCfg.TLSCACerts,
		)
		if err != nil {
			return nil, err
		}
	} else {
		creds = insecure.NewCredentials()
	}

	return &Config{
		ListenAddr:      rawFrontendCfg.ListenAddr,
		Protocol:        rawFrontendCfg.Protocol,
		MaxPayloadSize:  maxPayloadSize,
		DiscoveryMaxAge: discoverMaxAge,
		VTSAddress:      rawVtsCfg.ServerAddress,
		Cert:            rawFrontendCfg.Cert,
		CertKey:         rawFrontendCfg.CertKey,
		TLSCACerts:      rawVtsCfg.TLSCACerts,
		TLSCreds:        creds,
	}, nil
}

func readRawFrontendConfig(v *viper.Viper, defaultPort int) (*rawFrontendConfig, error) {
	cfg := &rawFrontendConfig{
		ListenAddr:      fmt.Sprintf("localhost:%d", defaultPort),
		Protocol:        DefaultProtocol,
		DiscoveryMaxAge: DefaultDiscoveryMaxAge,
		MaxPayloadSize:  DefaultMaxPayloadSize,
	}

	loader := config.NewLoader(cfg)
	if err := loader.LoadFromViper(v); err != nil {
		return nil, fmt.Errorf("could not load frontend config: %w", err)
	}

	return cfg, nil
}

func readRawVtsConfig(v *viper.Viper) (*rawVtsConfig, error) {
	cfg := &rawVtsConfig{
		ServerAddress: vts.DefaultServerAddr,
	}

	loader := config.NewNonExclusiveLoader(cfg)
	if err := loader.LoadFromViper(v); err != nil {
		return nil, fmt.Errorf("could not load vTS config: %w", err)
	}

	return cfg, nil
}
