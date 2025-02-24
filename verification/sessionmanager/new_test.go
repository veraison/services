// Copyright 2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package sessionmanager

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func Test_cfg_Validate(t *testing.T) {
	c := cfg{
		Backend: DefaultBackend,
	}
	assert.NoError(t, c.Validate())

	c = cfg{
		Backend: "ttlcache",
		BackendConfigs: map[string]interface{}{
			"ttlcache": map[string]string{
				"ttl": "1m",
			},
		},
	}
	assert.NoError(t, c.Validate())

	c = cfg{
		Backend: DefaultBackend,
		BackendConfigs: map[string]interface{}{
			"unexpected": map[string]string{
				"ttl": "1m",
			},
		},
	}
	assert.ErrorContains(t, c.Validate(), "unexpected directives: unexpected")
}

func Test_New(t *testing.T) {
	_, err := New(viper.New())
	assert.NoError(t, err)

	c := viper.New()
	c.Set("backend", "invalid")

	_, err = New(c)
	assert.ErrorContains(t, err, `backend "invalid" is not supported`)
}
