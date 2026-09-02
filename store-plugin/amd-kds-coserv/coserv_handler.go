// Copyright 2025-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package amdkdscoserv

import (
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"time"

	"github.com/veraison/corim/comid"
	"github.com/veraison/corim/coserv"
	"github.com/veraison/services/handler"
	"github.com/veraison/services/plugin"
	vtsstore "github.com/veraison/services/vts/endorsementstore"
)

type CoservProxyHandler struct {
	*vtsstore.StoreCommonParams
}

func constructVcekUrl(instance *coserv.StatefulInstance) string {
	// TODO(paulhowardarm) - deduce the product name and TCB parameters from the
	// stateful environment. Currently assuming 'Milan" and TCB=0.
	return fmt.Sprintf("https://kdsintf.amd.com/vcek/v1/Milan/%x", instance.Instance.Bytes())
}

func getVcekForInstance(instance *coserv.StatefulInstance) ([]byte, error) {
	url := constructVcekUrl(instance)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("HTTP GET error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error response from AMD KDS: %d %s\nResponse body: %s", resp.StatusCode, resp.Status, string(body))
	}

	// Read certificate bytes.
	certBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response error: %v", err)
	}

	return certBytes, nil
}

func (s *CoservProxyHandler) Init(params *plugin.Parameters) error {
	var cfg vtsstore.StoreCommonParams
	if err := (&cfg).FromParams(params); err != nil {
		s.StoreCommonParams = nil
	} else {
		s.StoreCommonParams = &cfg
	}
	return nil
}

func (s CoservProxyHandler) Fini() error {
	return nil
}

func (s CoservProxyHandler) GetName() string {
	return "amd-kds-coserv-proxy-handler"
}

func (s CoservProxyHandler) GetAttestationScheme() string {
	return SchemeName
}

func (s CoservProxyHandler) GetSupportedMediaTypes() map[string][]string {
	return map[string][]string{"coserv": CoservMediaTypes}
}

func (s CoservProxyHandler) addTrustAnchorForInstance(i *coserv.StatefulInstance, results *coserv.ResultSet) error {
	cert, err := getVcekForInstance(i)

	if err != nil {
		return err
	}

	// TODO(paulhowardarm)
	// We need some kind of cert here, representing this plug-in's authority to re-package from NVIDIA CoRIM
	// We probably also need an NVIDIA cert in the chain
	authority := s.FallbackAuthority

	block := &pem.Block{
		Type:    "CERTIFICATE",
		Headers: map[string]string{},
		Bytes:   cert,
	}

	pem := pem.EncodeToMemory(block)

	triple := comid.KeyTriple{
		Environment: comid.Environment{
			Instance: i.Instance,
		},
		VerifKeys: comid.CryptoKeys{
			comid.MustNewPKIXBase64Cert(string(pem)),
		},
	}

	akQuad := coserv.AKQuad{
		Authorities: comid.NewCryptoKeys().Add(authority),
		AKTriple:    &triple,
	}

	results.AddAttestationKeys(akQuad)

	return nil
}

func (s CoservProxyHandler) ExecuteCoservQuery(profile, query string) (*coserv.Coserv, error) {
	if s.StoreCommonParams == nil {
		return nil, errors.New("missing configurations for CoSERV service")
	}

	if !slices.Contains(SupportedCoservProfiles, profile) {
		return nil, handler.ErrUnsupported
	}

	var q coserv.Coserv
	if err := q.FromBase64Url(query); err != nil {
		return nil, err
	}

	if q.Query.EnvironmentSelector == nil {
		return nil, errors.New("AMD CoSERV proxy plug-in only supports environment queries")
	}

	if *q.Query.ArtifactType != coserv.ArtifactTypeTrustAnchors {
		return nil, errors.New("AMD CoSERV proxy plug-in can only provide Trust Anchor artifacts")
	}

	if q.Query.EnvironmentSelector.Groups != nil {
		return nil, errors.New("AMD CoSERV proxy plug-in can only provide for Instance environments, not Groups")
	}

	if q.Query.EnvironmentSelector.Classes != nil {
		return nil, errors.New("AMD CoSERV proxy plug-in can only provide for Instance environments, not Classes")
	}

	if q.Query.EnvironmentSelector.Instances == nil || len(*q.Query.EnvironmentSelector.Instances) == 0 {
		return nil, errors.New("AMD CoSERV proxy plug-in expects at least one Instance environment")
	}

	// Begin with an empty result set
	coservResult := *coserv.NewResultSet()

	// Loop over all of the instance environments in the query, and call the AMD KDS cloud service for each one.
	for _, i := range *q.Query.EnvironmentSelector.Instances {
		err := s.addTrustAnchorForInstance(&i, &coservResult)
		if err != nil {
			return nil, err
		}
	}

	coservResult.SetExpiry(time.Now().Add(s.MaxExpiry))

	// Add all results into the top-level CoSERV object
	err := q.AddResults(coservResult)
	if err != nil {
		return nil, err
	}
	return &q, nil
}

func (o CoservProxyHandler) GetKeyTriples(env *comid.Environment, scheme string, exact bool) ([]*comid.KeyTriple, error) {
	return nil, handler.ErrUnsupported
}

func (o CoservProxyHandler) GetValueTriples(env *comid.Environment, scheme string, exact bool) ([]*comid.ValueTriple, error) {
	return nil, handler.ErrUnsupported
}

func (b CoservProxyHandler) AddCorimBytes(data []byte, scheme string, activate bool) error {
	return handler.ErrUnsupported
}
