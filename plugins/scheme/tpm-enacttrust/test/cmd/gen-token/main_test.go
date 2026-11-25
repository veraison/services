// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"bytes"
	"testing"
)

func TestDecodeNonce(t *testing.T) {
	want := []byte{0xfb, 0xff}

	tests := map[string]string{
		"padded standard":   "+/8=",
		"unpadded standard": "+/8",
		"padded URL-safe":   "-_8=",
		"unpadded URL-safe": "-_8",
	}

	for name, encoded := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := decodeNonce(encoded)
			if err != nil {
				t.Fatalf("decodeNonce(%q) returned an error: %v", encoded, err)
			}
			if !bytes.Equal(got, want) {
				t.Errorf("decodeNonce(%q) = %x, want %x", encoded, got, want)
			}
		})
	}
}

func TestDecodeNonceRejectsInvalidEncoding(t *testing.T) {
	if _, err := decodeNonce("not base64!"); err == nil {
		t.Fatal("decodeNonce() accepted invalid base64")
	}
}
