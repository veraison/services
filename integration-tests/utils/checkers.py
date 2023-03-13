import os
import json
from jose import jwt


GENDIR = '__generated__'


def save_result(response, scheme, evidence):
    os.makedirs(f'{GENDIR}/results', exist_ok=True)
    outfile = f'{GENDIR}/results/{scheme}.{evidence}.jwt'

    try:
        result = response.json()["result"]
    except KeyError:
        raise ValueError("Did not receive an attestation result.")

    with open(outfile, 'w') as wfh:
        wfh.write(result)


def expected_result(response, expected, verifier_key):
    decoded = _extract_appraisal(response, verifier_key)

    with open(expected) as fh:
        expected_claims = json.load(fh)

    assert decoded["ear.status"] == expected_claims["ear.status"]

    for trust_claim, tc_value in decoded["ear.trustworthiness-vector"].items():
        expected_value = expected_claims["ear.trustworthiness-vector"][trust_claim]
        assert expected_value == tc_value, f'mismatch for claim "{trust_claim}"'

    assert decoded["ear.veraison.annotated-evidence"] == expected_claims["ear.veraison.annotated-evidence"]


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


