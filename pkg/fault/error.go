// Copyright 2024-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package fault

import (
	"errors"
	"fmt"
	"strings"
)

var (
	// ErrBadEvidence indicates that there is an issue with the evidence
	// provided by the attester.
	ErrBadEvidence = errors.New("bad evidence")

	// ErrBadInput indicates that there is an issue with user-provided
	// input.
	ErrBadInput = errors.New("bad input")

	// ErrSkippedPlugin may be returned from an IPluggable's Init() to
	// indicate that the interface could not be loaded due to some
	// anticipated circumstance, and that it should simply be skipped
	// instead of propagating the error.
	ErrSkippedPlugin = errors.New("skipped plugin")

	// ErrConfig indicates a problem with configuration provided to services.
	ErrConfig = errors.New("config")

	// ErrNotFound indicates that the items indicated by input could not be found.
	ErrNotFound = errors.New("not found")

	// ErrNotFound indicates that requested functionality or input is not
	// supported.
	ErrNotSupported = errors.New("not supported")

	// ErrNotSet indicates a parameter or configuration value has not been set.
	ErrNotSet = errors.New("not set")

	// ErrInvalid indicates provided value is invalid.
	ErrInvalid = errors.New("invalid")

	// ErrNoConnection indicates that a connection to the VTS server could
	// not be established.
	ErrNoConnection = errors.New("could not connect to VTS server")
)

// BadEvidence creates a new error wrapping ErrBadEvidence. The arguments must either
// - a single error
// - a format string followed by values to formatted into it (see fmt.Errorf())
// - a single struct that implements String()
func BadEvidence(args ...any) error {
	return makeError(ErrBadEvidence, args...)
}

// BadInput creates a new error wrapping ErrBadInput. The arguments must either
// - a single error
// - a format string followed by values to formatted into it (see fmt.Errorf())
// - a single struct that implements String()
func BadInput(args ...any) error {
	return makeError(ErrBadInput, args...)
}

// SkippedPlugin creates a new error wrapping ErrSkippedPlugin. The arguments must either
// - a single error
// - a format string followed by values to formatted into it (see fmt.Errorf())
// - a single struct that implements String()
func SkippedPlugin(args ...any) error {
	return makeError(ErrSkippedPlugin, args...)
}

// BadConfig creates a new error wrapping ErrConfig. The category is prefixed
// to the error, providing an indication of where in configuration to look for
// the problem. The arguments must either
// - a single error
// - a format string followed by values to formatted into it (see fmt.Errorf())
// - a single struct that implements String()
func BadConfig(category string, args ...any) error {
	err := fmt.Errorf("%s %w", category, ErrConfig)
	return makeError(err, args...)
}

// NotSupported creates a new error wrapping ErrNotSupported. The arguments must either
// - a single error
// - a format string followed by values to formatted into it (see fmt.Errorf())
// - a single struct that implements String()
func NotSupported(args ...any) error {
	return makeError(ErrNotSupported, args...)
}

// NotFound creates a new error wrapping ErrNotFound. The arguments must either
// - a single error
// - a format string followed by values to formatted into it (see fmt.Errorf())
// - a single struct that implements String()
func NotFound(args ...any) error {
	return makeError(ErrNotFound, args...)
}

// NotSet creates a new error wrapping ErrNotSet. The arguments must either
// - a single error
// - a format string followed by values to formatted into it (see fmt.Errorf())
// - a single struct that implements String()
func NotSet(args ...any) error {
	return makeError(ErrNotSet, args...)
}

// Invalid creates a new error wrapping ErrInvalid. The arguments must either
// - a single error
// - a format string followed by values to formatted into it (see fmt.Errorf())
// - a single struct that implements String()
func Invalid(args ...any) error {
	return makeError(ErrInvalid, args...)
}

// NoConnection creates a new error wrapping ErrNoConnection. The arguments
// must either
// - a single error
// - a format string followed by values to formatted into it (see fmt.Errorf())
// - a single struct that implements String()
func NoConnection(args ...any) error {
	return makeError(ErrNoConnection, args...)
}

const errorSeparator = "@:;:@"

// SerializeError serializes an error for transmission across RPC. The error is
// fully unwrapped and individual nested parts are joined with a unique
// separator into a single message.
func SerializeError(err error) error {
	if err == nil {
		return nil
	}

	messages := errorMessages(err)
	return errors.New(strings.Join(messages, errorSeparator))
}

// ParseError re-constitutes an error transmitted across RPC. The error message is
// split on a unique separator and the resulting parts are used to wrap each other,
// with the first part providing the outermost level.
func ParseError(err error) error {
	if err == nil {
		return nil
	}

	messages := strings.Split(err.Error(), errorSeparator)

	var ret error
	for i := len(messages) - 1; i >= 0; i-- {
		ret = wraptError{
			msg: messages[i],
			err: ret,
		}
	}

	return ret
}

// makeError creates a new error wrapping the error provided as the first
// argument using subsequent arguments. The subsequent args must be one of the
// following:
// - a single error
// - a format string followed by values to formatted into it (see fmt.Errorf())
// - a single struct that implements String()
func makeError(toWrap error, args ...any) error {
	if len(args) == 0 {
		return toWrap
	}

	head := args[0]
	args = args[1:]

	switch t := head.(type) {
	case error:
		if toWrap != nil {
			return fmt.Errorf("%w: %w", toWrap, t)
		} else {
			return t
		}
	case string:
		if toWrap != nil {
			msg := fmt.Sprintf("%%w: %s", t)
			args = append([]any{toWrap}, args...)
			return fmt.Errorf(msg, args...)
		} else {
			return fmt.Errorf(t, args...)
		}
	case interface{ String() string }:
		if toWrap != nil {
			return fmt.Errorf("%w: %s", toWrap, t.String())
		} else {
			return errors.New(t.String())
		}
	default:
		panic(fmt.Sprintf("valid error value %v (%T)", t, t))
	}
}

func errorMessages(input error) []string {
	var ret []string
	toProcess := []error{input}

	for len(toProcess) > 0 {
		var unwrapped []error

		for _, err := range toProcess {
			if err == nil {
				continue
			}

			ret = append(ret, err.Error())

			switch t := err.(type) {
			case interface{ Unwrap() error }:
				unwrapped = append(unwrapped, t.Unwrap())
			case interface{ Unwrap() []error }:
				unwrapped = append(unwrapped, t.Unwrap()...)
			}
		}

		toProcess = unwrapped
	}

	return ret
}

type wraptError struct {
	msg string
	err error
}

func (o wraptError) Error() string {
	return o.msg
}

func (o wraptError) Unwrap() error {
	return o.err
}

func (o wraptError) Is(other error) bool {
	if other == nil {
		return false
	}

	if o.Error() == other.Error() {
		return true
	}

	if o.err == nil {
		return false
	}

	if _, ok := o.err.(wraptError); !ok {
		// We've  at the the final wrapping layer. errors.Is uses == to
		// establish error (i.e. interface{}) equality, since that
		// amounts to comparing pointers, equality won't be preserved
		// across serialization/desrialization. Therefore, we need to
		// do string comparison rather than rely on errors.Is.
		return o.err.Error() == other.Error()
	}

	return errors.Is(o.err, other)
}
