// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package log

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/template"

	"github.com/moogar0880/problems"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/veraison/services/config"
)

// Global root logger. This is the logger used when the top-level log functions
// are used. All services loggers are derived from this via Named().
var logger *zap.SugaredLogger

// zap uses AtomicLevel to dynamically change logging level across all loggers in a tree.
var level zap.AtomicLevel

// Initialize a fall-back logger to be used before Init() is called with proper config.
func init() {
	rawLogger, err := zap.NewDevelopmentConfig().Build()
	if err != nil {
		panic(err)
	}

	logger = rawLogger.Sugar()
	level = zap.NewAtomicLevel()
}

// Supported logger encodings. Currently, only the defaults provided by zap are
// listed. Additional encodings may be specified via zap.RegisterEncoder (see
// zap docs: https://pkg.go.dev/go.uber.org/zap@v1.23.0#RegisterEncoder).
var supportedEncodings = map[string]bool{
	"console": true,
	"json":    true,
}

//  Fully exposing encoder configuration in Veraison config is going to be a
//  lot of hassle (need to provide mapstructure serialisation for all the
//  underlying encoders; and, arguably, would be unwieldy for the end-user. On
//  the other hand, it may be useful to have different formatting in different
//  contexts. As a middle ground, allow selecting between pre-defined formats
//  in the config rather than exposing all the individual settings.
var encoderConfigs = map[string]zapcore.EncoderConfig{
	"production":  zap.NewProductionEncoderConfig(),
	"development": zap.NewDevelopmentEncoderConfig(),
	"bare": {
		TimeKey:        zapcore.OmitKey,
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      zapcore.OmitKey,
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  zapcore.OmitKey,
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    ColorCapitalLevelEncoder,
		EncodeTime:     zapcore.EpochTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	},
}

// Config is, essentially, and adapter for zap.Config that defines mapstructure
// tags for zap.Config fields, and provides any additional Veraison-specific
// validation.
// Note: zap.Config defines field tags for JSON and YAML. Even though Veraison
// uses YAML config files, we do not load them directly, but use Viper, which
// uses mapstructure internally. Thus the need for mapstructure tags.
// Additionally, zap specifies camel-case for its config fields, whereas
// Verasion uses all-lower-case-with-hyphen-delimiter fields.
type Config struct {
	// Level is the minimum enabled logging level. Note that this is a dynamic
	// level, so calling Config.Level.SetLevel will atomically change the log
	// level of all loggers descended from this config.
	Level string `mapstructure:"level"`
	// Development puts the logger in development mode, which changes the
	// behavior of DPanicLevel and takes stacktraces more liberally.
	Development bool `mapstructure:"development" config:"zerodefault"`
	// DisableCaller stops annotating logs with the calling function's file
	// name and line number. By default, all logs are annotated.
	DisableCaller bool `mapstructure:"disable-caller" config:"zerodefault"`
	// DisableStacktrace completely disables automatic stacktrace capturing. By
	// default, stacktraces are captured for WarnLevel and above logs in
	// development and ErrorLevel and above in production.
	DisableStacktrace bool `mapstructure:"disable-stacktrace" config:"zerodefault"`
	// Sampling sets a sampling policy. A nil SamplingConfig disables sampling.
	Sampling *zap.SamplingConfig `mapstructure:"sampling" config:"zerodefault"`
	// Encoding sets the logger's encoding. Valid values are "json" and
	// "console", as well as any third-party encodings registered via
	// RegisterEncoder.
	Encoding string `mapstructure:"encoding"`
	// OutputPaths is a list of URLs or file paths to write logging output to.
	// See Open for details.
	OutputPaths []string `mapstructure:"output-paths"`
	// ErrorOutputPaths is a list of URLs to write internal logger errors to.
	// The default is standard error.
	//
	// Note that this setting only affects internal errors; for sample code that
	// sends error-level logs to a different location from info- and debug-level
	// logs, see the package-level AdvancedConfiguration example.
	ErrorOutputPaths []string `mapstructure:"err-output-paths"`
	// InitialFields is a collection of fields to add to the root logger.
	InitialFields map[string]interface{} `mapstructure:"initial-fields" config:"zerodefault"`
	// Format is the only field that is not directly transposed into
	// zap.Config. Instead, it is used to look up a pre-defined
	// zapcore.EncoderConfig.
	Format string `mapstructure:"format"`
}

func (o Config) Validate() error {
	if _, err := zapcore.ParseLevel(o.Level); err != nil {
		return fmt.Errorf("invalid logging level: %q", o.Level)
	}

	if _, ok := supportedEncodings[o.Encoding]; !ok {
		var supported []string

		for k := range supportedEncodings {
			supported = append(supported, k)
		}

		return fmt.Errorf("unexpected encoding: %q; must be one of: %s",
			o.Encoding, strings.Join(supported, ", "))
	}

	if _, ok := encoderConfigs[o.Format]; !ok {
		var supportedFormats []string

		for k := range encoderConfigs {
			supportedFormats = append(supportedFormats, k)
		}

		sort.Strings(supportedFormats)

		return fmt.Errorf("unexpected format: %q; must be one of: %s",
			o.Format, strings.Join(supportedFormats, ", "))
	}

	return nil
}

