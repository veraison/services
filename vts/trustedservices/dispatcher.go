// Copyright 2022-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package trustedservices

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

// TO DO The structure ClientData (Line 13 to Line 19) should be moved to a common location,
// where it should be accesible in code here as well, as Clients, decoding the cfgData
type ClientData struct {
	Type     string   `json:"type"`
	Url      string   `json:"url"`
	Insecure bool     `json:"in-secure"`
	CaCerts  []string `json:"ca-certs"`
	Hints    []string `json:"hints"`
}

// Representation of Client Configuration as read from the Dispatch File
type CfgData struct {
	ClientInfo map[string]ClientData
}

// Per Client Configuration data - to be passed to the Client Plugin Handler
type ClientConfig struct {
	DiscoveryURL string   `json:"url"`
	CACerts      []string `json:"ca_certs,omitempty"`
	Insecure     bool     `json:"insecure,omitempty"`
	crURL        string   // the challenge-response URL is discovered dynamically
}

type ClientInfo struct {
	Name string       // name of the Client, mapping to Scheme Name, Type in ClientData
	cfg  ClientConfig // Client Configuration
}

// Dispatcher stores, Client Details per supported component MediaTypes
// This enables efficient lookup
type Dispatcher struct {
	Client map[string]ClientInfo
}

func NewDispatcher(fp string) (*Dispatcher, error) {
	dt := &Dispatcher{Client: make(map[string]ClientInfo)}

	cd, err := loadcfgData(fp)
	if err != nil {
		return nil, fmt.Errorf("error loading cfg data: %w", err)
	}
	if err := dt.createDispatcher(cd); err != nil {
		return nil, fmt.Errorf("unable to create distach table: %w", err)
	}
	return dt, nil
}

func loadcfgData(fp string) (*CfgData, error) {
	cfg := &CfgData{ClientInfo: make(map[string]ClientData)}
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
		var cd ClientData
		err = json.Unmarshal(val, &cd)
		if err != nil {
			return nil, fmt.Errorf("error decoding client data for client: %s, from file %s: %w", k, fp, err)
		}
		cfg.ClientInfo[k] = cd
		if len(cd.Hints) == 0 {
			return nil, errors.New("no Hints provided")
		}
	}

	return cfg, nil
}

func (d *Dispatcher) createDispatcher(cd *CfgData) error {
	if cd.ClientInfo == nil {
		return errors.New("no cfg data to create table")
	}
	for _, val := range cd.ClientInfo {
		if err := d.createTableEntriesfromCfgData(&val); err != nil {
			return err
		}
	}
	return nil
}

func (d *Dispatcher) createTableEntriesfromCfgData(data *ClientData) error {
	var cl ClientInfo
	cl.Name = data.Type
	cl.cfg.DiscoveryURL = data.Url
	cl.cfg.CACerts = make([]string, len(data.CaCerts))
	cl.cfg.crURL = data.Url
	cl.cfg.Insecure = data.Insecure
	copy(cl.cfg.CACerts, data.CaCerts)
	// Hints is a list of MediaTypes supported by the individual client
	mts := data.Hints
	for _, mt := range mts {
		d.Client[mt] = cl
	}
	return nil
}

func (d *Dispatcher) LookupClientNameFromMediaType(mt string) (name string, err error) {
	if d.Client == nil {
		return "", errors.New("no client information to look for")
	}
	data, ok := d.Client[mt]
	if !ok {
		return "", fmt.Errorf("unable to lookup name for media type: %s", mt)
	}
	name = data.Name
	return name, nil
}

func (d *Dispatcher) LookupClientCfgFromMediaType(mt string) ([]byte, error) {
	if d.Client == nil {
		return nil, errors.New("no client information to look for")
	}
	data, ok := d.Client[mt]
	if !ok {
		return nil, fmt.Errorf("unable to lookup client config for media type: %s", mt)
	}

	cfg := data.cfg
	jc, err := json.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal client configuration %w", err)
	}
	return jc, nil
}

func (d *CfgData) lookupClientInfoFromMediaType(mt string) (*ClientData, error) {
	for _, val := range d.ClientInfo {
		for _, media := range val.Hints {
			if mt == media {
				return &val, nil
			}
		}
	}
	return nil, fmt.Errorf("unable to locate client information for media type: %s", mt)
}
