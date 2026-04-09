// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package plugin

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"math"

	"github.com/spf13/viper"
)

var (
	ErrNotSet = errors.New("parameter not set")
	ErrInvalid = errors.New("invalid parameter value")
)

// ParametersMapFromViper constructs a map[string]*Parameters from the
// provided *viper.Viper, where the top-level entries in the Viper become
// map keys and the corresponding nested entries are used to construct
// the Parameters.
// If keyTrans function is not nil, then it is called for each top-level
// Viper entry, and the output is used as map keys, rather than using the
// entry directly.
//
// For example:
//
//   v := viper.New()
//   v.Set("s1.p1", "foo")
//   v.Set("s2.p1", 1)
//   m1, err := ParametersMapFromViper(v, func(n string) string { return strings.ToUpper(n) })
//
//   m2 := map[string]*Parameters{
//       "S1": NewParameters().SetString("p1", "foo"),
//       "S2": NewParameters().SetInt("p1", 1),
//   }
//
// m1 and m2 above are equivalent.
func ParametersMapFromViper(v *viper.Viper, keyTrans func(string)string) (map[string]*Parameters, error) {
	ret := make(map[string]*Parameters)
	if v == nil {
		return ret, nil
	}

	for key := range v.AllSettings() {
		if sub := v.Sub(key); sub != nil {
			params, err := ParametersFromViper(sub)
			if err != nil {
				return nil, fmt.Errorf("entry %q: %w", key, err)
			}

			if keyTrans != nil {
				key = keyTrans(key)
			}

			ret[key] = params
		} else {
			return nil, fmt.Errorf("entry %q is not a map", key)
		}
	}

	return ret, nil
}

// ParametersFromJSON parses the provided JSON data and uses it to generate *Parameters.
func ParametersFromJSON(data []byte) (*Parameters, error) {
	ret := NewParameters()

	if err := ret.UnmarshalJSON(data); err != nil {
		return nil, err
	}

	return ret, nil
}

// ParametersFromViper generates a *Parameters from the provided *viper.Viper
func ParametersFromViper(v *viper.Viper) (*Parameters, error) {
	return ParametersFromMap(v.AllSettings())
}

// ParametersFromMap generates a *Parameters from the provided map[string]any
func ParametersFromMap(m map[string]any) (*Parameters, error) {
	ret := NewParameters()
	if err := ret.PopulateFromMap(m); err != nil {
		return nil, err
	}

	return ret, nil
}

// Parameters represent plugin configuration settings that are passed to a
// plugin on its initialization. These are, essentially, string keys mapping on
// values of supported types (string, []byte, int, int64, bool).
// Valid keys and their expected value types are plugin-specific.
type Parameters struct {
	values map[string]any
}

// NewParameters creates a new empty *Parameters.
func NewParameters() *Parameters {
	return &Parameters{
		values: make(map[string]any),
	}
}

// MarshalJSON serializes Parameters into a JSON object
func (o *Parameters) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.values)
}

// UnmarshalJSON deserializes a JSON object and uses it to populate the
// Parameters.
func (o *Parameters) UnmarshalJSON(data []byte) error {
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	// use temporary intermediate so that o's original state is preserved
	// on error
	tmp := NewParameters()
	for k, v := range m {
		if err := tmp.Set(k, v); err != nil {
			return err
		}
	}

	o.values = tmp.values
	return nil
}

// SetString sets the specified key to the specified string value and returns a
// pointer to the Parameters object. If the key already exists, it is
// overwritten.
func (o *Parameters) SetString(key, value string) *Parameters {
	return o.set(key, value)
}

// SetBytes sets the specified key to the specified []byte value and returns a
// pointer to the Parameters object. If the key already exists, it is
// overwritten.
func (o *Parameters) SetBytes(key string, value []byte) *Parameters {
	return o.set(key, base64.StdEncoding.EncodeToString(value))
}

