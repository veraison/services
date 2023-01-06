// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package log

import (
	"github.com/gin-gonic/gin"
	ahocorasick "github.com/petar-dambovaliev/aho-corasick"
	"go.uber.org/zap"
	"go.uber.org/zap/zapio"
)

// InitGinWriter sets up gin DefaultWriter to use Veraison logging.
func InitGinWriter() {
	zapWriter := &zapio.Writer{Log: Named("gin").Desugar(), Level: zap.InfoLevel}
	gin.DefaultWriter = NewGinColorWriter(zapWriter)
}

// GinColorWriter wraps zapio.Writer to provide warning and error highlighting
// inside gin traces.
type GinColorWriter struct {
	matcher ahocorasick.AhoCorasick
	writer  *zapio.Writer
}

func NewGinColorWriter(writer *zapio.Writer) *GinColorWriter {
	builder := ahocorasick.NewAhoCorasickBuilder(ahocorasick.Opts{
		AsciiCaseInsensitive: false,
		MatchOnlyWholeWords:  true,
		MatchKind:            ahocorasick.LeftMostLongestMatch,
		DFA:                  true,
	})

	matcher := builder.Build([]string{"[WARNING]", "[ERROR]"})

	return &GinColorWriter{matcher: matcher, writer: writer}
}

func (o *GinColorWriter) Write(p []byte) (int, error) {
	line := string(p)

	matches := o.matcher.FindAll(line)

	for i := len(matches) - 1; i >= 0; i -= 1 {
		m := matches[i]

		var colorizer func(a ...interface{}) string
		text := line[m.Start():m.End()]
		switch text {
		case "[WARNING]":
			colorizer = levelColorizers[zap.WarnLevel]
		case "[ERROR]":
			colorizer = levelColorizers[zap.ErrorLevel]
		default:
			DPanicw("unexpected gin log tag", "value", text)
		}

		line = line[:m.Start()] + colorizer(text) + line[m.End():]
	}

	return o.writer.Write([]byte(line + "\n"))
}
