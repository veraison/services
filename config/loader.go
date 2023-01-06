// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package config

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"golang.org/x/text/cases"
)

var ErrNilConfig = errors.New("nil configuration")

// IValidatable defines an interface of objects that can self-validate.
type IValidatable interface {
	Validate() error
}

// Loader is responsible for loading config from a source into a struct instance.
type Loader struct {
	config    *mapstructure.DecoderConfig
	exclusive bool
}

// NewLoader creates a new loader. dest must be a pointer to a struct instance
// to be populated, other wise, nil returned.
func NewLoader(dest interface{}) *Loader {
	loader := &Loader{exclusive: true}

	if err := loader.Init(dest); err != nil {
		panic(err)
	}

	return loader
}

// NewNonExclusiveLoader is just like NewLoader, but the loader returned allows
// ther to be unknown settings in the source.
func NewNonExclusiveLoader(dest interface{}) *Loader {
	loader := &Loader{exclusive: false}

	if err := loader.Init(dest); err != nil {
		panic(err)
	}
	return loader
}

// Init initializes the loader with the specified destination. If dest is not
// a pointer to a struct instance, an error is returned.
func (o *Loader) Init(dest interface{}) error {
	if dest == nil {
		return errors.New("cannot initialize loader with nil dest")
	}

	val := reflect.ValueOf(dest)

	if val.Kind() != reflect.Ptr || reflect.Indirect(val).Kind() != reflect.Struct {
		return fmt.Errorf("expected pointer to a struct but got %v", val.Kind())
	}

	o.config = &mapstructure.DecoderConfig{
		ErrorUnused:      o.exclusive,
		ErrorUnset:       true,
		WeaklyTypedInput: true,
		Metadata:         &mapstructure.Metadata{},
		Result:           dest,
	}

	return nil
}

// LoadFromMap populates the destination struct instance with values from the
// specified map[string]interface{} source.
func (o Loader) LoadFromMap(source map[string]interface{}) error {
	// This must be done before setDefaults() below because NewDecoder()
	// sets the default TagName  and MatchName in the config!
	decoder, err := mapstructure.NewDecoder(o.config)
	if err != nil {
		return err
	}

	// Use existing non-Zero values as defaults, by setting them in the
	// input and then overwriting them with those from the source.
	input, err := o.mapFromStruct(o.config.Result)
	if err != nil {
		return err
	}

	for key, val := range source {
		input[key] = val
	}

	if err = decoder.Decode(input); err != nil {
		msErr, ok := err.(*mapstructure.Error)
		if !ok {
			return err
		}

		var messageParts []string

		for _, subError := range msErr.Errors {
			parts := strings.Split(subError, "has invalid keys: ")
			if len(parts) > 1 {
				subMsg := fmt.Sprintf("unexpected directives: %s", parts[1])
				messageParts = append(messageParts, subMsg)
				continue
			}

			parts = strings.Split(subError, "has unset fields: ")
			if len(parts) > 1 {
				subMsg := fmt.Sprintf("directives not found: %s", parts[1])
				messageParts = append(messageParts, subMsg)
				continue
			}

			messageParts = append(messageParts, subError)
		}

		return errors.New(strings.Join(messageParts, "; "))
	}

	// Validate struct field formats based on their `valid` tags.
	if ok, err := govalidator.ValidateStruct(o.config.Result); !ok {
		return err
	}

	// Endorce arbitrary validation defined by the struct itself.
	if v, ok := o.config.Result.(IValidatable); ok {
		err = v.Validate()
	}

	return err
}

// LoadFromViper populates the destination struct instance with values from the
// specified *viper.Viper source.
func (o Loader) LoadFromViper(source *viper.Viper) error {
	if source == nil {
		return ErrNilConfig
	}
	return o.LoadFromMap(source.AllSettings())
}

func (o Loader) mapFromStruct(s interface{}) (map[string]interface{}, error) {
	val := reflect.Indirect(reflect.ValueOf(s))
	typ := val.Type()

	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct but got %s", typ.Name())
	}

	result := map[string]interface{}{}

	// This assumes that o.config.MatchName is not explicitly
	// set, and so defaults to strings.EqualFold.
	caser := cases.Fold()

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)

		// Treat Zero values as unset, unless "zerodefault" is set for the "config" tag.
		if fieldVal.IsZero() {
			allowZero := false
			confTag := field.Tag.Get("config")

			for _, part := range strings.Split(confTag, ",") {
				if part == "zerodefault" {
					allowZero = true
					break
				}
			}

			if !allowZero {
				continue
			}
		}

		// Extract map key name form the field tag, if it is set. The
		// tag name is specified in the decoder config, and is set to
		// "mapstructure" by NewDecoder() if not expressly specified.
		tag := field.Tag.Get(o.config.TagName)
		if i := strings.Index(tag, ","); i != -1 {
			tag = tag[:i]
		}

		var fieldName string
		if tag != "" {
			fieldName = tag
		} else {
			fieldName = caser.String(field.Name)
		}

		result[fieldName] = fieldVal.Interface()
	}

	return result, nil
}
