// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package log

import (
	"github.com/fatih/color"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var levelColorizers map[zapcore.Level]func(a ...interface{}) string

func init() {
	levelColorizers = map[zapcore.Level]func(a ...interface{}) string{
		zap.DebugLevel: color.New(color.FgBlue).SprintFunc(),
		zap.InfoLevel:  color.New(color.FgGreen).SprintFunc(),
		zap.WarnLevel:  color.New(color.FgYellow).SprintFunc(),
		zap.ErrorLevel: color.New(color.FgRed).SprintFunc(),
		zap.PanicLevel: color.New(color.FgHiRed).SprintFunc(),
	}
}

// An encoder for zap log levels that colorizes them based on their value.
func ColorCapitalLevelEncoder(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	colorizer, ok := levelColorizers[l]
	if !ok {
		colorizer = levelColorizers[zap.ErrorLevel]
	}

	enc.AppendString(colorizer(l.CapitalString()))
}
