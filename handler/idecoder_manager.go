// Copyright 2022-2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package handler

import "go.uber.org/zap"

type IDecoderManager interface {
	Init(dir string, logger *zap.SugaredLogger) error
	Close() error
	Dispatch(mediaType string, data []byte) (*EndorsementHandlerResponse, error)
	IsSupportedMediaType(mediaType string) bool
	GetSupportedMediaTypes() []string
}
