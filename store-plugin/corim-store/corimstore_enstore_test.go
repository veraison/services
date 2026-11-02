// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package corim_store

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/veraison/cmw"
	"github.com/veraison/corim/coserv"
)

func rimsWithItems(t *testing.T, n int) *cmw.CMW {
	t.Helper()

	rims, err := cmw.NewCollection("")
	require.NoError(t, err)

	keys := []string{"manifest-a", "manifest-b"}
	for _, key := range keys[:n] {
		item, err := cmw.NewMonad("application/corim+cbor", []byte{0xa0})
		require.NoError(t, err)
		require.NoError(t, rims.AddCollectionItem(key, item))
	}

	return rims
}

func Test_resultSetIsEmpty(t *testing.T) {
	expiry := time.Now().Add(time.Hour)

	testCases := []struct {
		title    string
		rs       func(t *testing.T) *coserv.ResultSet
		expected bool
	}{
		{
			title:    "nil result set",
			rs:       func(*testing.T) *coserv.ResultSet { return nil },
			expected: true,
		},
		{
			title: "empty",
			rs: func(*testing.T) *coserv.ResultSet {
				return coserv.NewResultSet().SetExpiry(expiry)
			},
			expected: true,
		},
		{
			title: "RIMs collection with no items",
			rs: func(t *testing.T) *coserv.ResultSet {
				return coserv.NewResultSet().SetRIMs(*rimsWithItems(t, 0)).SetExpiry(expiry)
			},
			expected: true,
		},
		{
			title: "zero-length slices",
			rs: func(*testing.T) *coserv.ResultSet {
				rs := coserv.NewResultSet().SetExpiry(expiry)
				rs.RVQ = &[]coserv.RefValQuad{}
				rs.EVQ = &[]coserv.EndValQuad{}
				rs.CEQ = &[]coserv.CondEndValQuad{}
				rs.AKQ = &[]coserv.AKQuad{}
				rs.TAS = &[]coserv.CoTSStmt{}
				rs.SourceArtifacts = &[]cmw.CMW{}
				return rs
			},
			expected: true,
		},
		{
			title: "RIMs only",
			rs: func(t *testing.T) *coserv.ResultSet {
				return coserv.NewResultSet().SetRIMs(*rimsWithItems(t, 1)).SetExpiry(expiry)
			},
		},
		{
			title: "RIMs only, several items",
			rs: func(t *testing.T) *coserv.ResultSet {
				return coserv.NewResultSet().SetRIMs(*rimsWithItems(t, 2)).SetExpiry(expiry)
			},
		},
		{
			title: "AKQ only",
			rs: func(*testing.T) *coserv.ResultSet {
				return coserv.NewResultSet().AddAttestationKeys(coserv.AKQuad{}).SetExpiry(expiry)
			},
		},
		{
			title: "RVQ only",
			rs: func(*testing.T) *coserv.ResultSet {
				return coserv.NewResultSet().AddReferenceValues(coserv.RefValQuad{}).SetExpiry(expiry)
			},
		},
		{
			title: "EVQ only",
			rs: func(*testing.T) *coserv.ResultSet {
				return coserv.NewResultSet().AddEndorsedValues(coserv.EndValQuad{}).SetExpiry(expiry)
			},
		},
		{
			title: "CEQ only",
			rs: func(*testing.T) *coserv.ResultSet {
				return coserv.NewResultSet().AddConditionalEndorsementValues(coserv.CondEndValQuad{}).SetExpiry(expiry)
			},
		},
		{
			title: "TAS only",
			rs: func(*testing.T) *coserv.ResultSet {
				return coserv.NewResultSet().AddCoTS(coserv.CoTSStmt{}).SetExpiry(expiry)
			},
		},
		{
			title: "source artifacts only",
			rs: func(t *testing.T) *coserv.ResultSet {
				return coserv.NewResultSet().AddSourceArtifacts(*rimsWithItems(t, 1)).SetExpiry(expiry)
			},
		},
		{
			title: "mixed",
			rs: func(t *testing.T) *coserv.ResultSet {
				rs := coserv.NewResultSet().
					AddReferenceValues(coserv.RefValQuad{}).
					AddAttestationKeys(coserv.AKQuad{}).
					SetRIMs(*rimsWithItems(t, 1)).
					SetExpiry(expiry)
				rs.EVQ = &[]coserv.EndValQuad{}
				return rs
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			assert.Equal(t, tc.expected, resultSetIsEmpty(tc.rs(t)))
		})
	}
}
