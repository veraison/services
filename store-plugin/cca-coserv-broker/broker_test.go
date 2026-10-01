// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package ccacoservbroker

import (
	"crypto"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fxamacker/cbor/v2"
	jose "github.com/go-jose/go-jose/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/veraison/apiclient/common"
	coservBuilder "github.com/veraison/apiclient/coserv"
	"github.com/veraison/corim/comid"
	"github.com/veraison/corim/coserv"
	"github.com/veraison/corim/profiles/cca"
	"github.com/veraison/eat"
	cose "github.com/veraison/go-cose"
	"github.com/veraison/services/handler"
	"github.com/veraison/services/plugin"
)

var (
	testCoSERVURI          = "http://veraison.example"
	testRequestResponseURI = "http://veraison.example/endorsement-distribution/v1/coserv/{query}"
	sampleInstID           = "AQcGBQQDAgEADw4NDAsKCQgXFhUUExIREB8eHRwbGhkY"
	sampleImplID           = "f0VMRgIBAQAAAAAAAAAAAAMAPgABAAAAUFgAAAAAAAA="
	testES256Key           = map[string]string{
		"kty": "EC",
		"crv": "P-256",
		"x":   "MKBCTNIcKUSDii11ySs3526iDZ8AiTo7Tu6KPAqv7D4",
		"y":   "4Etl6SRW2YiLUrN5vfvVHuhp7x8PxltmWWlbbM4IFyM",
	}
	testES256KeyPrivateD   = "870MB6gfuTJ4HtUnUvYMyJpr5eUZNP4Bk43bVdj3eAE"
	SupportedCoservProfile = "tag:arm.com,2025:cca_platform#1.0.0"
)

// Helper: returns CBOR representation of JWK test public key
func retCBORPublicKey(t *testing.T) map[int]interface{} {
	t.Helper()

	xBytes, err := base64.RawURLEncoding.DecodeString(testES256Key["x"])
	require.NoError(t, err)

	yBytes, err := base64.RawURLEncoding.DecodeString(testES256Key["y"])
	require.NoError(t, err)

	testES256CBORKey := map[int]interface{}{
		1:  2,
		3:  -7,
		-1: 1,
		-2: xBytes,
		-3: yBytes,
	}
	return testES256CBORKey
}

// Helper: create plugin.Parameters with custom input.
func customParams(t *testing.T, enable any, coservUrl any, insecure any, certs any, caching any, unsigned any) *plugin.Parameters {
	t.Helper()

	paramsMap := map[string]any{
		"enable-broker":       enable,
		"coserv-url":          coservUrl,
		"client-tls-insecure": insecure,
		"client-certs":        certs,
		"client-caching":      caching,
		"unsigned-coserv":     unsigned,
	}

	params := plugin.NewParameters()
	err := params.PopulateFromMap(paramsMap)
	require.NoError(t, err)

	return params
}

// Helper: newDefaultQuery returns a fresh, valid CoSERV query.
func newDefaultQuery(t *testing.T) *coserv.Coserv {
	t.Helper()

	profile, err := eat.NewProfile(ccaPlatformProfile)
	require.NoError(t, err)

	class := comid.NewClassBytes([]byte{0x00, 0x11, 0x22, 0x33})
	class.SetVendor("ACME").SetModel("ARM CCA")
	envSelector := coserv.NewEnvironmentSelector()
	envSelector.AddClass(coserv.StatefulClass{Class: class})
	queryStruct, err := coserv.NewEnvironmentQuery(
		coserv.ArtifactTypeReferenceValues,
		*envSelector,
		coserv.ResultTypeCollectedArtifacts,
	)
	require.NoError(t, err)
	return &coserv.Coserv{Profile: *profile, Query: *queryStruct}
}

// Helper: newDefaultResponse returns a fresh, valid CoSERV response.
func newDefaultResponse(t *testing.T, query *coserv.Coserv) *coserv.Coserv {
	t.Helper()

	return &coserv.Coserv{
		Profile: query.Profile,
		Query:   query.Query,
		Results: coserv.NewResultSet().
			SetExpiry(time.Now().Add(24 * time.Hour)).
			AddReferenceValues(coserv.RefValQuad{}),
	}
}

