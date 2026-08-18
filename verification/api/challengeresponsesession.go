// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

// The api package implements the REST API defined in
// https://github.com/veraison/docs/blob/main/api/challenge-response
package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
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

type EvidenceBlob struct {
	Type  string `json:"type"`
	Value []byte `json:"value"`
}

type nonce []byte

func (o nonce) MarshalJSON() ([]byte, error) {
	return json.Marshal(base64.URLEncoding.EncodeToString(o))
}

func (o *nonce) UnmarshalJSON(b []byte) error {
	var v string

	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	nonce, err := decodeSessionNonce(v)
	if err != nil {
		return fmt.Errorf("nonce must be valid base64url: %w", err)
	}

	*o = nonce

	return nil
}

type ChallengeResponseSession struct {
	id       string
	Status   Status        `json:"status"`
	Nonce    nonce         `json:"nonce"`
	Expiry   time.Time     `json:"expiry"`
	Accept   []string      `json:"accept"`
	Evidence *EvidenceBlob `json:"evidence,omitempty"`
	Result   *string       `json:"result,omitempty"`
}

func decodeSessionNonce(v string) ([]byte, error) {
	nonce, err := base64.URLEncoding.DecodeString(v)
	if err == nil {
		return nonce, nil
	}

	// Keep reading sessions created before the nonce wire format switched
	// from standard base64 to base64url.
	if nonce, stdErr := base64.StdEncoding.DecodeString(v); stdErr == nil {
		return nonce, nil
	}

	return nil, err
}

func (o *ChallengeResponseSession) SetEvidence(mt string, evidence []byte) {
	o.Evidence = &EvidenceBlob{Type: mt, Value: evidence}
}

func (o *ChallengeResponseSession) SetStatus(status Status) {
	o.Status = status
}

func (o *ChallengeResponseSession) SetResult(result []byte) {
	rs := string(result)
	o.Result = &rs
}