// SetInt sets the specified key to the specified int value and returns a
// pointer to the Parameters object. If the key already exists, it is
// overwritten.
func (o *Parameters) SetInt(key string, value int) *Parameters {
	return o.set(key, int64(value))
}

// SetInt64 sets the specified key to the specified int64 value and returns a
// pointer to the Parameters object. If the key already exists, it is
// overwritten.
func (o *Parameters) SetInt64(key string, value int64) *Parameters {
	return o.set(key, value)
}

// SetBool sets the specified key to the specified bool value and returns a
// pointer to the Parameters object. If the key already exists, it is
// overwritten.
func (o *Parameters) SetBool(key string, value bool) *Parameters {
	return o.set(key, value)
}

// Set the specified key to the specified value, provided value is of one of
// the supported types, otherwise ErrInvalid is returned. Supported types are
// string, []byte, int, int64, and bool. If the key already exists, it is
// overwritten.
func (o *Parameters) Set(key string, value any) error {
	switch t := value.(type) {
	case string, []byte, int64, bool:
		o.set(key, value)
	case int:
		o.set(key, int64(t))
	case float64:
		if t < float64(math.MinInt64) || t > float64(math.MaxInt64) {
			return fmt.Errorf("%w for %q: out of int64 range: %v", ErrInvalid, key, t)
		}

		if t != float64(int64(t)) {
			return fmt.Errorf("%w for %q: has fractional part: %v", ErrInvalid, key, t)
		}

		o.set(key, int64(t))
	default:
		return fmt.Errorf("%w for %q: %v (%T)", ErrInvalid, key, value, value)
	}

	return nil
}

// PopulateFromMap populates the Parameters from the provided map. This is
// equivalent to calling Set for each key and corresponding value in the map.
func (o *Parameters) PopulateFromMap(m map[string]any) error {
	for k, v := range m {
		if err := o.Set(k, v); err != nil {
			return err
		}
	}

	return nil
}

// PopulateFromViper populates the Parameters form the provided *viper.Viper.
// This is equivalent to first converting the Viper into a map[string]any and
// passing it to PopulateFromMap().
func (o *Parameters) PopulateFromViper(v *viper.Viper) error {
	return o.PopulateFromMap(v.AllSettings())
}

// Map returns a map[string]any corresponding to the Parameters.
func (o *Parameters) Map() map[string]any {
	return maps.Clone(o.values)
}

// Get returns the value corresponding to the specified key, or ErrNotSet if
// the key is not in Parameters.
func (o *Parameters) Get(key string) (any, error) {
	val, ok := o.values[key]
	if ok {
		return val, nil
	}

	return nil, fmt.Errorf("%w: %s", ErrNotSet, key)
}

// DefaultGet returns the value corresponding to the specified key, or the
// provided defaultValue, if the key is not in Parameters.
func (o *Parameters) DefaultGet(key string, defaultValue any) any {
	val, err := o.Get(key)
	if err != nil {
		return defaultValue
	}

	return val
}

// GetString returns the string value corresponding to the specified key. If
// the key is not set or the value is a string, return "" and set the error to
// ErrNotSet or ErrInvalid respectively.
func (o *Parameters) GetString(key string) (string, error) {
	return get(o, key, "", false)
}

// MustGetString is like GetString() but panics on error.
func (o *Parameters) MustGetString(key string) string {
	val, err := o.GetString(key)
	if err != nil {
		panic(err)
	}

	return val
}

// DefaultGetString returns the string value corresponding to the specified
// key. If the key is not set, defaultValue is returned instead. If the value
// is not a string, returns "" and the error is set to ErrInvalid.
func (o *Parameters) DefaultGetString(key, defaultValue string) (string, error) {
	return get(o, key, defaultValue, true)
}

// GetBytes returns the []byte value corresponding to the specified key. If the
// key is not set or the value is a []byte, return nil and set the error to
// ErrNotSet or ErrInvalid respectively.
func (o *Parameters) GetBytes(key string) ([]byte, error) {
	encoded, err := o.GetString(key)
	if err != nil {
		return nil, err
	}

	val, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("%w for %q: %q", ErrInvalid, key, encoded)
	}

	return val, nil
}

