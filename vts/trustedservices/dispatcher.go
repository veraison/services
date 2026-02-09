// Copyright 2022-2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package trustedservices

import (
	"encoding/json"
	"fmt"
	"os"
)

type clientDetails struct {
	Type     string   `json:"type"`
	Url      string   `json:"url"`
	Insecure bool     `json:"in-secure"`
	CaCerts  []string `json:"caCerts"`
	Hints    []string `json:"hint"`
}
type Dispatcher struct {
	ClientInfo map[string]clientDetails
}

var lvDispatcher Dispatcher

// LoadDispatchTable loads the Dispatch Table from the configuration
func LoadDispatchTable(fp string) error {
	data, err := os.ReadFile(fp)
	if err != nil {
		return fmt.Errorf("error reading dispatch table from %s: %w", fp, err)
	}
	if err := json.Unmarshal(data, &lvDispatcher); err != nil {
		return fmt.Errorf("error unmarshalling dispatch table: %w", err)
	}
	return nil
}