// Helper: returns Discovery Document with custom keys and capabilites.
func newDiscoveryDocument(t *testing.T, mt string, key any) *coserv.DiscoveryDocument {
	t.Helper()

	doc := coserv.DiscoveryDocument{}
	doc.SetVersion("1.2.3")
	doc.AddCapability(mt, []coserv.ArtifactSupport{coserv.ArtifactSupportCollected})
	doc.AddEndPoint("CoSERVRequestResponse", "/endpoint/{query}")

	switch {
	case key == "json":
		keyBytes, err := json.Marshal(testES256Key)
		require.NoError(t, err)
		doc.AddJwk(keyBytes)
	case key == "cbor":
		keyBytes, err := cbor.Marshal(retCBORPublicKey(t))
		require.NoError(t, err)
		doc.AddCoseKey(keyBytes)
	case key != nil:
		keyBytes, err := json.Marshal(key)
		require.NoError(t, err)
		doc.AddJwk(keyBytes)
	}

	return &doc
}

// Helper: signs CoSERV response.
func signCoSERV(t *testing.T, c *coserv.Coserv) []byte {
	t.Helper()

	xBytes, err := base64.RawURLEncoding.DecodeString(testES256Key["x"])
	require.NoError(t, err)

	yBytes, err := base64.RawURLEncoding.DecodeString(testES256Key["y"])
	require.NoError(t, err)

	dBytes, err := base64.RawURLEncoding.DecodeString(testES256KeyPrivateD)
	require.NoError(t, err)

	key, err := cose.NewKeyEC2(cose.AlgorithmES256, xBytes, yBytes, dBytes)
	require.NoError(t, err)

	signer, err := key.Signer()
	require.NoError(t, err)

	signed, err := c.Sign(signer)
	require.NoError(t, err)

	return signed
}

// Helper: returns a client, which calls a server with custom handler code. Teardown function is also returned. signed specifies if the coserv response is signed or unsigned
func customBrokerClient(t *testing.T, h http.HandlerFunc, signed bool) (*Broker, func()) {
	t.Helper()
	broker := NewBroker()

	client, teardown := common.NewTestingHTTPClient(h)

	qcfg := &coservBuilder.QueryConfig{}
	err := qcfg.SetRequestResponseURI(testRequestResponseURI)
	require.NoError(t, err)

	err = qcfg.SetClient(client)
	require.NoError(t, err)

	broker.coservclient = qcfg

	if signed {
		var jwk jose.JSONWebKey
		keyBytes, err := json.Marshal(testES256Key)
		require.NoError(t, err)
		err = jwk.UnmarshalJSON(keyBytes)
		require.NoError(t, err)
		pb, ok := jwk.Key.(crypto.PublicKey)
		require.True(t, ok)
		broker.keys = append(broker.keys, pb)
	}

	return broker, teardown
}

// Helper: creates a sample comid Environment for TrustAnchors
func sampleTrustAnchorIDs(t *testing.T) comid.Environment {
	t.Helper()

	instanceID, err := comid.NewUEIDInstance(sampleInstID)
	require.NoError(t, err)

	ta := comid.Environment{
		Instance: instanceID,
	}

	return ta
}

// Helper: creates a sample comid Environment for Reference Values
func sampleReferenceValueIDs(t *testing.T) comid.Environment {
	t.Helper()

	implIDbytes, err := base64.StdEncoding.DecodeString(sampleImplID)
	require.NoError(t, err)

	classID, err := cca.NewPlatformImplIDClassID(implIDbytes)
	require.NoError(t, err)

	rv := comid.Environment{
		Class: &comid.Class{ClassID: classID},
	}

	return rv
}

// Helper: returns coserv TA test result with updated expiry of 1 hr
func getCoSERVTAResponse(t *testing.T) coserv.Coserv {
	t.Helper()

	var res coserv.Coserv
	err := res.FromCBOR(CoSERVTrustAnchorResponse)
	require.NoError(t, err)

	res.Results.SetExpiry(time.Now().Add(1 * time.Hour))
	return res

}

// Helper: returns coserv RV test result with updated expiry of 1 hr
func getCoSERVRVResponse(t *testing.T) coserv.Coserv {
	t.Helper()

	var res coserv.Coserv
	err := res.FromCBOR(CoSERVReferenceValueResponse)
	require.NoError(t, err)

	res.Results.SetExpiry(time.Now().Add(1 * time.Hour))
	return res

}

