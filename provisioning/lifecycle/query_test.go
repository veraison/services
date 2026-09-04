// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package lifecycle

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/veraison/corim/comid"
	"github.com/veraison/corim/coserv"
	"github.com/veraison/eat"
	"github.com/veraison/swid"
)

func TestQuery_Valid(t *testing.T) {
	profile := &eat.Profile{}
	_ = profile.Set("tag:arm.com")
	artifactType := coserv.ArtifactTypeReferenceValues
	instance := coserv.StatefulInstance{
		Instance: comid.MustNewBytesInstance(
			comid.MustHexDecode(t, "feeefeee"),
		),
	}
	tagID := *swid.NewTagID("foo")

	validEnvSelector := *coserv.NewEnvironmentSelector().AddInstance(instance)
	validRimSelector := coserv.NewRimSelectorIDs().Add(&coserv.RimSelectorID{
		TagID: tagID,
		Type:  coserv.RimSelectorTypeCorim,
	})

	testCases := []struct {
		title string
		query Query
		err   string
	}{
		{
			title: "empty",
			query: Query{},
			err:   "no selector specified",
		},
		{
			title: "both selectors",
			query: Query{
				EnvironmentSelector: &validEnvSelector,
				RimSelector:         validRimSelector,
				Profile:             profile,
				ArtifactType:        &artifactType,
			},
			err: "environment and RIM selectors cannot be specified at the same time",
		},
		{
			title: "environment selector without profile",
			query: Query{
				EnvironmentSelector: &validEnvSelector,
				ArtifactType:        &artifactType,
			},
			err: "profile must be specified with an environment selector",
		},
		{
			title: "environment selector without artifact type",
			query: Query{
				EnvironmentSelector: &validEnvSelector,
				Profile:             profile,
			},
			err: "artifact type must be specified with an environment selector",
		},
		{
			title: "environment selector invalid",
			query: Query{
				EnvironmentSelector: &coserv.EnvironmentSelector{},
				Profile:             profile,
				ArtifactType:        &artifactType,
			},
			err: "invalid environment selector",
		},
		{
			title: "environment selector valid",
			query: Query{
				EnvironmentSelector: &validEnvSelector,
				Profile:             profile,
				ArtifactType:        &artifactType,
			},
		},
		{
			title: "rim selector invalid",
			query: Query{
				RimSelector: coserv.NewRimSelectorIDs(),
			},
			err: "invalid RIM selector",
		},
		{
			title: "rim selector with artifact type",
			query: Query{
				RimSelector:  validRimSelector,
				ArtifactType: &artifactType,
			},
			err: "artifact type cannot be specified with a RIM selector",
		},
		{
			title: "rim selector with profile",
			query: Query{
				RimSelector: validRimSelector,
				Profile:     profile,
			},
			err: "profile cannot be specified with a RIM selector",
		},
		{
			title: "rim selector valid",
			query: Query{
				RimSelector: validRimSelector,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			err := tc.query.Valid()

			if tc.err == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tc.err)
			}
		})
	}
}

func TestNewEnvironmentQuery(t *testing.T) {
	profile := &eat.Profile{}
	_ = profile.Set("tag:arm.com")
	artifactType := coserv.ArtifactTypeReferenceValues

	_, err := NewEnvironmentQuery(
		*profile,
		artifactType,
		coserv.EnvironmentSelector{},
	)
	assert.ErrorContains(t, err, "invalid environment selector")

	selector := *coserv.NewEnvironmentSelector().AddInstance(
		coserv.StatefulInstance{
			Instance: comid.MustNewBytesInstance(
				comid.MustHexDecode(t, "feeefeee"),
			),
		},
	)

	query, err := NewEnvironmentQuery(
		*profile,
		artifactType,
		selector,
	)

	assert.NoError(t, err)
	assert.NotNil(t, query)
	assert.Equal(t, profile, query.Profile)
	assert.Equal(t, artifactType, *query.ArtifactType)
	assert.Equal(t, selector, *query.EnvironmentSelector)
	assert.Nil(t, query.RimSelector)
}

func TestNewRimQuery(t *testing.T) {
	tagID := *swid.NewTagID("foo")

	_, err := NewRimQuery(
		coserv.RimSelectorTypeCorim,
		swid.TagID{},
	)
	assert.ErrorContains(t, err, "tag-id value is nil")

	query, err := NewRimQuery(
		coserv.RimSelectorTypeCorim,
		tagID,
	)

	assert.NoError(t, err)
	assert.NotNil(t, query)
	assert.Nil(t, query.EnvironmentSelector)
	assert.Nil(t, query.Profile)
	assert.Nil(t, query.ArtifactType)
	assert.NotNil(t, query.RimSelector)
}

