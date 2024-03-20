// Copyright 2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// BadEvidenceError represents an error due a problem with the received evidence.
// IEvidenceHandler implementations should return an instance of this
// (constructed using BadEvidence() below) if they could not process the
// provided evidence token.
type BadEvidenceError struct {
	Detail interface{}
}

// The goal of MarshalJSON and UnmarshalJSON below is to make
// serialization/deserialization as transparrent as possible. This means
// accurately preserving Detail's structure.

func (o BadEvidenceError) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"error": "bad evidence",
	}

	switch t := o.Detail.(type) {
	case string:
		m["detail-type"] = "string"
		m["detail"] = t
	case error:
		err := t
		var msgs []string
		for {
			msgs = append(msgs, err.Error())
			if uerr, ok := err.(interface{ Unwrap() error }); ok {
				err = uerr.Unwrap()
			} else {
				break
			}
		}

		m["detail-type"] = "error"
		m["detail"] = msgs
	default:
		m["detail-type"] = "other"
		m["detail"] = o.Detail
	}

	return json.Marshal(m)
}

func (o *BadEvidenceError) UnmarshalJSON(data []byte) error {
	var m map[string]interface{}

	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	errType, ok := m["error"]
	if !ok || errType != "bad evidence" {
		return errors.New("not a BadEvidenceError")
	}

	detailType, ok := m["detail-type"]
	if !ok {
		return errors.New("missing detail-type")
	}
	switch detailType {
	case "string":
		switch t := m["detail"].(type) {
		case string:
			o.Detail = t
		case []byte:
			o.Detail = string(t)
		default:
			return fmt.Errorf("unexpected value for string detail: %v (%T)", t, t)
		}
	case "error":
		switch t := m["detail"].(type) {
		case []interface{}:
			if len(t) < 1 { // nolint:gocritic
				return errors.New("empty messages for error detail")
			} else if len(t) == 1 {
				o.Detail = errors.New(fmt.Sprint(t[0]))
			} else {
				err := errors.New(fmt.Sprint(t[len(t)-1]))
				for i := len(t) - 2; i >= 0; i-- {
					err = wraptError{
						msg: fmt.Sprint(t[i]),
						err: err,
					}
				}
				o.Detail = err
			}
		default:
			return fmt.Errorf("unexpected value for string detail: %v (%T)", t, t)
		}
	case "other":
		o.Detail = m["detail"]
	default:
		return fmt.Errorf("unexpected detail type: %s", detailType)
	}

	return nil
}

func (o BadEvidenceError) Error() string {
	data, err := o.MarshalJSON()
	if err != nil {
		panic(err)
	}

	return string(data)
}

func (o BadEvidenceError) ToString() string {
	switch t := o.Detail.(type) {
	case string:
		return fmt.Sprintf("bad evidence: %s", t)
	case error:
		return fmt.Sprintf("bad evidence: %s", t.Error())
	default:
		return fmt.Sprintf("bad evidence: %v", t)
	}
}

func (o BadEvidenceError) Unwrap() error {
	switch t := o.Detail.(type) {
	case error:
		return t
	default:
		return nil
	}
}

func (o BadEvidenceError) Is(other error) bool {
	if _, ok := other.(BadEvidenceError); ok {
		return true
	}

	if wrapt, ok := other.(interface{ Unwrap() error }); ok {
		return o.Is(wrapt.Unwrap())
	}

	return false
}

// BadEvidence creates a new BadEvidenceError instance using the provided args
// to construct the detail. If no args are specified, the generic detail of
// "invalid"  is used. If exactly one argument is specified, it is used as the
// detial. If more than one ergument is specified, the behavior depends on the
// type of the first argument.
// When args[0] is a string a new error is created using fmt.Errorf, using
// args[0] as the format, and that error is used as the detail.
// Otherwise, the entire args slice is used as the detail.
func BadEvidence(args ...interface{}) error {
	var detail interface{}

	switch len(args) {
	case 0:
		detail = "invalid"
	case 1:
		detail = args[0]
	default: // args longer than 1
		switch t := args[0].(type) {
		case string:
			detail = fmt.Errorf(t, args[1:]...)
		default:
			detail = args[0]
		}
	}

	return BadEvidenceError{detail}
}

func ParseError(err error) error {
	if err == nil {
		return nil
	}

	msg := err.Error()

	if strings.HasPrefix(msg, "bad evidence: ") {
		return BadEvidenceError{msg[14:]}
	}

	var bee BadEvidenceError
	var decErr error
	if decErr = json.Unmarshal([]byte(msg), &bee); decErr == nil {
		return bee
	}

	return err
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
	if o.Error() == other.Error() {
		return true
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
