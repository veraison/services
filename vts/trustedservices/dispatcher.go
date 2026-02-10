// Copyright 2022-2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package trustedservices

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/veraison/services/handler"
)

type ClientDetails struct {
	Type     string   `json:"type"`
	Url      string   `json:"url"`
	Insecure bool     `json:"in-secure"`
	CaCerts  []string `json:"caCerts"`
	Hints    []string `json:"hint"`
}
type Dispatcher struct {
	ClientInfo map[string]ClientDetails
}

var lvDispatcher Dispatcher

func NewDispatcher(fp string) (*Dispatcher, error) {
	lvDispatcher := &Dispatcher{ClientInfo: make(map[string]ClientDetails)}
	data, err := os.ReadFile(fp)
	if err != nil {
		return nil, fmt.Errorf("error reading dispatch table from %s: %w", fp, err)
	}
	jmap := make(map[string]json.RawMessage)
	err = json.Unmarshal(data, &jmap)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling table from %s: %w", fp, err)
	}
	for k, val := range jmap {
		var cd ClientDetails
		err = json.Unmarshal(val, &cd)
		// check error
		lvDispatcher.addClientInfo(k, &cd)
	}
	return lvDispatcher, nil
}

func (d *Dispatcher) addClientInfo(key string, val *ClientDetails) error {
	if d.ClientInfo == nil {
		return errors.New("no client info to add")
	}
	d.ClientInfo[key] = *val
	return nil
}

func (d *Dispatcher) GetClientConfigFromClientName(name string) (cfg []byte, err error) {
	if d.ClientInfo == nil {
		// return
	}

	found := false
	for k, val := range d.ClientInfo {
		if k == name {
			cfg, err = json.Marshal(val)
			if err != nil {
				// Check Error
			}
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("unable to locate client for name: %s", name)
	}
	return cfg, nil
}

func GetComponentVerifierFromMediaType(mt string) (handler.IComponentVerifierClientHandler, error) {

	return nil, nil
}
