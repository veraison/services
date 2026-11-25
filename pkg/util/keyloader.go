// Copyright 2025-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package util

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/afero"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// IKeyLoader defines the interface for loading signing keys
type IKeyLoader interface {
	// Load the signing key from the specified location, returning it as a
	// []byte, if successful, or an error if not.
	Load(location *url.URL) ([]byte, error)
}

type KeyLoader struct {
	loaders map[string]IKeyLoader
}

func NewKeyLoader(fs afero.Fs) *KeyLoader {
	return &KeyLoader{
		loaders: map[string]IKeyLoader{
			"file": NewFileKeyLoader(fs),
			"aws":  NewAwsKeyLoader(context.TODO()),
		},
	}
}

func (o KeyLoader) Load(location *url.URL) ([]byte, error) {
	var scheme string

	if location.Scheme == "" {
		scheme = "file"
	} else {
		scheme = location.Scheme
	}

	actualLoader, ok := o.loaders[scheme]
	if !ok {
		return nil, fmt.Errorf("invalid key loader scheme: %s", scheme)
	}

	return actualLoader.Load(location)
}

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
// Where <file-path> is the absolute path to the file containing the key.
func (o FileKeyLoader) Load(location *url.URL) ([]byte, error) {
	b, err := afero.ReadFile(o.fs, location.Path)
	if err != nil {
		return nil, fmt.Errorf("loading signing key from %q: %w", location.Path, err)
	}

	return b, nil
}

// AwsKeyLoader is IKeyLoader implementation that loads the key from AWS Secrets Manager.
type AwsKeyLoader struct {
	context context.Context
}

// NewAwsKeyLoader creates a new NewAwsKeyLoader using the specified Context.
func NewAwsKeyLoader(c context.Context) *AwsKeyLoader {
	return &AwsKeyLoader{c}
}

// Load the key from the specified URL. The url must be in the following format:
//
//	aws:<region>/<secret-name>
//
// Where <region> is the AWS region (e.g. eu-west-1), and <secret-name> is the
// name under which the key is stored in the AWS Secrets Manager
func (o AwsKeyLoader) Load(location *url.URL) ([]byte, error) {
	parts := strings.Split(location.Opaque, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid AWS key location: %s", location.Opaque)
	}
	region, name := parts[0], parts[1]

	cfg, err := config.LoadDefaultConfig(o.context, config.WithRegion(region))
	if err != nil {
		return nil, err
	}

	client := secretsmanager.NewFromConfig(cfg)

	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(name),
		VersionStage: aws.String("AWSCURRENT"),
	}

	resp, err := client.GetSecretValue(o.context, input)
	if err != nil {
		return nil, err
	}

	return []byte(*resp.SecretString), nil
}