func setLevel(name string) {
	canonicalName := strings.TrimSpace(strings.ToLower(name))

	switch canonicalName {
	case "debug", "trace":
		SetLevel(zap.DebugLevel)
	case "info":
		SetLevel(zap.InfoLevel)
	case "warn", "warning":
		SetLevel(zap.WarnLevel)
	case "error":
		SetLevel(zap.ErrorLevel)
	default:
		panic(fmt.Sprintf("unexpected level name: %q", name))
	}
}

// Zap returns a zap.Config populated with the field values of this config.
func (o Config) Zap() zap.Config {

	// XXX(setrofim): conceptually, this ought to be done inside
	// zap.Config.Build(), but since we can't override that, this seems
	// like the next best place.
	setLevel(o.Level)

	return zap.Config{
		Level:             level,
		Development:       o.Development,
		DisableCaller:     o.DisableCaller,
		DisableStacktrace: o.DisableStacktrace,
		Sampling:          o.Sampling,
		Encoding:          o.Encoding,
		OutputPaths:       o.OutputPaths,
		ErrorOutputPaths:  o.ErrorOutputPaths,
		InitialFields:     o.InitialFields,

		EncoderConfig: encoderConfigs[o.Format],
	}
}

// VerboseViper enables verbose output level for Viper loggers. This is exposed as a separate function as one might want to enable this before initializing logging.
func VerboseViper() {
	// jww (jwalterweatherman) is the logger used by Viper enabling this
	// allows debugging configuration issues.
	jww.SetLogThreshold(jww.LevelTrace)
	jww.SetStdoutThreshold(jww.LevelTrace)
}

// SetLevel sets the global logging level.
func SetLevel(l zapcore.Level) {
	level.SetLevel(l)
}

// GetLevel returns the current global logging level.
func GetLevel() zapcore.Level {
	return level.Level()
}

// Init initializes services global logging. All loggers are going to be
// derived from a root created here.
// The classifiers are used to identify the executable/application that is
// being logged, e.g. when naming the output files, they are applied to
// templated paths.
func Init(v *viper.Viper, classifiers map[string]interface{}) error {
	cfg := Config{
		Level:            "info",
		Encoding:         "console",
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		Format:           "bare",
	}

	loader := config.NewLoader(&cfg)
	if err := loader.LoadFromViper(v); err != nil {
		return err
	}

	if err := resolveTemplates(&cfg.OutputPaths, classifiers); err != nil {
		return err
	}

	if err := resolveTemplates(&cfg.ErrorOutputPaths, classifiers); err != nil {
		return err
	}

	rawLogger, err := cfg.Zap().Build()
	if err != nil {
		return err
	}

	logger = rawLogger.Sugar()

	return nil
}

func resolveTemplates(texts *[]string, vals map[string]interface{}) error {
	var processed []string // nolint:prealloc
	var rawBuff []byte

	if vals == nil {
		return nil
	}

	templ := template.New("paths")
	buff := bytes.NewBuffer(rawBuff)

	for _, text := range *texts {
		templ, err := templ.Parse(text)
		if err != nil {
			return err
		}

		buff.Truncate(0)
		if err := templ.Execute(buff, vals); err != nil {
			return err
		}

		processed = append(processed, buff.String())
	}

	*texts = processed

	return nil
}

// Sync loggers' io sinks.
func Sync() error {
	return logger.Sync()
}

// Create a new logger (derived from the global root) with the specified name.
func Named(name string) *zap.SugaredLogger {
	return logger.Named(name)
}

// NamedWriter creates an io.Writer that utilizes a zap logger with the
// specified name at the specifed level.
func NamedWriter(name string, level zapcore.Level) io.Writer {
	return WriterFromZap(Named(name), level)
}

type logWriter struct {
	write func(args ...interface{})
}

func (o logWriter) Write(p []byte) (int, error) {
	o.write(strings.TrimSuffix(string(p), "\n"))
	return len(p), nil
}

const (
	DebugLevel = zap.DebugLevel
	TraceLevel = zap.DebugLevel
	InfoLevel  = zap.InfoLevel
	WarnLevel  = zap.WarnLevel
	ErrorLevel = zap.ErrorLevel
)

// WriterFromZap returns an io.Writer utilzing the provided zap logger at the
// specified level.
func WriterFromZap(logger *zap.SugaredLogger, level zapcore.Level) io.Writer {
	var writeFunc func(args ...interface{})

	switch level {
	case zapcore.DebugLevel:
		writeFunc = logger.Debug
	case zapcore.InfoLevel:
		writeFunc = logger.Info
	case zapcore.WarnLevel:
		writeFunc = logger.Warn
	case zapcore.ErrorLevel:
		writeFunc = logger.Error
	default:
		panic(fmt.Sprintf("unexpected level name: %q", level))
	}

	return logWriter{writeFunc}
}

// Debug uses fmt.Sprint to construct and log a message.
func Debug(args ...interface{}) {
	logger.Debug(args...)
}