func TestQuery_ToCBOR(t *testing.T) {
	profile := &eat.Profile{}
	_ = profile.Set("tag:arm.com")
	artifactType := coserv.ArtifactTypeReferenceValues

	selector := *coserv.NewEnvironmentSelector().AddInstance(
		coserv.StatefulInstance{
			Instance: comid.MustNewBytesInstance(
				comid.MustHexDecode(t, "feeefeee"),
			),
		},
	)

	query := Query{
		Profile:             profile,
		ArtifactType:        &artifactType,
		EnvironmentSelector: &selector,
	}

	data, err := query.ToCBOR()

	assert.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestQuery_ToCBOR_Invalid(t *testing.T) {
	query := Query{}

	data, err := query.ToCBOR()

	assert.Nil(t, data)
	assert.ErrorContains(t, err, "validating ELM Query")
	assert.ErrorContains(t, err, "no selector specified")
}

func TestQuery_FromCBOR(t *testing.T) {
	profile := &eat.Profile{}
	_ = profile.Set("tag:arm.com")
	artifactType := coserv.ArtifactTypeReferenceValues

	selector := *coserv.NewEnvironmentSelector().AddInstance(
		coserv.StatefulInstance{
			Instance: comid.MustNewBytesInstance(
				comid.MustHexDecode(t, "feeefeee"),
			),
		},
	)

	original := Query{
		Profile:             profile,
		ArtifactType:        &artifactType,
		EnvironmentSelector: &selector,
	}

	data, err := original.ToCBOR()
	assert.NoError(t, err)

	var decoded Query
	err = decoded.FromCBOR(data)

	assert.NoError(t, err)
	assert.Equal(t, original, decoded)
}

func TestQuery_FromCBOR_InvalidEncoding(t *testing.T) {
	var query Query

	err := query.FromCBOR([]byte{0xff, 0xff, 0xff})

	assert.ErrorContains(t, err, "decoding ELM Query from CBOR")
}

func TestQuery_FromCBOR_InvalidQuery(t *testing.T) {
	query := Query{}

	data, err := query.ToCBOR()
	assert.Error(t, err)
	assert.Nil(t, data)
}

func TestQuery_ToJSON(t *testing.T) {
	profile := &eat.Profile{}
	_ = profile.Set("tag:arm.com")
	artifactType := coserv.ArtifactTypeReferenceValues

	selector := *coserv.NewEnvironmentSelector().AddInstance(
		coserv.StatefulInstance{
			Instance: comid.MustNewBytesInstance(
				comid.MustHexDecode(t, "feeefeee"),
			),
		},
	)

	query := Query{
		Profile:             profile,
		ArtifactType:        &artifactType,
		EnvironmentSelector: &selector,
	}

	data, err := query.ToJSON()

	assert.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestQuery_ToJSON_Invalid(t *testing.T) {
	query := Query{}

	data, err := query.ToJSON()

	assert.Nil(t, data)
	assert.ErrorContains(t, err, "validating ELM Query")
	assert.ErrorContains(t, err, "no selector specified")
}

func TestQuery_FromJSON(t *testing.T) {
	profile := &eat.Profile{}
	_ = profile.Set("tag:arm.com")
	artifactType := coserv.ArtifactTypeReferenceValues

	selector := *coserv.NewEnvironmentSelector().AddInstance(
		coserv.StatefulInstance{
			Instance: comid.MustNewBytesInstance(
				comid.MustHexDecode(t, "feeefeee"),
			),
		},
	)

	original := Query{
		Profile:             profile,
		ArtifactType:        &artifactType,
		EnvironmentSelector: &selector,
	}

	data, err := original.ToJSON()
	assert.NoError(t, err)

	var decoded Query
	err = decoded.FromJSON(data)

	assert.NoError(t, err)
	assert.Equal(t, original, decoded)
}

func TestQuery_FromJSON_Invalid(t *testing.T) {
	var query Query

	err := query.FromJSON([]byte(`{"profile":`))

	assert.Error(t, err)
}

func TestQuery_ToEDN(t *testing.T) {
	profile := &eat.Profile{}
	_ = profile.Set("tag:arm.com")
	artifactType := coserv.ArtifactTypeReferenceValues

	selector := *coserv.NewEnvironmentSelector().AddInstance(
		coserv.StatefulInstance{
			Instance: comid.MustNewBytesInstance(
				comid.MustHexDecode(t, "feeefeee"),
			),
		},
	)

	query := Query{
		Profile:             profile,
		ArtifactType:        &artifactType,
		EnvironmentSelector: &selector,
	}

	edn, err := query.ToEDN()

	assert.NoError(t, err)
	assert.NotEmpty(t, edn)
}

func TestQuery_ToEDN_Invalid(t *testing.T) {
	query := Query{}

	edn, err := query.ToEDN()

	assert.Empty(t, edn)
	assert.ErrorContains(t, err, "failed encoding ELM Query object")
	assert.ErrorContains(t, err, "no selector specified")
}