func TestBroker_InitParamsNil(t *testing.T) {
	broker := NewBroker()
	err := broker.Init(nil)

	assert.ErrorContains(t, err, "parameters should be non-nil")
}

func TestBroker_InitEnableBrokerNotBool(t *testing.T) {
	params := customParams(t, "fake", "", false, "", false, true)

	broker := NewBroker()
	err := broker.Init(params)

	assert.ErrorContains(t, err, "enable-broker config is not a string!")
}

func TestBroker_InitEnableBrokerFalse(t *testing.T) {
	params := customParams(t, false, "", false, "", false, true)

	broker := NewBroker()
	err := broker.Init(params)

	assert.ErrorContains(t, err, "cca-coserv-broker is disabled!")
}

func TestBroker_InitMissingCoSERVURL(t *testing.T) {

	params := plugin.NewParameters()
	err := params.Set("enable-broker", true)
	require.NoError(t, err)

	broker := NewBroker()
	err = broker.Init(params)
	assert.ErrorContains(t, err, "coserv-url config is not provided!")
}

func TestBroker_InitCoservURLNotString(t *testing.T) {
	params := customParams(t, true, false, false, "", false, true)

	broker := NewBroker()
	err := broker.Init(params)
	assert.ErrorContains(t, err, "coserv-url config is not a string!")
}

func TestBroker_InitCoservURLMalformed(t *testing.T) {
	params := customParams(t, true, "://invalid", false, "", false, true)

	broker := NewBroker()
	err := broker.Init(params)
	assert.ErrorContains(t, err, "failed to set coserv-url in discovery client!")
}

func TestBroker_InitInsecureNotBool(t *testing.T) {
	params := customParams(t, true, testCoSERVURI, "invalid", "", false, true)

	broker := NewBroker()
	err := broker.Init(params)
	assert.ErrorContains(t, err, "client-tls-insecure config is not a bool!")
}

func TestBroker_InitCertsNotString(t *testing.T) {
	params := customParams(t, true, testCoSERVURI, true, false, false, true)

	broker := NewBroker()
	err := broker.Init(params)
	assert.ErrorContains(t, err, "client-certs config is not a string!")
}

func TestBroker_InitCacheNotBool(t *testing.T) {
	params := customParams(t, true, testCoSERVURI, true, "/path/example1.crt,/path/example2.crt", "invalid", true)

	broker := NewBroker()
	err := broker.Init(params)
	assert.ErrorContains(t, err, "client-caching config is not a bool!")
}

func TestBroker_InitDiscoveryRunFailed(t *testing.T) {

	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer mock.Close()

	params := customParams(t, true, mock.URL, true, "/path/example1.crt,/path/example2.crt", true, true)

	broker := NewBroker()
	err := broker.Init(params)
	assert.ErrorContains(t, err, "failed to fetch discovery document!")
}

func TestBroker_InitUnsignedConfigNotBool(t *testing.T) {
	doc := newDiscoveryDocument(t, signedCCAPlatformProfile, "json")
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, coservBuilder.DiscoveryMediaTypeJson, r.Header.Get("Accept"))
		w.Header().Set("Content-Type", coservBuilder.DiscoveryMediaTypeJson)
		w.Header().Set("Cache-Control", "max-age=60")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(doc)
	}))
	defer mock.Close()

	params := customParams(t, true, mock.URL, true, "", true, "invalid")

	broker := NewBroker()
	err := broker.Init(params)
	assert.ErrorContains(t, err, "unsigned-coserv config is not bool!")
}

func TestBroker_UnsignedMediaTypeNotFound(t *testing.T) {
	doc := newDiscoveryDocument(t, signedCCAPlatformProfile, "json")
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, coservBuilder.DiscoveryMediaTypeJson, r.Header.Get("Accept"))
		w.Header().Set("Content-Type", coservBuilder.DiscoveryMediaTypeJson)
		w.Header().Set("Cache-Control", "max-age=60")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(doc)
	}))
	defer mock.Close()

	params := customParams(t, true, mock.URL, true, "", true, true)

	broker := NewBroker()
	err := broker.Init(params)
	errMsg := fmt.Sprintf("remote coserv service is unable to serve %s mediatype response", unSignedCCAPlatformProfile)
	assert.ErrorContains(t, err, errMsg)
}