// Info uses fmt.Sprint to construct and log a message.
func Info(args ...interface{}) {
	logger.Info(args...)
}

// Warn uses fmt.Sprint to construct and log a message.
func Warn(args ...interface{}) {
	logger.Warn(args...)
}

// Error uses fmt.Sprint to construct and log a message.
func Error(args ...interface{}) {
	logger.Error(args...)
}

// DPanic uses fmt.Sprint to construct and log a message. In development, the
// logger then panics. (See DPanicLevel for details.)
func DPanic(args ...interface{}) {
	logger.DPanic(args...)
}

// Panic uses fmt.Sprint to construct and log a message, then panics.
func Panic(args ...interface{}) {
	logger.Panic(args...)
}

// Fatal uses fmt.Sprint to construct and log a message, then calls os.Exit.
func Fatal(args ...interface{}) {
	logger.Fatal(args...)
}

// Debugf uses fmt.Sprintf to log a templated message.
func Debugf(template string, args ...interface{}) {
	logger.Debugf(template, args...)
}

// Infof uses fmt.Sprintf to log a templated message.
func Infof(template string, args ...interface{}) {
	logger.Infof(template, args...)
}

// Warnf uses fmt.Sprintf to log a templated message.
func Warnf(template string, args ...interface{}) {
	logger.Warnf(template, args...)
}

// Errorf uses fmt.Sprintf to log a templated message.
func Errorf(template string, args ...interface{}) {
	logger.Errorf(template, args...)
}

// DPanicf uses fmt.Sprintf to log a templated message. In development, the
// logger then panics. (See DPanicLevel for details.)
func DPanicf(template string, args ...interface{}) {
	logger.DPanicf(template, args...)
}

// Panicf uses fmt.Sprintf to log a templated message, then panics.
func Panicf(template string, args ...interface{}) {
	logger.Panicf(template, args...)
}

// Fatalf uses fmt.Sprintf to log a templated message, then calls os.Exit.
func Fatalf(template string, args ...interface{}) {
	logger.Fatalf(template, args...)
}

// Debugw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
//
// When debug-level logging is disabled, this is much faster than
//
//	s.With(keysAndValues).Debug(msg)
func Debugw(msg string, keysAndValues ...interface{}) {
	logger.Debugw(msg, keysAndValues...)
}

// Infow logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func Infow(msg string, keysAndValues ...interface{}) {
	logger.Infow(msg, keysAndValues...)
}

// Warnw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func Warnw(msg string, keysAndValues ...interface{}) {
	logger.Warnw(msg, keysAndValues...)
}

// Errorw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func Errorw(msg string, keysAndValues ...interface{}) {
	logger.Errorw(msg, keysAndValues...)
}

// DPanicw logs a message with some additional context. In development, the
// logger then panics. (See DPanicLevel for details.) The variadic key-value
// pairs are treated as they are in With.
func DPanicw(msg string, keysAndValues ...interface{}) {
	logger.DPanicw(msg, keysAndValues...)
}

// Panicw logs a message with some additional context, then panics. The
// variadic key-value pairs are treated as they are in With.
func Panicw(msg string, keysAndValues ...interface{}) {
	logger.Panicw(msg, keysAndValues...)
}

// Fatalw logs a message with some additional context, then calls os.Exit. The
// variadic key-value pairs are treated as they are in With.
func Fatalw(msg string, keysAndValues ...interface{}) {
	logger.Fatalw(msg, keysAndValues...)
}

// Debugln uses fmt.Sprintln to construct and log a message.
func Debugln(args ...interface{}) {
	logger.Debugln(args...)
}

// Infoln uses fmt.Sprintln to construct and log a message.
func Infoln(args ...interface{}) {
	logger.Infoln(args...)
}

// Warnln uses fmt.Sprintln to construct and log a message.
func Warnln(args ...interface{}) {
	logger.Warnln(args...)
}

// Errorln uses fmt.Sprintln to construct and log a message.
func Errorln(args ...interface{}) {
	logger.Errorln(args...)
}

// DPanicln uses fmt.Sprintln to construct and log a message. In development, the
// logger then panics. (See DPanicLevel for details.)
func DPanicln(args ...interface{}) {
	logger.DPanicln(args...)
}

// Panicln uses fmt.Sprintln to construct and log a message, then panics.
func Panicln(args ...interface{}) {
	logger.Panicln(args...)
}

// Fatalln uses fmt.Sprintln to construct and log a message, then calls os.Exit.
func Fatalln(args ...interface{}) {
	logger.Fatalln(args...)
}

// LogProblem logs a problems.StatusProblem reported  by the api. 500 probelms
// are logged as errors, everything else is logged as warnings.
func LogProblem(logger *zap.SugaredLogger, prob *problems.DefaultProblem) {
	var logFunc func(msg string, args ...interface{})

	if prob.Status >= 500 {
		logFunc = logger.Errorw
	} else {
		logFunc = logger.Warnw
	}

	logFunc("problem encountered", "title", prob.Title, "detail", prob.Detail)
}
