// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package log

import (
	"net/url"
	"testing"
	"time"

	"github.com/fatih/color"
	"github.com/hashicorp/go-hclog"
	"github.com/moogar0880/problems"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zapio"
)

type testSink struct {
	URL  *url.URL
	Data *[]byte
}

func (o *testSink) Write(p []byte) (n int, err error) {
	*o.Data = append(*o.Data, p...)
	return len(p), nil
}

func (o testSink) Sync() error {
	return nil
}

func (o testSink) Close() error {
	return nil
}

// https://www.youtube.com/watch?v=kccONko4xYE&t=3
func testSinkFactoryFactory(dest *[]byte) func(*url.URL) (zap.Sink, error) {
	return func(u *url.URL) (zap.Sink, error) {
		return &testSink{URL: u, Data: dest}, nil
	}
}

func Test_Init_ok(t *testing.T) {
	var output []byte
	err := zap.RegisterSink("test", testSinkFactoryFactory(&output))
	require.NoError(t, err)

	color.NoColor = true

	v := viper.New()
	v.Set("output-paths", "test:")
	v.Set("err-output-paths", "test:")
	v.Set("format", "bare")
	v.Set("level", "debug")

	err = Init(v, nil)
	require.NoError(t, err)

	Debug("debug")
	Debugln("debugln")
	Debugw("debugw", "a", 1)
	Debugf("debugf %d", 1)

	Info("info")
	Infoln("infoln")
	Infow("infow", "a", 1)
	Infof("infof %d", 1)

	Warn("warn")
	Warnln("warnln")
	Warnw("warnw", "a", 1)
	Warnf("warnf %d", 1)

	Error("error")
	Errorln("errorln")
	Errorw("errorw", "a", 1)
	Errorf("errorf %d", 1)

	err = Sync()
	require.NoError(t, err)

	expectedOutput := "" +
		"DEBUG	debug\nDEBUG	debugln\nDEBUG	debugw	{\"a\": 1}\nDEBUG	debugf 1\n" +
		"INFO	info\nINFO	infoln\nINFO	infow	{\"a\": 1}\nINFO	infof 1\n" +
		"WARN	warn\nWARN	warnln\nWARN	warnw	{\"a\": 1}\nWARN	warnf 1\n" +
		"ERROR	error\nERROR	errorln\nERROR	errorw	{\"a\": 1}\nERROR	errorf 1\n"

	assert.Equal(t, expectedOutput, string(output))
}

func Test_LogProblem(t *testing.T) {
	var output []byte
	err := zap.RegisterSink("test2", testSinkFactoryFactory(&output))
	require.NoError(t, err)

	color.NoColor = true

	v := viper.New()
	v.Set("output-paths", "test2:")
	v.Set("err-output-paths", "test2:")
	v.Set("format", "bare")
	v.Set("level", "debug")

	err = Init(v, nil)
	require.NoError(t, err)
	logger := Named("test")

	prob := problems.DefaultProblem{
		Title:  "status500",
		Detail: "internal server error",
		Status: 500,
	}

	LogProblem(logger, &prob)

	assert.Contains(t, string(output), "ERROR\ttest\tproblem encountered")

	output = []byte{}

	prob = problems.DefaultProblem{
		Title:  "status500",
		Detail: "internal server error",
		Status: 400,
	}

	LogProblem(logger, &prob)

	assert.Contains(t, string(output), "WARN\ttest\tproblem encountered")
}

func Test_SetGet_Level(t *testing.T) {
	SetLevel(zap.ErrorLevel)
	assert.Equal(t, zap.ErrorLevel, GetLevel())

	setLevel("debug")
	assert.Equal(t, zap.DebugLevel, GetLevel())
}

func Test_HCLogger(t *testing.T) {
	var output []byte
	err := zap.RegisterSink("test3", testSinkFactoryFactory(&output))
	require.NoError(t, err)

	color.NoColor = true

	v := viper.New()
	v.Set("output-paths", "test3:")
	v.Set("err-output-paths", "test3:")
	v.Set("format", "bare")
	v.Set("level", "info")

	err = Init(v, nil)
	require.NoError(t, err)
	logger := Named("test")

	hcLogger := NewLogger(logger)

	hcLogger.Log(hclog.Warn, "warn msg")
	hcLogger.Info("info msg")
	hcLogger.Error("err msg")

	expected := "WARN\ttest\twarn msg\nINFO\ttest\tinfo msg\nERROR\ttest\terr msg\n"

	assert.Equal(t, expected, string(output))

	assert.False(t, hcLogger.IsTrace())
	assert.False(t, hcLogger.IsDebug())
	assert.True(t, hcLogger.IsInfo())
	assert.True(t, hcLogger.IsWarn())
	assert.True(t, hcLogger.IsError())

	assert.Equal(t, "test", hcLogger.Name())

	subLogger := hcLogger.Named("sub")
	assert.Equal(t, "test.sub", subLogger.Name())

	newLogger := subLogger.ResetNamed("new")
	assert.Equal(t, "new", newLogger.Name())

	output = []byte{}

	withLogger := hcLogger.With("a", 1)
	withLogger.Info("info msg")
	assert.Equal(t, "INFO\ttest\tinfo msg\t{\"a\": 1}\n", string(output))

	output = []byte{}

	writer := hcLogger.StandardWriter(nil)
	n, err := writer.Write([]byte("writer msg"))
	assert.NoError(t, err)
	assert.Equal(t, 10, n)

	err = (writer.(*zapio.Writer)).Sync()
	require.NoError(t, err)

	assert.Equal(t, "INFO\ttest\twriter msg\n", string(output))
}

