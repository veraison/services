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

// URLSafeNonce is a wrapper around []byte that marshals/unmarshals using URL-safe base64
type URLSafeNonce []byte

func (n URLSafeNonce) MarshalJSON() ([]byte, error) {
	if n == nil {
		return []byte("null"), nil
	}
	encoded := base64.URLEncoding.EncodeToString(n)
	return json.Marshal(encoded)
}

func (n *URLSafeNonce) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	
	decoded, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return err
	}
	
	*n = URLSafeNonce(decoded)
	return nil
}

type EvidenceBlob struct {
	Type  string `json:"type"`
	Value []byte `json:"value"`
}

type ChallengeResponseSession struct {
	id       string
	Status   Status        `json:"status"`
	Nonce    URLSafeNonce  `json:"nonce"`
	Expiry   time.Time     `json:"expiry"`
	Accept   []string      `json:"accept"`
	Evidence *EvidenceBlob `json:"evidence,omitempty"`
	Result   *string       `json:"result,omitempty"`
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
