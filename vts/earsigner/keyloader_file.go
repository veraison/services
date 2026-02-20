// Copyright 2025-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package earsigner

import (
	"fmt"
	"net/url"

	"github.com/spf13/afero"
)

// FileKeyLoader is IKeyLoader implementation that loads the key from a file path.
type FileKeyLoader struct {
	fs afero.Fs
}

// NewFileKeyLoader creates a new FileKeyLoader using the specified Fs.
func NewFileKeyLoader(fs afero.Fs) *FileKeyLoader {
	return &FileKeyLoader{fs}
}

// Load they key from the specified URL. The url must be in one of the following formats:
//
//	<file-path>
//	file:<file-path>
//
// Where <file-path> is the abosute path to the file contianing the key.
func (o FileKeyLoader) Load(location *url.URL) ([]byte, error) {
	b, err := afero.ReadFile(o.fs, location.Path)
	if err != nil {
		return nil, fmt.Errorf("loading signing key from %q: %w", location.Path, err)
	}

	return b, nil
}
