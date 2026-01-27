# Copyright 2023-2024 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0
import os
import json
from datetime import datetime, timedelta
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
    decoded_submods = _extract_submods(response, verifier_key)

    with open(expected) as fh:
        expected_submods = json.load(fh)

    for key, expected_claims in expected_submods.items():
        try:
            decoded_claims = decoded_submods[key]
            print("Key exists in the dictionary.")
        except KeyError:
            print(f"Key {key} does not exist in the dictionary.")

        assert decoded_claims["ear.status"] == expected_claims["ear.status"]
        print(f"Evaluating Submod with SubModName {key}")
        if "ear.appraisal-policy-id" in expected_claims:
            assert decoded_claims["ear.appraisal-policy-id"] ==\
                    expected_claims["ear.appraisal-policy-id"]

        for trust_claim, tc_value in decoded_claims["ear.trustworthiness-vector"].items():
            expected_value = expected_claims["ear.trustworthiness-vector"][trust_claim]
            assert expected_value == tc_value, f'mismatch for claim "{trust_claim}", for {key}'

        if "ear.veraison.annotated-evidence" in expected_claims:
            assert decoded_claims["ear.veraison.annotated-evidence"] == \
            expected_claims["ear.veraison.annotated-evidence"]

        if "ear.veraison.policy-claims" in expected_claims:
            assert decoded_claims["ear.veraison.policy-claims"] == \
                expected_claims["ear.veraison.policy-claims"]


def check_policy(response, active, name, rules_file):
    policy = _extract_policy(response.json())

    _check_within_period(policy['ctime'], timedelta(seconds=60))

    if active is not None:
        assert policy['active'] == active

    if  name:
        assert policy['name'] == name

    assert policy['type'] == 'opa'

    if rules_file:
        with open(rules_file) as fh:
            rules = fh.read()

        assert policy['rules'] == rules


def check_policy_list(response, have_active):
    active_count = 0
    for entry in response.json():
        policy = _extract_policy(entry)
        _check_within_period(policy['ctime'], timedelta(seconds=60))
        if policy['active']:
            active_count += 1

    assert (have_active and active_count == 1) or \
            (not have_active and active_count == 0)


def _check_within_period(dt, period):
    now = datetime.now().replace(tzinfo=dt.tzinfo)
    assert now > dt > (now - period)


def _extract_submods(response, key_file):
    try:
        result = response.json()["result"]
    except KeyError:
        raise ValueError("Did not receive an attestation result.")

    with open(key_file) as fh:
        key = json.load(fh)

    decoded = jwt.decode(result, key=key, algorithms=['ES256'])

    return decoded["submods"]


def _extract_policy(data):
    policy = data
    policy['ctime'] = datetime.fromisoformat(policy['ctime'])
    return policy
