package capability

import (
	"github.com/lestrrat-go/jwx/v2/jwk"
)

type WellKnownInfo struct {
	PublicKey    jwk.Key           `json:"ear-verification-key,omitempty"`
	MediaTypes   []string          `json:"media-types"`
	Version      string            `json:"version"`
	State        string            `json:"state"`
	ApiEndpoints map[string]string `json:"api-endpoints"`
}

func NewWellKnownInfoObj(key jwk.Key, mediaTypes []string, version string, state string, endpoints map[string]string) (*WellKnownInfo, error) {
	obj := &WellKnownInfo{
		PublicKey:    key,
		MediaTypes:   mediaTypes,
		Version:      version,
		State:        state,
		ApiEndpoints: endpoints,
	}

	return obj, nil
}
