// Copyright 2025-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package earsigner

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

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

	config, err := config.LoadDefaultConfig(o.context, config.WithRegion(region))
	if err != nil {
		return nil, err
	}

	client := secretsmanager.NewFromConfig(config)

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
