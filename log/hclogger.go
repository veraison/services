// Copyright 2022-2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package log

import (
	"io"
	stdlog "log"
	"reflect"
	"unsafe"

	"github.com/hashicorp/go-hclog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapio"
)

// HCLogger is a wrapper around zap logger used by Veraison that implements
// go-hclog.Logger interface expected by go-plugin plugins.
type HCLogger struct {
	logger   *zap.SugaredLogger
	internal bool

	name *string
}

func NewLogger(logger *zap.SugaredLogger) *HCLogger {
	return &HCLogger{logger: logger, internal: false}
}

// NewInternalLogger returns a new logger that logs all Info level messages at
// Debug level. This is to allow treating info-level messages form 3rd-party
// libararies as debug-level for our services.
func NewInternalLogger(logger *zap.SugaredLogger) *HCLogger {
	return &HCLogger{logger: logger, internal: true}
}

func (o *HCLogger) Log(level hclog.Level, msg string, args ...interface{}) {
	switch level {
	case hclog.Trace:
		o.Trace(msg, args...)
	case hclog.Debug:
		o.Debug(msg, args...)
	case hclog.Info:
		o.Info(msg, args...)
	case hclog.Warn:
		o.Warn(msg, args...)
	case hclog.Error:
		o.Error(msg, args...)
	default:
		// Panic if we run into an unexpected level, and logging is in
		// development mode. In production, this will be logged at the
		// highest level (i.e. always appear in the output).
		o.logger.DPanicw(msg, args...)
	}
}

// Emit a message and key/value pairs at the TRACE level
func (o *HCLogger) Trace(msg string, args ...interface{}) {
	o.logger.Debugw(msg, args...)
}

// Emit a message and key/value pairs at the DEBUG level
func (o *HCLogger) Debug(msg string, args ...interface{}) {
	o.logger.Debugw(msg, args...)
}

// Emit a message and key/value pairs at the INFO level
func (o *HCLogger) Info(msg string, args ...interface{}) {
	if o.internal {
		o.logger.Debugw(msg, args...)
	} else {
		o.logger.Infow(msg, args...)
	}
}

// Emit a message and key/value pairs at the WARN level
func (o *HCLogger) Warn(msg string, args ...interface{}) {
	o.logger.Warnw(msg, args...)
}

// Emit a message and key/value pairs at the ERROR level
func (o *HCLogger) Error(msg string, args ...interface{}) {
	o.logger.Errorw(msg, args...)
}

// Indicate if TRACE logs would be emitted. This and the other Is* guards
// are used to elide expensive logging code based on the current level.
func (o *HCLogger) IsTrace() bool {
	// zap does not implement trace level, so we treat it as debug.
	return GetLevel() <= zap.DebugLevel
}

// Indicate if DEBUG logs would be emitted. This and the other Is* guards
func (o *HCLogger) IsDebug() bool {
	return GetLevel() <= zap.DebugLevel
}

// Indicate if INFO logs would be emitted. This and the other Is* guards
func (o *HCLogger) IsInfo() bool {
	return GetLevel() <= zap.InfoLevel
}

// Indicate if WARN logs would be emitted. This and the other Is* guards
func (o *HCLogger) IsWarn() bool {
	return GetLevel() <= zap.WarnLevel
}

// Indicate if ERROR logs would be emitted. This and the other Is* guards
func (o *HCLogger) IsError() bool {
	return GetLevel() <= zap.ErrorLevel
}

// ImpliedArgs returns With key/value pairs
func (o *HCLogger) ImpliedArgs() []interface{} {
	return nil
}

// Creates a sublogger that will always have the given key/value pairs
func (o *HCLogger) With(args ...interface{}) hclog.Logger {
	return &HCLogger{logger: o.logger.With(args...)}
}

// Returns the Name of the logger
func (o *HCLogger) Name() string {
	var readName string

	if o.name == nil {
		fv := reflect.ValueOf(o.logger).Elem().FieldByName("base").Elem().FieldByName("name")
		readName = fv.String()
		o.name = &readName
	}

	return *o.name
}

// Create a logger that will prepend the name string on the front of all messages.
// If the logger already has a name, the new value will be appended to the current
// name. That way, a major subsystem can use this to decorate all it's own logs
// without losing context.
func (o *HCLogger) Named(name string) hclog.Logger {
	return &HCLogger{logger: o.logger.Named(name)}
}

// Create a logger that will prepend the name string on the front of all messages.
// This sets the name of the logger to the value directly, unlike Named which honor
// the current name as well.
func (o *HCLogger) ResetNamed(name string) hclog.Logger {
	// SugaredLoggers does not have an equivalent method -- it provides no
	// way to reset the name heirarchy. So we reflect to override the
	// hidden field value directly. This is cursed, but I can't think of a
	// nicer way of doing this, and I'm sure it'll be ok... (famous last
	// words)
	newLogger := o.logger.Named(name)

	fv := reflect.ValueOf(newLogger).Elem().FieldByName("base").Elem().FieldByName("name")

	//  "name" is unexported so we can't just SetString it directly, but...
	fv = reflect.NewAt(fv.Type(), unsafe.Pointer(fv.UnsafeAddr())).Elem()

	fv.SetString(name)

	return &HCLogger{logger: newLogger}
}

// Updates the level. This should affect all related loggers as well,
// unless they were created with IndependentLevels. If an
// implementation cannot update the level on the fly, it should no-op.
func (o *HCLogger) SetLevel(level hclog.Level) {
	// NO-OP
	// We do not want to allow plugins to change logging level.
}

// Returns the current level
func (o *HCLogger) GetLevel() hclog.Level {
	return hclog.Level(o.logger.Level())
}

// Return a value that conforms to the stdlib log.Logger interface
func (o *HCLogger) StandardLogger(opts *hclog.StandardLoggerOptions) *stdlog.Logger {
	return zap.NewStdLog(o.logger.Desugar())
}

// Return a value that conforms to io.Writer, which can be passed into log.SetOutput()
func (o *HCLogger) StandardWriter(opts *hclog.StandardLoggerOptions) io.Writer {
	return &zapio.Writer{Log: o.logger.Desugar(), Level: GetLevel()}
}
