// Copyright 2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package nvidiacoserv

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/veraison/corim/comid"
	"github.com/veraison/corim/corim"
	"github.com/veraison/corim/coserv"
)

// ----- JSON Helper Types -----

// This struct models the API response from a call to the RIM service
type RimServiceResponse struct {
	Id          string `json:"id"`
	Rim         string `json:"rim"`
	Sha256      string `json:"sha256"`
	LastUpdated string `json:"last_updated"`
	RimFormat   string `json:"rim_format"`
	RequestId   string `json:"request_id"`
}

type CoservProxyHandler struct{}

var (
	dummyAuthority = []byte{0xab, 0xcd, 0xef}
)

func (s CoservProxyHandler) GetName() string {
	return "nvidia-coserv-proxy-handler"
}

func (s CoservProxyHandler) GetAttestationScheme() string {
	return SchemeName
}

func (s CoservProxyHandler) GetSupportedMediaTypes() []string {
	return CoservMediaTypes
}

func callRimService(rimid *string) (*RimServiceResponse, error) {
	// Create an HTTP request to the NVIDIA RIM service.
	urlStr := fmt.Sprintf("https://rim.attestation.nvidia.com/v1/rim/%s", *rimid)

	resp, err := http.Get(urlStr)
	if err != nil {
		return nil, fmt.Errorf("HTTP GET error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error response from NVIDIA RIM service: %d %s\nResponse body: %s", resp.StatusCode, resp.Status, string(body))
	}

	// Read the successful response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response error: %v", err)
	}

	// Successful responses are JSON, modelled by the RimServiceResponse class,
	// so deserialize into that.
	var rimServiceResponse RimServiceResponse
	json.Unmarshal(body, &rimServiceResponse)

	return &rimServiceResponse, nil
}

func (s CoservProxyHandler) addReferenceValuesForClass(query *coserv.Query, c *coserv.StatefulClass, results *coserv.ResultSet) error {
	if *c.Class.Vendor != "NVIDIA" {
		return fmt.Errorf("vendors other than NVIDIA not supported by this proxy plug-in")
	}

	if c.Measurements != nil {
		return fmt.Errorf("NVIDIA CoSERV proxy plug-in does not expect stateful class environments")
	}

	if c.Class.Model == nil {
		return fmt.Errorf("no Model field supplied in NVIDIA class environment")
	}

	rimid := c.Class.Model

	rimServiceResponse, err := callRimService(rimid)

	if err != nil {
		return err
	}

	// The actual RIM is a base64-encoded field, so pull this out as a byte array.
	rimBytes, err := base64.StdEncoding.DecodeString(rimServiceResponse.Rim)

	if err != nil {
		return fmt.Errorf("failed to base64 decode the RIM byte string %s", rimServiceResponse.Rim)
	}

	// If we want the raw source artifacts, add the RIM verbatim into the result set.
	if query.ResultType == coserv.ResultTypeBoth || query.ResultType == coserv.ResultTypeSourceArtifacts {
		// TODO(paulhowardarm) - Figure out how to turn rimBytes into CMW and AddSourceArtifacts here
		// results.AddSourceArtifacts(...)
	}

	// If we want collected artifacts, we need to parse the CORIM
	if query.ResultType == coserv.ResultTypeBoth || query.ResultType == coserv.ResultTypeCollectedArtifacts {
		if rimServiceResponse.RimFormat != "CORIM" {
			return fmt.Errorf("CoSERV proxy plug-in cannot produce collected artifacts for non-CORIM format %s", rimServiceResponse.RimFormat)
		}

		// We now know that the format is CORIM, so it should be possible to decode as such
		var scorim corim.SignedCorim
		err = scorim.FromCOSE(rimBytes)
		if err != nil {
			return fmt.Errorf("failed to parse COSE: %v", err)
		}

		// Loop over the tags in the CoRIM payload.
		for _, t := range scorim.UnsignedCorim.Tags {
			cborTag, cborData := t.Number, t.Content

			// We'll just look at CoMID tags (and NVIDIA RIMs only contain these anyway)
			if cborTag == corim.ComidTag {
				var c comid.Comid
				err = c.FromCBOR(cborData)

				if err != nil {
					return fmt.Errorf("failed to populate CoMID from CBOR: %v", err)
				}

				// We'll just look at reference value triples in the CoMID
				for _, triple := range c.Triples.ReferenceValues.Values {
					// Turn each triple into a quad
					// TODO(paulhowardarm) - This authority is a dummy value.
					// We need some kind of cert here, representing this plug-in's authority to re-package from NVIDIA CoRIM
					// We probably also need an NVIDIA cert in the chain
					authority, err := comid.NewCryptoKeyTaggedBytes(dummyAuthority)

					if err != nil {
						return fmt.Errorf("failed to make authority tagged bytes: %v", err)
					}

					rvQuad := coserv.RefValQuad{
						Authorities: &[]comid.CryptoKey{*authority},
						RVTriple:    &triple,
					}

					results.AddReferenceValues(rvQuad)
				}
			}
		}
	}

	return nil
}

func (s CoservProxyHandler) GetEndorsements(tenantID string, query string) ([]byte, error) {
	var q coserv.Coserv
	if err := q.FromBase64Url(query); err != nil {
		return nil, err
	}

	if q.Query.ArtifactType != coserv.ArtifactTypeReferenceValues {
		return nil, fmt.Errorf("NVIDIA CoSERV proxy plug-in can only provide Reference Value artifacts")
	}

	if q.Query.EnvironmentSelector.Groups != nil {
		return nil, fmt.Errorf("NVIDIA CoSERV proxy plug-in can only provide for Class environments, not Groups")
	}

	if q.Query.EnvironmentSelector.Instances != nil {
		return nil, fmt.Errorf("NVIDIA CoSERV proxy plug-in can only provide for Class environments, not Instances")
	}

	if q.Query.EnvironmentSelector.Classes == nil || len(*q.Query.EnvironmentSelector.Classes) == 0 {
		return nil, fmt.Errorf("NVIDIA CoSERV proxy plug-in expects at least one Class environment")
	}

	// Begin with an empty result set
	coservResult := *coserv.NewResultSet()

	// Loop over all of the class environments in the query, and call the NVIDIA RIM cloud service for each one.
	for _, c := range *q.Query.EnvironmentSelector.Classes {
		err := s.addReferenceValuesForClass(&q.Query, &c, &coservResult)
		if err != nil {
			return nil, err
		}
	}

	// Set expiry on the results - fairly arbitrary expiry time of 1 hour
	coservResult.SetExpiry(time.Now().Add(time.Hour))

	// Add all results into the top-level CoSERV object
	q.AddResults(coservResult)

	return q.ToCBOR()
}
