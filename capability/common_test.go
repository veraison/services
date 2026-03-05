// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package capability_test

import (
	"testing"
	"time"

	"github.com/veraison/services/capability"
	"go.uber.org/zap"
)

func TestParseCacheMaxAge(t *testing.T) {
	tests := []struct {
		name   string
		s      string
		dflt   time.Duration
		logger *zap.SugaredLogger
		want   time.Duration
	}{
		{
			name:   "empty string returns default",
			s:      "",
			dflt:   10 * time.Second,
			logger: zap.NewNop().Sugar(),
			want:   10 * time.Second,
		},
		{
			name:   "valid duration string is parsed correctly",
			s:      "1h30m",
			dflt:   10 * time.Second,
			logger: zap.NewNop().Sugar(),
			want:   90 * time.Minute,
		},
		{
			name:   "invalid duration string returns default",
			s:      "invalid",
			dflt:   10 * time.Second,
			logger: zap.NewNop().Sugar(),
			want:   10 * time.Second,
		},
		{
			name:   "negative duration string returns default",
			s:      "-1h",
			dflt:   10 * time.Second,
			logger: zap.NewNop().Sugar(),
			want:   10 * time.Second,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := capability.ParseCacheMaxAge(tt.s, tt.dflt, tt.logger)
			if got != tt.want {
				t.Errorf("ParseCacheMaxAge() = %v, want %v", got, tt.want)
			}
		})
	}
}
