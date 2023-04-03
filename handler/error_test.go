// Copyright 2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_BadEvidenceError_marshalling_roundtrip(t *testing.T) {
	type TestVector struct {
		Title string
		Bee   BadEvidenceError
	}
	tvs := []TestVector{
		{"string-unexpected", BadEvidenceError{"error"}},
		{"error-unexpected", BadEvidenceError{errors.New("error")}},
		{"wrapped-unexpected", BadEvidenceError{
			fmt.Errorf("wrapped: %w", errors.New("error")),
		}},
		{"int-unexpected", BadEvidenceError{42}},
		{"string-crypto", BadEvidenceError{"wrong key"}},
	}

	for _, tv := range tvs {
		fmt.Println(tv.Title)

		data, err := json.Marshal(tv.Bee)
		require.NoError(t, err)

		var decoded BadEvidenceError
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, decoded.Error(), tv.Bee.Error())
	}
}

func Test_BadEvidenceError_parse(t *testing.T) {
	expected := BadEvidenceError{"something went wrong"}

	input := errors.New("bad evidence: something went wrong")

	err := ParseError(input)

	assert.True(t, errors.Is(err, expected))
	assert.Equal(t, expected.Error(), err.Error())

	input2 := errors.New(expected.Error())

	err = ParseError(input2)

	assert.True(t, errors.Is(err, expected))
	assert.Equal(t, expected.Error(), err.Error())
}

func Test_BadEvidenceError_is(t *testing.T) {
	bee := BadEvidence("error")

	assert.True(t, errors.Is(bee, BadEvidenceError{}))
	assert.True(t, errors.Is(BadEvidenceError{}, bee))

	err := fmt.Errorf("caught: %w", bee)
	assert.True(t, errors.Is(err, BadEvidenceError{}))
	assert.True(t, errors.Is(BadEvidenceError{}, err))
}

func Test_BadEvidenceError_wrapping(t *testing.T) {
	err1 := errors.New("error 1")
	err2 := fmt.Errorf("got error: %w", err1)
	bee := BadEvidence(err2)

	out := ParseError(errors.New(bee.Error()))
	assert.True(t, errors.Is(out, bee))
	assert.True(t, errors.Is(out, err2))
	assert.True(t, errors.Is(out, err1))
}
