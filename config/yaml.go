// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package config

import (
	"fmt"
	"os"

	yaml "gopkg.in/yaml.v3"
)

type YAMLReader struct {
	Stores map[string]Store
}

func NewYAMLReader() *YAMLReader {
	var reader YAMLReader
	reader.Init()
	return &reader
}

func (o *YAMLReader) Init() {
	o.Stores = make(map[string]Store)
}

func (o *YAMLReader) Read(buf []byte) (int, error) {
	var sm map[string]map[string]interface{}
	if err := yaml.Unmarshal(buf, &sm); err != nil {
		return 0, err
	}

	// check no errors before updating reader stores.
	processed := make(map[string]Store)

	for name := range sm {
		if _, ok := o.Stores[name]; ok {
			return 0, fmt.Errorf("%w: %q", ErrStoreAlreadyExists, name)
		}

		store := make(Store)
		collapseMap(sm[name], store, "")
		processed[name] = store
	}

	for name, store := range processed {
		o.Stores[name] = store
	}

	return len(buf), nil
}

func (o *YAMLReader) ReadFile(path string) (int, error) {
	buf, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}

	return o.Read(buf)
}

func (o *YAMLReader) GetStores() map[string]Store {
	out := make(map[string]Store)

	for k, v := range o.Stores {
		out[k] = v
	}

	return out
}

func (o *YAMLReader) GetStore(name string) (Store, error) {
	s, ok := o.Stores[name]
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrStoreNotFound, name)
	}

	return s, nil
}

func (o *YAMLReader) MustGetStore(name string) Store {
	s, err := o.GetStore(name)
	if err != nil {
		s = Store{}
	}

	return s
}

// Collapse a map possibly containing nested maps into a single level,
// generating unique keys by prefxing dot-spearated nesting levels to the nested keys.
func collapseMap(in map[string]interface{}, out map[string]interface{}, prefix string) {
	for k, v := range in {
		m, ok := v.(map[string]interface{})
		if ok {
			newPrefix := prefix + k + "."
			collapseMap(m, out, newPrefix)
		} else {
			out[prefix+k] = v
		}
	}
}
