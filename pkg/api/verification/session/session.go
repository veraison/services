// Copyright 2022-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

// The api package implements the REST API defined in
// https://github.com/veraison/docs/blob/main/api/challenge-response
package session

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/veraison/services/pkg/util"
)

type Session struct {
	ID       string        `json:"id,omitempty"`
	Status   Status        `json:"status"`
	Nonce    nonce         `json:"nonce"`
	Expiry   time.Time     `json:"expiry"`
	Accept   []string      `json:"accept"`
	Evidence *evidenceBlob `json:"evidence,omitempty"`
	Result   *string       `json:"result,omitempty"`
}

func (o *Session) SetEvidence(mt string, evidence []byte) {
	o.Evidence = &evidenceBlob{Type: mt, Value: evidence}
}

func (o *Session) SetStatus(status Status) {
	o.Status = status
}

func (o *Session) SetResult(result []byte) {
	rs := string(result)
	o.Result = &rs
}

type evidenceBlob struct {
	Type  string `json:"type"`
	Value []byte `json:"value"`
}

type nonce []byte

func (o nonce) MarshalJSON() ([]byte, error) {
	return json.Marshal(base64.RawURLEncoding.EncodeToString(o))
}

func (o *nonce) UnmarshalJSON(b []byte) error {
	var v string

	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	nonce, err := util.Base64URLToBytes(v)
	if err != nil {
		return fmt.Errorf("nonce must be valid base64url: %w", err)
	}

	*o = nonce

	return nil
}
