// Copyright 2022-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

// The api package implements the REST API defined in
// https://github.com/veraison/docs/blob/main/api/challenge-response
package session

import (
	"encoding/json"
	"fmt"
)

type Status uint8

const (
	StatusWaiting Status = iota
	StatusProcessing
	StatusComplete
	StatusFailed
)

func (o Status) String() string {
	switch o {
	case StatusWaiting:
		return "waiting"
	case StatusProcessing:
		return "processing"
	case StatusComplete:
		return "complete"
	case StatusFailed:
		return "failed"
	}
	return "unknown"
}

func (o *Status) FromString(s string) error {
	switch s {
	case "waiting":
		*o = StatusWaiting
	case "processing":
		*o = StatusProcessing
	case "complete":
		*o = StatusComplete
	case "failed":
		*o = StatusFailed
	default:
		return fmt.Errorf("unknown status %s", s)
	}
	return nil
}

func (o Status) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.String())
}

func (o *Status) UnmarshalJSON(b []byte) error {
	var s string

	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	return o.FromString(s)
}
