package capability

import (
	"github.com/lestrrat-go/jwx/v2/jwk"
)

const (
	WellKnownMediaType = "application/vnd.veraison.discovery+json"
)

type WellKnownInfo struct {
	PublicKey    jwk.Key           `json:"ear-verification-key,omitempty"`
	MediaTypes   []string          `json:"media-types"`
	Version      string            `json:"version"`
	ServiceState string            `json:"service-state"`
	ApiEndpoints map[string]string `json:"api-endpoints"`
}

func NewWellKnownInfoObj(key jwk.Key, mediaTypes []string, version string, serviceState string, endpoints map[string]string) (*WellKnownInfo, error) {
	obj := &WellKnownInfo{
		PublicKey:    key,
		MediaTypes:   mediaTypes,
		Version:      version,
		ServiceState: serviceState,
		ApiEndpoints: endpoints,
	}

	return obj, nil
}