func TestBroker_SignedMediaTypeNotFound(t *testing.T) {
	doc := newDiscoveryDocument(t, unSignedCCAPlatformProfile, nil)
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, coservBuilder.DiscoveryMediaTypeJson, r.Header.Get("Accept"))
		w.Header().Set("Content-Type", coservBuilder.DiscoveryMediaTypeJson)
		w.Header().Set("Cache-Control", "max-age=60")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(doc)
	}))
	defer mock.Close()

	params := customParams(t, true, mock.URL, true, "", true, false)

	broker := NewBroker()
	err := broker.Init(params)
	errMsg := fmt.Sprintf("remote coserv service is unable to serve %s mediatype response", signedCCAPlatformProfile)
	assert.ErrorContains(t, err, errMsg)
}

func TestBroker_NoKeyPresentInDiscovery(t *testing.T) {
	doc := newDiscoveryDocument(t, signedCCAPlatformProfile, nil)
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, coservBuilder.DiscoveryMediaTypeJson, r.Header.Get("Accept"))
		w.Header().Set("Content-Type", coservBuilder.DiscoveryMediaTypeJson)
		w.Header().Set("Cache-Control", "max-age=60")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(doc)
	}))
	defer mock.Close()

	params := customParams(t, true, mock.URL, true, "", true, false)

	broker := NewBroker()
	err := broker.Init(params)
	assert.ErrorContains(t, err, "signed coserv is configured for result but no keys found in discovery document!")
}

func TestBroker_PublicKeyFailureinDiscovery(t *testing.T) {
	testKey := maps.Clone(testES256Key)
	testKey["crv"] = "P-384"
	doc := newDiscoveryDocument(t, signedCCAPlatformProfile, testKey)
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, coservBuilder.DiscoveryMediaTypeJson, r.Header.Get("Accept"))
		w.Header().Set("Content-Type", coservBuilder.DiscoveryMediaTypeJson)
		w.Header().Set("Cache-Control", "max-age=60")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(doc)
	}))
	defer mock.Close()

	params := customParams(t, true, mock.URL, true, "", true, false)

	broker := NewBroker()
	err := broker.Init(params)
	assert.ErrorContains(t, err, "failed to parse JWK key in discovery document!")
}

func TestBroker_InitSuccessUnSigned(t *testing.T) {
	doc := newDiscoveryDocument(t, unSignedCCAPlatformProfile, nil)
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, coservBuilder.DiscoveryMediaTypeJson, r.Header.Get("Accept"))
		w.Header().Set("Content-Type", coservBuilder.DiscoveryMediaTypeJson)
		w.Header().Set("Cache-Control", "max-age=60")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(doc)
	}))
	defer mock.Close()

	params := customParams(t, true, mock.URL, true, "", true, true)

	broker := NewBroker()
	err := broker.Init(params)
	assert.NoError(t, err)

	expectedURI := mock.URL + "/endpoint/{query}"
	assert.Equal(t, expectedURI, broker.coservclient.RequestResponseURI.Raw())
}

//nolint:dupl
func TestBroker_InitSuccessSignedWithJWKKey(t *testing.T) {
	doc := newDiscoveryDocument(t, signedCCAPlatformProfile, "json")
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, coservBuilder.DiscoveryMediaTypeJson, r.Header.Get("Accept"))
		w.Header().Set("Content-Type", coservBuilder.DiscoveryMediaTypeJson)
		w.Header().Set("Cache-Control", "max-age=60")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(doc)
	}))
	defer mock.Close()

	params := customParams(t, true, mock.URL, true, "", true, false)

	broker := NewBroker()
	err := broker.Init(params)
	assert.NoError(t, err)

	expectedURI := mock.URL + "/endpoint/{query}"
	assert.Equal(t, expectedURI, broker.coservclient.RequestResponseURI.Raw())
}

