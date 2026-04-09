// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package plugin

import (
	"fmt"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)


func TestParameters_round_trip(t *testing.T) {
	params := NewParameters().
			SetString("p1", "foo").
			SetInt("p2", 1).
			SetInt64("p3", 2).
			SetBytes("p4", []byte{0xde, 0xad, 0xbe, 0xef}).
			SetBool("p5", true)

	data, err := params.MarshalJSON()
	assert.NoError(t, err)

	var out Parameters
	err = out.UnmarshalJSON(data)
	assert.NoError(t, err)

	assert.Equal(t, params.MustGetString("p1"), out.MustGetString("p1"))
	assert.Equal(t, params.MustGetInt("p2"), out.MustGetInt("p2"))
	assert.Equal(t, params.MustGetInt64("p3"), out.MustGetInt64("p3"))
	assert.Equal(t, params.MustGetBytes("p4"), out.MustGetBytes("p4"))
	assert.Equal(t, params.MustGetBool("p5"), out.MustGetBool("p5"))
}

func TestParameters_bytes(t *testing.T) {
	bytes := []byte{0xde, 0xad, 0xbe, 0xef}
	params := NewParameters().SetBytes("p1", bytes)
	assert.Equal(t, bytes, params.MustGetBytes("p1"))
}

func TestParameters_float(t *testing.T) {
	params := NewParameters()

	err := params.Set("p1", 1.0)
	assert.NoError(t, err)
	assert.Equal(t, 1, params.MustGetInt("p1"))

	err = params.Set("p2", 1.1)
	assert.ErrorIs(t, err, ErrInvalid)

	err = params.Set("p2", 9223372036854775808.0)
	assert.ErrorIs(t, err, ErrInvalid)
}

func TestParameters_set_get(t *testing.T) {
	bytes := []byte{0xde, 0xad, 0xbe, 0xef}
	params := NewParameters().SetInt("p1", 1).SetBytes("p2", bytes)

	val, err := params.GetInt("p1")
	assert.NoError(t, err)
	assert.Equal(t, 1, val)

	_, err = params.GetString("p1")
	assert.ErrorIs(t, err, ErrInvalid)

	_, err = params.GetString("p3")
	assert.ErrorIs(t, err, ErrNotSet)

	val, err = params.DefaultGetInt("p1", 2)
	assert.NoError(t, err)
	assert.Equal(t, 1, val)

	otherBytes := []byte{0x0b, 0xad, 0xf0, 0x0d}
	val2, err := params.DefaultGetBytes("p2", otherBytes)
	assert.NoError(t, err)
	assert.Equal(t, bytes, val2)

	val2, err = params.DefaultGetBytes("p3", otherBytes)
	assert.NoError(t, err)
	assert.Equal(t, otherBytes, val2)

	_, err = params.DefaultGetInt("p2", 2)
	assert.ErrorIs(t, err, ErrInvalid)

	val3 := params.DefaultGet("p1", "foo")
	assert.Equal(t, int64(1), val3)
}

func TestParametersFromMap(t *testing.T) {
	m := map[string]any{
		"p1": "foo",
		"p2": int64(2),
	}
	params, err := ParametersFromMap(m)
	assert.NoError(t, err)
	assert.Equal(t, "foo", params.MustGetString("p1"))
	assert.Equal(t, 2, params.MustGetInt("p2"))
	assert.Equal(t, m, params.Map())

	_, err = params.GetBool("p3")
	assert.ErrorIs(t, err, ErrNotSet)

	_, err = ParametersFromMap(map[string]any{
		"p1": []string{"foo", "bar"},
	})
	assert.ErrorIs(t, err, ErrInvalid)
}

func TestParametersFromViper(t *testing.T) {
	v := viper.New()
	v.Set("p1", "foo")
	v.Set("p2", 2)

	params, err := ParametersFromViper(v)
	assert.NoError(t, err)
	assert.Equal(t, "foo", params.MustGetString("p1"))
	assert.Equal(t, 2, params.MustGetInt("p2"))

	_, err = params.GetBool("p3")
	assert.ErrorIs(t, err, ErrNotSet)

	v.Set("p3.a", true)
	_, err = ParametersFromViper(v)
	assert.ErrorIs(t, err, ErrInvalid)

	err = NewParameters().PopulateFromViper(v)
	assert.ErrorIs(t, err, ErrInvalid)
}

func TestParametersFromJSON(t *testing.T) {
	bytes := []byte(`{"p1":"foo","p2":2}`)
	params, err := ParametersFromJSON(bytes)
	assert.NoError(t, err)
	assert.Equal(t, "foo", params.MustGetString("p1"))
	assert.Equal(t, 2, params.MustGetInt("p2"))

	_, err = params.GetBool("p3")
	assert.ErrorIs(t, err, ErrNotSet)

	bytes = []byte(`{"p1":["foo","bar"]}`)
	err = params.UnmarshalJSON(bytes)
	assert.ErrorIs(t, err, ErrInvalid)
	assert.Equal(t, "foo", params.MustGetString("p1"))
}

func TestPametersMapFromViper(t *testing.T) {
	v := viper.New()
	v.Set("s1.p1", "foo")
	v.Set("s1.p2", 2)
	v.Set("s2.p1", "bar")
	v.Set("s2.p2", "baz")

	m, err := ParametersMapFromViper(v, nil)
	assert.NoError(t, err)
	assert.Equal(t, "foo", m["s1"].MustGetString("p1"))
	assert.Equal(t, 2, m["s1"].MustGetInt("p2"))
	assert.Equal(t, "bar", m["s2"].MustGetString("p1"))
	assert.Equal(t, "baz", m["s2"].MustGetString("p2"))

	_, ok := m["p3"]
	assert.False(t, ok)

	m, err = ParametersMapFromViper(v, func(name string) string {
		return fmt.Sprintf("%s-suffix", name)
	})
	assert.NoError(t, err)
	assert.Equal(t, "foo", m["s1-suffix"].MustGetString("p1"))

	_, ok = m["p1"]
	assert.False(t, ok)

	v.Set("s3.p1.a", 1)
	_, err = ParametersMapFromViper(v, nil)
	assert.ErrorIs(t, err, ErrInvalid)
}
