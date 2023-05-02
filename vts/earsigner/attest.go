// Copyright 2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package earsigner

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hf/nsm"
	"github.com/hf/nsm/request"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

func nitroAttest(k jwk.Key) ([]byte, error) {
	buf, err := json.Marshal(k)
	if err != nil {
		return nil, fmt.Errorf("failure marshalling JWK: %w", err)
	}

	ns, err := nsm.OpenDefaultSession()
	defer ns.Close()

	if err != nil {
		return nil, fmt.Errorf("failure opening NSM device: %w", err)
	}

	res, err := ns.Send(&request.Attestation{PublicKey: buf})
	if nil != err {
		return nil, fmt.Errorf("failure sending request over NSM device: %w", err)
	}

	if res.Error != "" {
		return nil, fmt.Errorf("response from NSM device contains error: %v", res.Error)
	}

	if res.Attestation == nil || res.Attestation.Document == nil {
		return nil, errors.New("response from NSM device does not contain an attestation")
	}

	return res.Attestation.Document, nil
}