//nolint:dupl
func TestBroker_InitSuccessSignedWithCBORKey(t *testing.T) {
	doc := newDiscoveryDocument(t, signedCCAPlatformProfile, "cbor")
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, coservBuilder.DiscoveryMediaTypeJson, r.Header.Get("Accept"))
		w.Header().Set("Content-Type", coservBuilder.DiscoveryMediaTypeCbor)
		w.Header().Set("Cache-Control", "max-age=60")
		w.WriteHeader(http.StatusOK)
		_ = cbor.NewEncoder(w).Encode(doc)
	}))
	defer mock.Close()

	params := customParams(t, true, mock.URL, true, "", true, false)

	broker := NewBroker()
	err := broker.Init(params)
	assert.NoError(t, err)

	expectedURI := mock.URL + "/endpoint/{query}"
	assert.Equal(t, expectedURI, broker.coservclient.RequestResponseURI.Raw())
}

func TestBroker_GetName(t *testing.T) {
	broker := NewBroker()
	assert.Equal(t, "cca-coserv-broker", broker.GetName())
}

func TestBroker_GetAttestationScheme(t *testing.T) {
	broker := NewBroker()
	assert.Equal(t, "ARM_CCA", broker.GetAttestationScheme())
}

func TestBroker_GetSupportedMediaTypes(t *testing.T) {
	broker := NewBroker()
	mediaTypes := broker.GetSupportedMediaTypes()
	assert.Equal(t, len(mediaTypes), 1)

	expected := []string{
		`application/coserv+cbor; profile="tag:arm.com,2025:cca_platform#1.0.0"`,
	}

	coservTypes, ok := mediaTypes["coserv"]
	assert.True(t, ok)
	assert.Equal(t, coservTypes, expected)
}

func TestBroker_fetchCoservResponseUnsignedFetchFailed(t *testing.T) {
	query := newDefaultQuery(t)
	base64Query, err := query.ToBase64Url()
	require.NoError(t, err)

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	broker, teardown := customBrokerClient(t, h, false)
	defer teardown()

	_, err = broker.ExecuteCoservQuery(SupportedCoservProfile, base64Query)
	assert.ErrorContains(t, err, "unable to Fetch CoSERV response")
}

func TestBroker_fetchCoservResponseSignedFetchFailed(t *testing.T) {
	query := newDefaultQuery(t)
	base64Query, err := query.ToBase64Url()
	require.NoError(t, err)

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	broker, teardown := customBrokerClient(t, h, true)
	defer teardown()

	_, err = broker.ExecuteCoservQuery(SupportedCoservProfile, base64Query)
	assert.ErrorContains(t, err, "unable to Fetch CoSERV response")
}

func TestBroker_ExecuteCoservQueryUnsupportedProfile(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	broker, teardown := customBrokerClient(t, h, true)
	defer teardown()

	_, err := broker.ExecuteCoservQuery("", "")
	assert.ErrorIs(t, err, handler.ErrUnsupported)
}

func TestBroker_ExecuteCoservQuerySignedIncorrectResp(t *testing.T) {
	query := newDefaultQuery(t)
	base64Query, err := query.ToBase64Url()
	require.NoError(t, err)

	resp := []byte{
		0x01, 0x22,
	}

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, coservBuilder.SignedCoSERVMediaType, r.Header.Get("Accept"))
		w.Header().Set("Content-Type", coservBuilder.SignedCoSERVMediaType)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(resp)
	})

	broker, teardown := customBrokerClient(t, h, true)
	defer teardown()

	_, err = broker.ExecuteCoservQuery(SupportedCoservProfile, base64Query)
	assert.ErrorContains(t, err, "invalid COSE_Sign1_Tagged object")
}

func TestBroker_ExecuteCoservQuerySignedCoSERVExtractionFailed(t *testing.T) {
	query := newDefaultQuery(t)
	base64Query, err := query.ToBase64Url()
	require.NoError(t, err)

	resp := newDefaultResponse(t, query)
	respCBOR := signCoSERV(t, resp)
	require.NoError(t, err)

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, coservBuilder.SignedCoSERVMediaType, r.Header.Get("Accept"))
		w.Header().Set("Content-Type", coservBuilder.SignedCoSERVMediaType)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(respCBOR)
	})

	broker, teardown := customBrokerClient(t, h, true)
	defer teardown()

	result, err := broker.ExecuteCoservQuery(SupportedCoservProfile, base64Query)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestBroker_fetchCoservResponseExpired(t *testing.T) {
	query := newDefaultQuery(t)
	base64Query, err := query.ToBase64Url()
	require.NoError(t, err)

	resp := newDefaultResponse(t, query)
	resp.Results.SetExpiry(time.Now().Add(-24 * time.Hour))
	respCBOR, err := resp.ToCBOR()
	require.NoError(t, err)

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, coservBuilder.UnsignedCoSERVMediaType, r.Header.Get("Accept"))
		w.Header().Set("Content-Type", coservBuilder.UnsignedCoSERVMediaType)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(respCBOR)
	})

	broker, teardown := customBrokerClient(t, h, false)
	defer teardown()

	_, err = broker.ExecuteCoservQuery(SupportedCoservProfile, base64Query)
	assert.ErrorContains(t, err, "coserv response has expired")
}

