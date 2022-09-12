package policy

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/veraison/services/config"
	"github.com/veraison/services/proto"
)

type PolicyAgent struct {
	Backend IBackend
}

func (o *PolicyAgent) Init(cfg config.Store) error {
	if err := o.Backend.Init(cfg); err != nil {
		return err
	}

	return nil
}

// GetBackendName returns a string containing the name of the backend used by
// the agend.
func (o *PolicyAgent) GetBackendName() string {
	return o.Backend.GetName()
}

// Evaluate the provided policy w.r.t. to the specified evidence and
// endorsements, and return an updated AttestationResult. The policy may
// overwrite the result status or any of the values in the result trust vector.
func (o *PolicyAgent) Evaluate(
	ctx context.Context,
	policy *Policy,
	result *proto.AttestationResult,
	evidence *proto.EvidenceContext,
	endorsements []string,
) (*proto.AttestationResult, error) {
	resultBytes, err := result.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("could not marshal provided result: %w", err)
	}

	var resultMap map[string]interface{}
	if err = json.Unmarshal(resultBytes, &resultMap); err != nil {
		return nil, fmt.Errorf("could not unmarshal provided result: %w", err)
	}

	updatedByPolicy, err := o.Backend.BackEndEvaluate(
		ctx,
		policy.Rules,
		resultMap,
		evidence.Evidence.AsMap(),
		endorsements,
	)
	if err != nil {
		return nil, fmt.Errorf("could not evaluate policy: %w", err)
	}

	// TODO(setrofim): at this stage, we have the opportunity to log or
	// otherwise communicate/identify the changes to the AttestationResult
	// made by policy, if we want each entry in the result to have a
	// clearly-traceable origin.

	updatedStatus, ok := updatedByPolicy["status"]
	if !ok {
		return nil, fmt.Errorf(ErrNoStatus, updatedByPolicy)
	}

	if updatedStatus != "" {
		resultMap["status"] = updatedByPolicy["status"]
	}

	updatedTV, ok := updatedByPolicy["trust-vector"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf(ErrNoTV, updatedByPolicy)
	}

	for k, v := range updatedTV {
		if v != "" {
			resultMap["trust-vector"].(map[string]interface{})[k] = v
		}
	}

	evalBytes, err := json.Marshal(resultMap)
	if err != nil {
		return nil, fmt.Errorf("could not marshal updated result: %w", err)
	}

	var evaluatedResult proto.AttestationResult

	if err = evaluatedResult.UnmarshalJSON(evalBytes); err != nil {

		return nil, fmt.Errorf(ErrBadResult, err, evalBytes)
	}

	evaluatedResult.AppraisalPolicyID = policy.ID

	return &evaluatedResult, nil
}

func (o *PolicyAgent) GetBackEnd() IBackend {
	return o.Backend
}

func (o *PolicyAgent) SetBackEnd(b IBackend) {
	o.Backend = b
}

func (o *PolicyAgent) Close() {
	o.Backend.Close()
}
