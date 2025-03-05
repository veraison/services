// Copyright 2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package earsigner

import (
	"context"
	"fmt"
	"net/url"

	"github.com/spf13/afero"
)

type KeyLoader struct {
	loaders map[string]IKeyLoader
}

func NewKeyLoader(fs afero.Fs) *KeyLoader {
	return &KeyLoader{
		loaders: map[string]IKeyLoader{
			"file": NewFileKeyLoader(fs),
			"aws": NewAwsKeyLoader(context.TODO()),
		},
	}
}

func (o KeyLoader) Load(location *url.URL) ([]byte, error) {
	var scheme string

	if location.Scheme == "" {
		scheme = "file"
	} else {
		scheme =  location.Scheme
	}

	actualLoader, ok := o.loaders[scheme]
	if !ok {
		return nil, fmt.Errorf("invalid key loader scheme: %s", scheme)
	}

	return actualLoader.Load(location)
}