func TestBroker_fetchCoservResponseQueryNotEqual(t *testing.T) {
	query := newDefaultQuery(t)
	respCBOR, err := newDefaultResponse(t, query).ToCBOR()
	require.NoError(t, err)

	rt := coserv.ResultTypeSourceArtifacts
	query.Query.ResultType = &rt
	base64Query, err := query.ToBase64Url()
	require.NoError(t, err)

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, coservBuilder.UnsignedCoSERVMediaType, r.Header.Get("Accept"))
		w.Header().Set("Content-Type", coservBuilder.UnsignedCoSERVMediaType)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(respCBOR)
	})

	broker, teardown := customBrokerClient(t, h, false)
	defer teardown()

	_, err = broker.ExecuteCoservQuery(SupportedCoservProfile, base64Query)
	assert.ErrorContains(t, err, "coserv query is not same as query in response")
}

func TestBroker_ExecuteCoservQueryClientNotInit(t *testing.T) {
	broker := NewBroker()

	_, err := broker.ExecuteCoservQuery(SupportedCoservProfile, "")
	assert.ErrorContains(t, err, "coserv client is not initialized, cannot execute coserv query")
}

func TestBroker_ExecuteCoservQueryBadQuery(t *testing.T) {
	broker := NewBroker()
	broker.coservclient = &coservBuilder.QueryConfig{}

	_, err := broker.ExecuteCoservQuery(SupportedCoservProfile, "invalid")
	assert.ErrorContains(t, err, "decoding unsigned CoSERV query")
}

func TestBroker_ExecuteCoservQueryEmptyResponse(t *testing.T) {
	query := newDefaultQuery(t)
	base64Query, err := query.ToBase64Url()
	require.NoError(t, err)

	resp := newDefaultResponse(t, query)
	resp.Results.AKQ = nil
	resp.Results.RVQ = nil
	respCBOR, err := resp.ToCBOR()
	require.NoError(t, err)

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, coservBuilder.UnsignedCoSERVMediaType, r.Header.Get("Accept"))
		w.Header().Set("Content-Type", coservBuilder.UnsignedCoSERVMediaType)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(respCBOR)
	})

	broker, teardown := customBrokerClient(t, h, false)
	defer teardown()

	_, err = broker.ExecuteCoservQuery(SupportedCoservProfile, base64Query)
	assert.EqualError(t, err, handler.ErrNotFound.Error())
}

func TestBroker_ExecuteCoservQueryUnsignedSuccess(t *testing.T) {
	query := newDefaultQuery(t)
	base64Query, err := query.ToBase64Url()
	require.NoError(t, err)

	respCBOR, err := newDefaultResponse(t, query).ToCBOR()
	require.NoError(t, err)

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, coservBuilder.UnsignedCoSERVMediaType, r.Header.Get("Accept"))
		w.Header().Set("Content-Type", coservBuilder.UnsignedCoSERVMediaType)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(respCBOR)
	})

	broker, teardown := customBrokerClient(t, h, false)
	defer teardown()

	result, err := broker.ExecuteCoservQuery(SupportedCoservProfile, base64Query)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestBroker_ExecuteCoservQuerySignedSuccess(t *testing.T) {
	query := newDefaultQuery(t)
	base64Query, err := query.ToBase64Url()
	require.NoError(t, err)

	resp := newDefaultResponse(t, query)
	respCBOR := signCoSERV(t, resp)
	require.NoError(t, err)

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, coservBuilder.SignedCoSERVMediaType, r.Header.Get("Accept"))
		w.Header().Set("Content-Type", coservBuilder.SignedCoSERVMediaType)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(respCBOR)
	})

	broker, teardown := customBrokerClient(t, h, true)
	defer teardown()

	result, err := broker.ExecuteCoservQuery(SupportedCoservProfile, base64Query)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestBroker_fetchEndorsementsClientNoInit(t *testing.T) {
	broker := NewBroker()
	ta := sampleTrustAnchorIDs(t)

	_, err := broker.GetKeyTriples(&ta, "", true)
	assert.ErrorContains(t, err, "CoSERV Client is not initialized, cannot fetch")
}

