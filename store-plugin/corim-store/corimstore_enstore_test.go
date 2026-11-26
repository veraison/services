// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package corim_store

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/veraison/cmw"
	corimstore "github.com/veraison/corim-store/pkg/store"
	"github.com/veraison/corim/comid"
	"github.com/veraison/corim/corim"
	"github.com/veraison/corim/coserv"
	"github.com/veraison/eat"
	"github.com/veraison/services/handler"
	vtscoserv "github.com/veraison/services/vts/coserv"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
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

var (
	testClassID       = []byte{0x01, 0x02, 0x03, 0x04}
	testComponentName = "test-component"
)

func newTestStore(t *testing.T, observed zapcore.Core) *Store {
	t.Helper()

	cfg := corimstore.NewConfig("sqlite3", filepath.Join(t.TempDir(), "store.db"))
	cs, err := corimstore.Open(context.Background(), cfg)
	require.NoError(t, err)
	require.NoError(t, cs.Init())
	t.Cleanup(func() { assert.NoError(t, cs.Close()) })

	authority, err := comid.NewCryptoKeyTaggedBytes("test")
	require.NoError(t, err)

	return &Store{
		Store:  cs,
		logger: zap.New(observed).Sugar(),
		CoservCfg: &vtscoserv.StoreConfig{
			Authority: authority,
			MaxExpiry: time.Hour,
		},
	}
}

func addTestRefValue(t *testing.T, s *Store) {
	t.Helper()

	triple := &comid.ValueTriple{
		Environment: comid.Environment{Class: comid.NewClassBytes(testClassID)},
		Measurements: *comid.NewMeasurements().Add(&comid.Measurement{
			Val: comid.Mval{Name: &testComponentName},
		}),
	}
	tag := comid.NewComid().
		SetTagIdentity("test-tag", 0).
		AddEntity("test", nil, comid.RoleCreator).
		AddReferenceValue(triple)
	require.NoError(t, tag.Valid())

	c := corim.NewUnsignedCorim().
		SetID("test-corim").
		SetProfile("http://example.com").
		AddComid(tag).
		AddEntity("test", nil, corim.RoleManifestCreator)
	data, err := c.ToCBOR()
	require.NoError(t, err)

	require.NoError(t, s.Store.AddBytes(data, "", true))
}

func testRVQuery(t *testing.T) string {
	t.Helper()

	profile, err := eat.NewProfile("http://example.com")
	require.NoError(t, err)

	artifactType := coserv.ArtifactTypeReferenceValues
	resultType := coserv.ResultTypeCollectedArtifacts
	query := coserv.Coserv{
		Profile: *profile,
		Query: coserv.Query{
			ArtifactType: &artifactType,
			EnvironmentSelector: coserv.NewEnvironmentSelector().
				AddClass(coserv.StatefulClass{
					Class: comid.NewClassBytes(testClassID),
				}),
			ResultType: &resultType,
		},
	}
	encoded, err := query.ToBase64Url()
	require.NoError(t, err)

	return encoded
}

func Test_ExecuteCoservQuery_empty_result_set(t *testing.T) {
	core, _ := observer.New(zap.WarnLevel)
	store := newTestStore(t, core)

	res, err := store.ExecuteCoservQuery("http://example.com", testRVQuery(t))

	assert.ErrorIs(t, err, handler.ErrNotFound)
	require.NotNil(t, res)
	assert.True(t, resultSetIsEmpty(res.Results))
}

func Test_ExecuteCoservQuery_non_empty_result_set(t *testing.T) {
	core, logs := observer.New(zap.WarnLevel)
	store := newTestStore(t, core)
	addTestRefValue(t, store)

	res, err := store.ExecuteCoservQuery("http://example.com", testRVQuery(t))

	require.NoError(t, err)
	require.NotNil(t, res)
	assert.False(t, resultSetIsEmpty(res.Results))
	assert.Equal(t, 0, logs.Len())
}
