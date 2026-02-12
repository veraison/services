// Copyright 2022-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package trustedservices

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDispatcher_ok(t *testing.T) {
	fp := "testdata/dispatch-table.json"
	_, err := NewDispatcher(fp)
	require.NoError(t, err)
}

func TestLookupClientNameFromMediaType_OK(t *testing.T) {
	// Initial condition
	fp := "testdata/dispatch-table.json"
	d, err := NewDispatcher(fp)
	require.NoError(t, err)
	// Expected Name
	exp := "veraison-local-client"
	mt := "application/vnd.veraison.tsm-report+cbor; provider=arm_psa"
	name, err := d.LookupClientNameFromMediaType(mt)
	require.NoError(t, err)
	assert.Equal(t, exp, name)
}

func TestLookupClientNameFromMediaType_NOK(t *testing.T) {
	// Initial condition
	fp := "testdata/dispatch-table.json"
	d, err := NewDispatcher(fp)
	require.NoError(t, err)
	expError := "unable to lookup name for media type: application/vnd.veraison.tsm-report+cbor; provider=remote"
	mt := "application/vnd.veraison.tsm-report+cbor; provider=remote"
	_, err = d.LookupClientNameFromMediaType(mt)
	require.NotNil(t, err)
	assert.EqualError(t, err, expError)

	d = &Dispatcher{}
	_, err = d.LookupClientNameFromMediaType(mt)
	expError = "no client data to look for"
	require.NotNil(t, err)
	assert.EqualError(t, err, expError)
}

func TestLookupClientCfgFromMediaType_OK(t *testing.T) {
	var cfg ClientConfig
	// Initial condition
	fp := "testdata/dispatch-table.json"
	d, err := NewDispatcher(fp)
	require.NoError(t, err)
	// Expected Client Cfg
	exp := ClientConfig{DiscoveryURL: "https://localhost:8443", CACerts: []string{"../../../deployments/docker/src/certs/rootCA.crt"}, Insecure: true, crURL: ""}
	mt := "application/vnd.veraison.tsm-report+cbor; provider=arm_psa"
	data, err := d.LookupClientCfgFromMediaType(mt)
	require.NoError(t, err)
	err = json.Unmarshal(data, &cfg)
	require.NoError(t, err)
	assert.Equal(t, cfg.DiscoveryURL, exp.DiscoveryURL)
	assert.Equal(t, len(cfg.CACerts), len(exp.CACerts))
	for i, cert := range exp.CACerts {
		assert.Equal(t, cert, cfg.CACerts[i])
	}
	assert.Equal(t, cfg.Insecure, exp.Insecure)
}

func TestLookupClientCfgFromMediaType_NOK(t *testing.T) {
	// Initial condition
	fp := "testdata/dispatch-table.json"
	d, err := NewDispatcher(fp)
	require.NoError(t, err)
	expError := "unable to lookup client config for media type: application/vnd.veraison.tsm-report+cbor; provider=remote"
	mt := "application/vnd.veraison.tsm-report+cbor; provider=remote"
	_, err = d.LookupClientCfgFromMediaType(mt)
	require.NotNil(t, err)
	assert.EqualError(t, err, expError)

	d = &Dispatcher{}
	_, err = d.LookupClientCfgFromMediaType(mt)
	expError = "no client data to look for"
	require.NotNil(t, err)
	assert.EqualError(t, err, expError)
}
