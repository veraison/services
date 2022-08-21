// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package policy

// Policy allows enforcing additional constraints on top of the regular attestation schemes.
type Policy struct {
	// ID is used to reference the policy in the result.
	ID string

	// Rules of the policy to be interpreted and execute by the policy agent.
	Rules string
}
