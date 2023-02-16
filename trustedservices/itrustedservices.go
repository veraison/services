package trustedservices

import (
	"github.com/veraison/services/proto"
)

// Aliasing these here, as they are part of the TrustedServices interface,
// and we should avoid leaking the implementation detail that they're  actually
// implemented via protobufs.
type Endorsement proto.Endorsement
type AttestationToken proto.AttestationToken
type AppraisalContext proto.AppraisalContext

type ITrustedServices interface {
	GetAttestation(token *AttestationToken) (*AppraisalContext, error)
	GetSupportedVerificationMediaTypes() ([]string, error)
	AddRefValues(referenceValues []*Endorsement) error
	AddTrustAnchor(trustAnchors []*Endorsement) error
}
