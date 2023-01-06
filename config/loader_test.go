// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package config

import (
	"errors"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type person struct {
	Name       string
	Occupation string
	Age        int
}

func (o person) Validate() error {
	if o.Age < 0 {
		return errors.New("age cannot be negative")
	}

	return nil
}

func Test_Loader_bad_init(t *testing.T) {
	var loader Loader

	err := loader.Init(7)
	assert.ErrorContains(t, err, "expected pointer to a struct but got int")

	var p person
	err = loader.Init(p)
	assert.ErrorContains(t, err, "expected pointer to a struct but got struct")
}

func Test_Loader_mapFromString(t *testing.T) {
	leela := person{
		Name: "Turanga leela",
	}

	expected := map[string]interface{}{
		"name": "Turanga leela",
	}

	var res person
	loader := NewLoader(&res)
	require.NotNil(t, loader)

	m, err := loader.mapFromStruct(leela)
	require.NoError(t, err)
	assert.Equal(t, expected, m)

	m, err = loader.mapFromStruct(&leela)
	require.NoError(t, err)
	assert.Equal(t, expected, m)
}

func Test_Loader_load_from_map_ok(t *testing.T) {
	var zoidberg person

	loader := NewLoader(&zoidberg)
	require.NotNil(t, loader)

	err := loader.LoadFromMap(map[string]interface{}{
		"name":       "John A. Zoidberg, MD.",
		"occupation": "doctor",
		"age":        86,
	})
	require.NoError(t, err)

	assert.Equal(t, "John A. Zoidberg, MD.", zoidberg.Name)
	assert.Equal(t, "doctor", zoidberg.Occupation)
	assert.Equal(t, 86, zoidberg.Age)
}

func Test_Loader_load_from_viper_ok(t *testing.T) {
	fry := person{
		Name:       "Philip J. Fry",
		Occupation: "Pizza Delivery Boy",
		Age:        25,
	}

	loader := NewLoader(&fry)

	v := viper.New()

	v.Set("occupation", "Intergalactic Delivery Boy")
	v.Set("age", 1025)

	err := loader.LoadFromViper(v)
	require.NoError(t, err)

	assert.Equal(t, "Philip J. Fry", fry.Name)
	assert.Equal(t, "Intergalactic Delivery Boy", fry.Occupation)
	assert.Equal(t, 1025, fry.Age)
}

func Test_Loader_load_from_viper_nil(t *testing.T) {
	var person person

	loader := NewLoader(&person)

	err := loader.LoadFromViper(nil)
	assert.ErrorContains(t, err, "nil configuration")
}

func Test_Loader_extra_value(t *testing.T) {
	input := map[string]interface{}{
		"name":       "Turanga Leela",
		"occupation": "captain",
		"age":        24,
		"numEyes":    1,
	}

	res := person{}
	loader := NewLoader(&res)
	require.NotNil(t, loader)

	err := loader.LoadFromMap(input)
	assert.ErrorContains(t, err, "unexpected directives: numEyes")

	neloader := NewNonExclusiveLoader(&res)
	err = neloader.LoadFromMap(input)
	assert.NoError(t, err)
}

func Test_Loader_missing_value(t *testing.T) {
	input := map[string]interface{}{
		"name": "Turanga Leela",
		"age":  24,
	}

	res := person{}
	loader := NewLoader(&res)
	require.NotNil(t, loader)

	err := loader.LoadFromMap(input)
	assert.ErrorContains(t, err, "directives not found: Occupation")
}

func Test_Loader_invalid_value(t *testing.T) {
	input := map[string]interface{}{
		"name":       "Philip J. Fry",
		"occupation": "own grandfather",
		"age":        -28,
	}

	res := person{}
	loader := NewLoader(&res)
	require.NotNil(t, loader)

	err := loader.LoadFromMap(input)
	assert.ErrorContains(t, err, "age cannot be negative")
}

func Test_Loader_zero_values(t *testing.T) {
	res := person{
		Name:       "Philip J. Fry",
		Occupation: "age-reducing tar victim",
		Age:        0,
	}
	loader := NewLoader(&res)
	require.NotNil(t, loader)

	// Default zero value is semantically identical to an unset value.
	err := loader.LoadFromMap(map[string]interface{}{})
	assert.ErrorContains(t, err, "directives not found: Age")

	// Zero values in the input are ok, however.
	err = loader.LoadFromMap(map[string]interface{}{"age": 0})
	assert.NoError(t, err)

	// Setting `config:"zerodefault"` tag on a field propagates zero-value defaults.
	ageless := struct {
		Name       string
		Occupation string `config:"zerodefault"`
		Age        int    `config:"zerodefault"`
	}{Name: "Philip J. Fry", Occupation: "", Age: 0}

	loader = NewLoader(&ageless)
	require.NotNil(t, loader)
	err = loader.LoadFromMap(map[string]interface{}{})
	assert.NoError(t, err)
}

func Test_Loader_renamed_field(t *testing.T) {
	leela := struct {
		Name         string
		Species      string
		NumberOfEyes int `mapstructure:"number-of-eyes"`
	}{}

	loader := NewLoader(&leela)
	require.NotNil(t, loader)

	input := map[string]interface{}{
		"name":           "Turanga Leela",
		"species":        "cyclops",
		"number-of-eyes": 1,
	}

	err := loader.LoadFromMap(input)
	require.NoError(t, err)
	assert.Equal(t, 1, leela.NumberOfEyes)
}

func Test_Loader_valid_tags(t *testing.T) {
	cfg := struct {
		ServerAddress string `valid:"dialstring" mapstructure:"addr"`
	}{}

	loader := NewLoader(&cfg)

	testCases := map[string]string{
		"localhost:8080":        "",
		"127.0.0.1:1234":        "",
		"742 Evergreen Terrace": "ServerAddress: 742 Evergreen Terrace does not validate as dialstring",
	}

	for addr, expected := range testCases {
		input := map[string]interface{}{
			"addr": addr,
		}

		err := loader.LoadFromMap(input)

		if expected == "" {
			assert.NoError(t, err)
		} else {
			assert.ErrorContains(t, err, expected)
		}
	}

}

func Test_Loader_load_from_viper_env_ok(t *testing.T) {
	fry := person{
		Name:       "Philip J. Fry",
		Occupation: "Pizza Delivery Boy",
		Age:        25,
	}

	loader := NewLoader(&fry)

	v := viper.New()

	v.Set("occupation", "Intergalactic Delivery Boy")
	v.Set("age", 1025)

	err := loader.LoadFromViper(v)
	require.NoError(t, err)

	assert.Equal(t, "Philip J. Fry", fry.Name)
	assert.Equal(t, "Intergalactic Delivery Boy", fry.Occupation)
	assert.Equal(t, 1025, fry.Age)
}
