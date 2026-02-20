// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package appraisal

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/veraison/corim/comid"
	"github.com/veraison/ear"
	"github.com/veraison/services/config"
	"github.com/veraison/services/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// Context is the appraisal context for an attestation. It maintains the state
// of the attestation through the pipeline, between the calls to the
// scheme-specific functionality.
type Context struct {
	Scheme            string                 `json:"scheme"`
	TrustAnchorIDs    []*comid.Environment   `json:"trust-anchor-ids"`
	ReferenceValueIDs []*comid.Environment   `json:"reference-value-ids"`
	Evidence          *Evidence              `json:"evidence"`
	Claims            map[string]any         `json:"claims"`
	Result            *ear.AttestationResult `json:"result"`
	SignedEAR         []byte                 `json:"signed-ear"`
}

// NewContext instantiates a new Context using the provided evidence. The
// AttestationResult result inside the context is initialized with a single
// submod called "ERROR" (this is replaced when the attestation scheme is
// known).
func NewContext(evidence *Evidence) *Context {
	ac := Context{
		Evidence: evidence,
		Claims:   make(map[string]any),
		// "ERROR" is used for the default submod name as we don't know
		// the attestation scheme yet; the submod will be moved under the scheme
		// name at the time the scheme is set.
		Result: ear.NewAttestationResult("ERROR", config.Version, config.Developer),
	}

	encodedNonce := base64.URLEncoding.EncodeToString(evidence.Nonce)
	ac.Result.Nonce = &encodedNonce

	return &ac
}

// SetScheme sets the scheme used in the attestation tracked by this Context.
// This updates the submod in the AttestationResult from "ERROR" to the name of
// the scheme.
func (o *Context) SetScheme(scheme string) error {
	var ok bool
	if _, ok = o.Result.Submods[scheme]; ok {
		return fmt.Errorf("submod %q already exists in result", scheme)
	}

	o.Scheme = scheme
	// now that the scheme is known, move the default submod (which is
	// currently under "ERROR") to under the scheme's name.
	o.Result.Submods[scheme], ok = o.Result.Submods["ERROR"]
	if !ok {
		return errors.New("submod \"ERROR\" not in result; has the scheme already been set?")
	}
	delete(o.Result.Submods, "ERROR")

	o.InitPolicyID()

	return nil
}

// StoreLabel returns the label that should be used when querying the CoRIMs
// store for items associated with attestation tracked by this context.
func (o *Context) StoreLabel() string {
	return fmt.Sprintf("%s/%s", o.Evidence.TenantID, o.Scheme)
}

// SetAllClaims sets all claims in all submods in the AttestationResult tracked
// by this Context to the specified value.
func (o *Context) SetAllClaims(claim ear.TrustClaim) {
	for _, submod := range o.Result.Submods {
		submod.TrustVector.SetAll(claim)
	}
}

// AddPolicyClaim adds the specified claim to all submods in the
// AttestationResult tracked by this Context. (Note: this is primarily intended
// for problem reporting, as there aren't many other claims that it would make
// sense to set for all submods).
func (o *Context) AddPolicyClaim(name, claim string) {
	for _, submod := range o.Result.Submods {
		if submod.AppraisalExtensions.VeraisonPolicyClaims == nil {
			claimsMap := make(map[string]any)
			submod.AppraisalExtensions.VeraisonPolicyClaims = &claimsMap
		}
		(*submod.AppraisalExtensions.VeraisonPolicyClaims)[name] = claim
	}
}

// InitPolicyID initializes the AppraisalPolicyID in the AttestationResult
// tracked by this Context to value based on the attestation scheme.
func (o *Context) InitPolicyID() {
	for _, submod := range o.Result.Submods {
		policyID := fmt.Sprintf("policy:%s", o.Scheme)
		submod.AppraisalPolicyID = &policyID
	}
}

// UpdatePolicyID updates AppraisalPolicyID in all submods with the specified
// element that gets joined with "/" to the existing AppraisalPolicyID value.
func (o *Context) UpdatePolicyID(polID string) error {
	for _, submod := range o.Result.Submods {
		updatedID := strings.Join([]string{*submod.AppraisalPolicyID, polID}, "/")
		submod.AppraisalPolicyID = &updatedID
	}

	return nil
}

// DescribeTrustAnchorIDs returns a string value describing the trust anchor
// IDs tracked by this Context.
func (o *Context) DescribeTrustAnchorIDs() string {
	buf, _ := json.Marshal(o.TrustAnchorIDs)
	return string(buf)
}

// ToProtobuf converts this Context to proto.AppraisalContext
func (o *Context) ToProtobuf() (*proto.AppraisalContext, error) {
	trustAnchorIDStrings := make([]string, len(o.TrustAnchorIDs))
	for i, taID := range o.TrustAnchorIDs {
		taJSON, err := json.Marshal(taID)
		if err != nil {
			return nil, err
		}

		trustAnchorIDStrings[i] = string(taJSON)
	}

	referenceValueIDStrings := make([]string, len(o.ReferenceValueIDs))
	for i, rvID := range o.ReferenceValueIDs {
		rvJSON, err := json.Marshal(rvID)
		if err != nil {
			return nil, err
		}

		referenceValueIDStrings[i] = string(rvJSON)
	}

	pbClaims, err := claimsToStruct(o.Claims)
	if err != nil {
		return nil, err
	}

	return &proto.AppraisalContext{
		Evidence: &proto.EvidenceContext{
			TenantId:       o.Evidence.TenantID,
			TrustAnchorIds: trustAnchorIDStrings,
			ReferenceIds:   referenceValueIDStrings,
			Evidence:       pbClaims,
		},
		Result: o.SignedEAR,
	}, nil
}

func claimsToStruct(m map[string]any) (*structpb.Struct, error) {
	// annoyingly, structpb can't handle a, e.g., []int64, though
	// it can handle []any that contains int64's,so we need to
	// do this normalization;
	normalized := make(map[string]any)
	for k, v := range m {
		normalized[k] = normalize(v)
	}

	return structpb.NewStruct(normalized)
}

func normalize(v any) any {
	val := reflect.ValueOf(v)

	switch val.Kind() {
	case reflect.Slice:
		normalizedSlice := reflect.MakeSlice(
			reflect.SliceOf(reflect.TypeFor[any]()),
			val.Len(),
			val.Cap(),
		)

		for i := 0; i < val.Len(); i++ {
			normalizedSlice.Index(i).Set(reflect.ValueOf(normalize(val.Index(i).Interface())))
		}

		return normalizedSlice.Interface()
	case reflect.Map:
		normalizedMap := reflect.MakeMap(
			reflect.MapOf(
				reflect.TypeFor[string](),
				reflect.TypeFor[any](),
			),
		)

		for _, key := range val.MapKeys() {
			normalizedMap.SetMapIndex(key, reflect.ValueOf(normalize(val.MapIndex(key).Interface())))
		}

		return normalizedMap.Interface()
	default:
		return v
	}
}
