// Copyright 2022-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package vts

import (
	"crypto/x509"
	"strings"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc/credentials"

	"github.com/veraison/services/pkg/config"
	"github.com/veraison/services/pkg/util"
)

const DefaultServerAddr = "127.0.0.1:50051"

// TODO(tho) load balancing config
// See https://github.com/grpc/grpc/blob/master/doc/load-balancing.md
type rawConfig struct {
	ServerAddress string   `mapstructure:"server-addr" valid:"dialstring"`
	ListenAddress string   `mapstructure:"listen-addr" valid:"dialstring" config:"zerodefault"`
	UseTLS        bool     `mapstructure:"tls" config:"zerodefault"`
	ServerCert    string   `mapstructure:"cert" config:"zerodefault"`
	ServerCertKey string   `mapstructure:"cert-key" config:"zerodefault"`
	TLSCACerts    []string `mapstructure:"tls-ca-certs" config:"zerodefault"`
	CoRIMCACerts  []string `mapstructure:"corim-ca-certs" config:"zerodefault"`
}

type ServerConfig struct {
	ListenAddress string
	CoRIMCACerts  *x509.CertPool
	TLSCreds      credentials.TransportCredentials
}

func CreateServerConfigFromViper(v *viper.Viper, logger *zap.SugaredLogger) (*ServerConfig, error) {
	var err error

	raw := &rawConfig{ServerAddress: DefaultServerAddr}
	configLoader := config.NewLoader(raw)
	if err := configLoader.LoadFromViper(v); err != nil {
		return nil, err
	}

	logger.Debugw("VTS server configuration", "config", raw)

	serverConfig := ServerConfig{}

	if raw.ListenAddress != "" {
		serverConfig.ListenAddress = raw.ListenAddress
	} else {
		logger.Debug("setting listening port from server-addr")
		// note: the indexing will succeed as ServerAddress has been validated as a dialstring.
		serverConfig.ListenAddress = ":" + strings.Split(raw.ServerAddress, ":")[1]
	}

	logger.Info("loading CoRIM CA certificates")
	if serverConfig.CoRIMCACerts, err = util.LoadCACerts(raw.CoRIMCACerts); err != nil {
		return nil, err
	}

	if raw.UseTLS {
		logger.Info("loading TLS credentials")
		serverConfig.TLSCreds, err = util.LoadTLSCreds(raw.ServerCert, raw.ServerCertKey, raw.TLSCACerts)
		if err != nil {
			return nil, err
		}
	}

	return &serverConfig, nil
}

type ClientConfig struct {
	ServerAddress string
	TLSCreds      credentials.TransportCredentials
}