// MustGetBytes is like GetBytes() but panics on error.
func (o *Parameters) MustGetBytes(key string) []byte {
	val, err := o.GetBytes(key)
	if err != nil {
		panic(err)
	}

	return val
}

// DefaultGetBytes returns the []byte value corresponding to the specified key.
// If the key is not set, defaultValue is returned instead. If the value is not
// a []byte, returns nil and the error is set to ErrInvalid.
func (o *Parameters) DefaultGetBytes(key string, defaultValue []byte) ([]byte, error) {
	encoded, err := o.DefaultGetString(key, "")
	if err != nil {
		return nil, err
	} else if encoded == "" {
		return defaultValue, nil
	}

	val, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("%w for %q: %q", ErrInvalid, key, encoded)
	}

	return val, nil
}

// GetInt64 returns the int64 value corresponding to the specified key. If the
// key is not set or the value is a int64, return 0 and set the error to
// ErrNotSet or ErrInvalid respectively.
func (o *Parameters) GetInt64(key string) (int64, error) {
	return get(o, key, int64(0), false)
}

// MustGetInt64 is like GetInt64() but panics on error.
func (o *Parameters) MustGetInt64(key string) int64 {
	val, err := o.GetInt64(key)
	if err != nil {
		panic(err)
	}

	return val
}


// DefaultGetInt64 returns the int64 value corresponding to the specified key.
// If the key is not set, defaultValue is returned instead. If the value is not
// a int64, returns 0 and the error is set to ErrInvalid.
func (o *Parameters) DefaultGetInt64(key string, defaultValue int64) (int64, error) {
	return get(o, key, defaultValue, true)
}

// GetInt returns the int value corresponding to the specified key. If the key
// is not set or the value is a int, return 0 and set the error to ErrNotSet or
// ErrInvalid respectively.
func (o *Parameters) GetInt(key string) (int, error) {
	val, err := o.GetInt64(key)
	return int(val), err
}

// MustGetInt is like GetInt() but panics on error.
func (o *Parameters) MustGetInt(key string) int {
	val, err := o.GetInt(key)
	if err != nil {
		panic(err)
	}

	return val
}

// DefaultGetInt returns the int value corresponding to the specified key.
// If the key is not set, defaultValue is returned instead. If the value is not
// a int, returns 0 and the error is set to ErrInvalid.
func (o *Parameters) DefaultGetInt(key string, defaultValue int) (int, error) {
	val, err := o.DefaultGetInt64(key, int64(defaultValue))
	return int(val), err
}

// GetBool returns the bool value corresponding to the specified key. If the
// key is not set or the value is a bool, return false and set the error to
// ErrNotSet or ErrInvalid respectively.
func (o *Parameters) GetBool(key string) (bool, error) {
	return get(o, key, false, false)
}

// MustGetBool is like GetBool() but panics on error.
func (o *Parameters) MustGetBool(key string) bool {
	val, err := o.GetBool(key)
	if err != nil {
		panic(err)
	}

	return val
}

// DefaultGetBool returns the bool value corresponding to the specified key.
// If the key is not set, defaultValue is returned instead. If the value is not
// a bool, returns false and the error is set to ErrInvalid.
func (o *Parameters) DefaultGetBool(key string, defaultValue bool) (bool, error) {
	return get(o, key, defaultValue, true)
}

func (o *Parameters) set(key string, value any) *Parameters {
	o.values[key] = value
	return o
}

func get[T any](params *Parameters, key string, defaultValue T, withDefault bool) (T, error) {
	val, err := params.Get(key)
	if err != nil {
		if withDefault {
			return defaultValue, nil
		} else {
			return defaultValue, err
		}
	}

	switch t := val.(type) {
	case *T:
		return *t, nil
	case T:
		return t, nil
	default:
		return defaultValue, fmt.Errorf("%w for %q: %v", ErrInvalid, key, val)
	}
}
