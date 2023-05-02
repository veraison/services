package capability

import (
	"github.com/lestrrat-go/jwx/v2/jwk"
)

const (
	WellKnownMediaType = "application/vnd.veraison.discovery+json"
)

type TEE struct {
	Name       string `json:"tee-name"`
	EvidenceId string `json:"evidence-id"`
	Evidence   []byte `json:"evidence"`
}

func NewTEE(name string, evidenceID string, evidence []byte) (*TEE, error) {
	return &TEE{
		Name:       name,
		EvidenceId: evidenceID,
		Evidence:   evidence,
	}, nil
}

type WellKnownInfo struct {
	PublicKey    jwk.Key           `json:"ear-verification-key,omitempty"`
	Tee          *TEE              `json:"ear-tee,omitempty"`
	MediaTypes   []string          `json:"media-types"`
	Version      string            `json:"version"`
	ServiceState string            `json:"service-state"`
	ApiEndpoints map[string]string `json:"api-endpoints"`
}

func NewWellKnownInfoObj(
	key jwk.Key, mediaTypes []string, version string, serviceState string,
	endpoints map[string]string, tee *TEE,
) (*WellKnownInfo, error) {
	obj := &WellKnownInfo{
		PublicKey:    key,
		MediaTypes:   mediaTypes,
		Version:      version,
		ServiceState: serviceState,
		ApiEndpoints: endpoints,
	}

	if tee != nil {
		obj.Tee = tee
	}

	return obj, nil
}