func TestBroker_fetchEndorsementsServerError(t *testing.T) {
	ta := sampleTrustAnchorIDs(t)

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	broker, teardown := customBrokerClient(t, h, false)
	defer teardown()

	_, err := broker.GetKeyTriples(&ta, "", true)
	assert.ErrorContains(t, err, "unable to Fetch CoSERV response")
}

func TestBroker_GetKeyTriplesEmptyResponse(t *testing.T) {
	ta := sampleTrustAnchorIDs(t)

	resp := getCoSERVTAResponse(t)
	resp.Results.AKQ = nil
	respCBOR, err := resp.ToCBOR()
	require.NoError(t, err)

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, coservBuilder.UnsignedCoSERVMediaType, r.Header.Get("Accept"))
		w.Header().Set("Content-Type", coservBuilder.UnsignedCoSERVMediaType)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(respCBOR)
	})

	broker, teardown := customBrokerClient(t, h, false)
	defer teardown()

	_, err = broker.GetKeyTriples(&ta, "", true)
	assert.EqualError(t, err, handler.ErrNotFound.Error())
}

func TestBroker_GetKeyTriplesSuccess(t *testing.T) {
	ta := sampleTrustAnchorIDs(t)

	resp := getCoSERVTAResponse(t)
	respCBOR, err := resp.ToCBOR()
	require.NoError(t, err)

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, coservBuilder.UnsignedCoSERVMediaType, r.Header.Get("Accept"))
		w.Header().Set("Content-Type", coservBuilder.UnsignedCoSERVMediaType)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(respCBOR)
	})

	broker, teardown := customBrokerClient(t, h, false)
	defer teardown()

	kts, err := broker.GetKeyTriples(&ta, "", true)
	assert.NoError(t, err)
	assert.NotEqual(t, 0, len(kts))
}

func TestBroker_GetValueTriplesEmptyResponse(t *testing.T) {
	rv := sampleReferenceValueIDs(t)

	resp := getCoSERVRVResponse(t)
	resp.Results.RVQ = nil
	respCBOR, err := resp.ToCBOR()
	require.NoError(t, err)

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, coservBuilder.UnsignedCoSERVMediaType, r.Header.Get("Accept"))
		w.Header().Set("Content-Type", coservBuilder.UnsignedCoSERVMediaType)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(respCBOR)
	})

	broker, teardown := customBrokerClient(t, h, false)
	defer teardown()

	_, err = broker.GetValueTriples(&rv, "", true)
	assert.EqualError(t, err, handler.ErrNotFound.Error())
}

func TestBroker_GetValueTriplesSuccess(t *testing.T) {
	rv := sampleReferenceValueIDs(t)

	resp := getCoSERVRVResponse(t)
	respCBOR, err := resp.ToCBOR()
	require.NoError(t, err)

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, coservBuilder.UnsignedCoSERVMediaType, r.Header.Get("Accept"))
		w.Header().Set("Content-Type", coservBuilder.UnsignedCoSERVMediaType)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(respCBOR)
	})

	broker, teardown := customBrokerClient(t, h, false)
	defer teardown()

	vts, err := broker.GetValueTriples(&rv, "", true)
	assert.NoError(t, err)
	assert.NotEqual(t, 0, len(vts))
}

func TestBroker_AddCorimBytes(t *testing.T) {
	broker := NewBroker()
	err := broker.AddCorimBytes([]byte{}, "", true)
	assert.ErrorIs(t, err, handler.ErrUnsupported)
}

func TestBroker_SetEndorsementsState(t *testing.T) {
	broker := NewBroker()
	err := broker.SetEndorsementsState("", []byte{}, true)
	assert.ErrorIs(t, err, handler.ErrUnsupported)
}

func TestBroker_Fini(t *testing.T) {
	broker := NewBroker()

	ret := broker.Fini()
	assert.Nil(t, ret)
}
