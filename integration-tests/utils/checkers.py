import os
import json
from jose import jwt

GENDIR = '__generated__'


def save_result(response, scheme, evidence):
    os.makedirs(f'{GENDIR}/results', exist_ok=True)
    jwt_outfile = f'{GENDIR}/results/{scheme}.{evidence}.jwt'

    try:
        result = response.json()["result"]
    except KeyError:
        raise ValueError("Did not receive an attestation result.")

    with open(jwt_outfile, 'w') as wfh:
        wfh.write(result)

    decoded = jwt.decode(result, key="", options={'verify_signature': False})

    json_outfile = f'{GENDIR}/results/{scheme}.{evidence}.json'
    with open(json_outfile, 'w') as wfh:
        json.dump(decoded, wfh, indent=4)


def compare_to_expected_result(response, expected, verifier_key):
    decoded = _extract_appraisal(response, verifier_key)

    with open(expected) as fh:
        expected_claims = json.load(fh)

    assert decoded["ear.status"] == expected_claims["ear.status"]

    if "ear.appraisal-policy-id" in expected_claims:
        assert decoded["ear.appraisal-policy-id"] ==\
                expected_claims["ear.appraisal-policy-id"]

    for trust_claim, tc_value in decoded["ear.trustworthiness-vector"].items():
        expected_value = expected_claims["ear.trustworthiness-vector"][trust_claim]
        assert expected_value == tc_value, f'mismatch for claim "{trust_claim}"'

    if "ear.veraison.annotated-evidence" in expected_claims:
        assert decoded["ear.veraison.annotated-evidence"] == \
                expected_claims["ear.veraison.annotated-evidence"]

    if "ear.veraison.policy-claims" in expected_claims:
        assert decoded["ear.veraison.policy-claims"] == \
                expected_claims["ear.veraison.policy-claims"]


def _extract_appraisal(response, key_file):
    try:
        result = response.json()["result"]
    except KeyError:
        raise ValueError("Did not receive an attestation result.")

    with open(key_file) as fh:
        key = json.load(fh)

    decoded = jwt.decode(result, key=key, algorithms=['ES256'])

    num_submods = len(decoded["submods"])
    if num_submods != 1:
        raise ValueError(f"Unexpected number of submods in result. Wanted 1, found {num_submods}.")

    return decoded["submods"].popitem()[1]


