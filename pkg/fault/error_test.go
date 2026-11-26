// Copyright 2024-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package fault

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError_round_trip(t *testing.T) {
	inError := BadEvidence("%s %d", "foo", 1)
	outError := ParseError(SerializeError(inError))
	assert.True(t, errors.Is(outError, ErrBadEvidence))
	assert.True(t, errors.Is(outError, inError))
}
