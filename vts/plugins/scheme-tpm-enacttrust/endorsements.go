// Copyright 2021-2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"fmt"
)

type Endorsements struct {
	Digest string
}

func (e *Endorsements) Populate(strings []string) error {
	l := len(strings)

	if l != 1 {
		return fmt.Errorf("incorrect endorsements number: want 1, got %d", l)
	}

	e.Digest = strings[0]

	return nil
}
