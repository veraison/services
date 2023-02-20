package trustedservices

import (
	"github.com/veraison/services/proto"
)


type EndorsementType int32

const (
	Unset            EndorsementType = iota
	ReferenceValue
	VerificationKey
)

// Aliasing these here, as they are part of the TrustedServices interface,
// and we should avoid leaking the implementation detail that they're  actually
// implemented via protobufs.
type Endorsement {
	Scheme string
	Type EndorsementType

}
type AppraisalContext proto.AppraisalContext

type AttestationToken struct {
	TenantId  int
	Data      []byte
	MediaType string
}

type ITrustedServices interface {
	GetAttestation(token *AttestationToken) (*AppraisalContext, error)
	GetSupportedVerificationMediaTypes() ([]string, error)
	AddRefValues(referenceValues []*Endorsement) error
	AddTrustAnchor(trustAnchors []*Endorsement) error
}