func Test_GinColorWriter(t *testing.T) {
	var output []byte
	err := zap.RegisterSink("test4", testSinkFactoryFactory(&output))
	require.NoError(t, err)

	color.NoColor = false

	v := viper.New()
	v.Set("output-paths", "test4:")
	v.Set("err-output-paths", "test4:")
	v.Set("format", "bare")
	v.Set("level", "info")

	err = Init(v, nil)
	require.NoError(t, err)

	logger := Named("test")
	writer := &zapio.Writer{Log: logger.Desugar(), Level: zap.InfoLevel}
	ginWriter := NewGinColorWriter(writer)

	_, err = ginWriter.Write([]byte("normal message"))
	require.NoError(t, err)

	_, err = ginWriter.Write([]byte("[WARNING] message"))
	require.NoError(t, err)

	_, err = ginWriter.Write([]byte("[ERROR] message"))
	require.NoError(t, err)

	require.NoError(t, writer.Sync())

	expected := "\x1b[32mINFO\x1b[0m\ttest\tnormal message\n\x1b[32mINFO\x1b[0m\ttest\t\x1b[33m[WARNING]\x1b[0m message\n\x1b[32mINFO\x1b[0m\ttest\t\x1b[31m[ERROR]\x1b[0m message\n"

	assert.Equal(t, expected, string(output))
}

func Test_Color(t *testing.T) {
	var enc TestArrayEncoder

	color.NoColor = false

	ColorCapitalLevelEncoder(zap.InfoLevel, &enc)
	require.Len(t, enc.Elems, 1)

	expected := []byte("\x1b[32mINFO\x1b[0m")
	actual := []byte(enc.Elems[0].(string))
	assert.Equal(t, expected, actual)
}

// Adapted from zap internal implementation.
type TestArrayEncoder struct {
	Elems []interface{}
}

func (s *TestArrayEncoder) AppendReflected(v interface{}) error {
	s.Elems = append(s.Elems, v)
	return nil
}

func (s *TestArrayEncoder) AppendArray(v zapcore.ArrayMarshaler) error   { return nil }
func (s *TestArrayEncoder) AppendObject(v zapcore.ObjectMarshaler) error { return nil }

func (s *TestArrayEncoder) AppendBool(v bool)              { s.Elems = append(s.Elems, v) }
func (s *TestArrayEncoder) AppendByteString(v []byte)      { s.Elems = append(s.Elems, string(v)) }
func (s *TestArrayEncoder) AppendComplex128(v complex128)  { s.Elems = append(s.Elems, v) }
func (s *TestArrayEncoder) AppendComplex64(v complex64)    { s.Elems = append(s.Elems, v) }
func (s *TestArrayEncoder) AppendDuration(v time.Duration) { s.Elems = append(s.Elems, v) }
func (s *TestArrayEncoder) AppendFloat64(v float64)        { s.Elems = append(s.Elems, v) }
func (s *TestArrayEncoder) AppendFloat32(v float32)        { s.Elems = append(s.Elems, v) }
func (s *TestArrayEncoder) AppendInt(v int)                { s.Elems = append(s.Elems, v) }
func (s *TestArrayEncoder) AppendInt64(v int64)            { s.Elems = append(s.Elems, v) }
func (s *TestArrayEncoder) AppendInt32(v int32)            { s.Elems = append(s.Elems, v) }
func (s *TestArrayEncoder) AppendInt16(v int16)            { s.Elems = append(s.Elems, v) }
func (s *TestArrayEncoder) AppendInt8(v int8)              { s.Elems = append(s.Elems, v) }
func (s *TestArrayEncoder) AppendString(v string)          { s.Elems = append(s.Elems, v) }
func (s *TestArrayEncoder) AppendTime(v time.Time)         { s.Elems = append(s.Elems, v) }
func (s *TestArrayEncoder) AppendUint(v uint)              { s.Elems = append(s.Elems, v) }
func (s *TestArrayEncoder) AppendUint64(v uint64)          { s.Elems = append(s.Elems, v) }
func (s *TestArrayEncoder) AppendUint32(v uint32)          { s.Elems = append(s.Elems, v) }
func (s *TestArrayEncoder) AppendUint16(v uint16)          { s.Elems = append(s.Elems, v) }
func (s *TestArrayEncoder) AppendUint8(v uint8)            { s.Elems = append(s.Elems, v) }
func (s *TestArrayEncoder) AppendUintptr(v uintptr)        { s.Elems = append(s.Elems, v) }

func Test_resolveTemplates(t *testing.T) {
	templates := []string{
		"hello, {{ .name }}!",
		"verbatum text",
		"created by {{ .unspecified }}",
	}

	vals := map[string]interface{}{
		"name":   "world",
		"unused": "whatever",
	}

	err := resolveTemplates(&templates, vals)
	require.NoError(t, err)

	assert.Equal(t, "hello, world!", templates[0])
	assert.Equal(t, "verbatum text", templates[1])
	assert.Equal(t, "created by <no value>", templates[2])
}
