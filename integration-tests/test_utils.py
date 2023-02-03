#!/usr/bin/env python3

import json
import subprocess
from jose import jwt


def write_json(new_data, filename):
    with open(filename,'r+') as file:
        # Load existing data into a dict
        file_data = json.load(file)
        
        if file_data == None:
            return 1

        # Join new_data with file_data
        file_data.update(new_data)
        # Set file's current position at offset.
        file.seek(0)
        # Convert back to json
        json.dump(file_data, file, indent = 4) 

        return 0


def generate_token(response):
    # Obtain nonce value from resource creation response 
    nonce_val = {"psa-nonce": response.json()["nonce"]}
    
    # Create new json template
    template_without_nonce = "/test-vectors/verification/json/psa-claims-profile-2-integ-without-nonce.json"
    template_with_nonce = "/test-vectors/verification/json/psa-claims-profile-2-integ-with-nonce.json"
    create_new_file = subprocess.run(["cp " + template_without_nonce + " " + template_with_nonce], shell=True)
    assert create_new_file.returncode == 0

    # Add nonce to json template
    success_nonce = write_json(nonce_val, template_with_nonce)
    assert success_nonce == 0

    # Generate token with evcli using template with nonce and key
    evcli_create = subprocess.run(["evcli psa create --claims=" + template_with_nonce + " --key=/test-vectors/verification/keys/ec-p256.jwk --token=/test-vectors/verification/cbor/attester.cbor"], shell=True)
    assert evcli_create.returncode == 0

def decode_attestation_result(response, key):
    currJWT = response.json()["result"]

    with open(key,'r+') as key_file:
        ecdsa_key = json.load(key_file)
    
    ecdsa_key.pop("d")
    decoded = jwt.decode(currJWT, key=ecdsa_key, algorithms=['ES256'])
    
    return decoded

def verify_good_attestation_results(response, template, key):
    decoded = decode_attestation_result(response, key)

    with open(template,'r+') as file:
        # Load existing data into a dict
        file_data = json.load(file)
        
        if file_data == None:
            return 1

    assert decoded["ear.status"] == "affirming"

    # TODO Tighten conditions when there are references for the trustworthiness vector
    for vec in decoded["ear.trustworthiness-vector"]:
        assert decoded["ear.trustworthiness-vector"][vec] == 0 or decoded["ear.trustworthiness-vector"][vec] == 2 or decoded["ear.trustworthiness-vector"][vec] == 3

    for key in decoded["ear.veraison.processed-evidence"]:
        assert decoded["ear.veraison.processed-evidence"][key] == file_data[key]

# TODO add this function to the relevant tavern test. See example in test_end_to_end_success.tavern.yaml (line 84-88)
def verify_bad_signature_attestation_results(response, template, key):
    decoded = decode_attestation_result(response, key)

    with open(template,'r+') as file:
        # Load existing data into a dict
        file_data = json.load(file)
        
        if file_data == None:
            return 1

    assert decoded["ear.status"] == "warning"

    for trust_claim in decoded["ear.trustworthiness-vector"]:
        assert decoded["ear.trustworthiness-vector"][trust_claim] == 99

    for key in decoded["ear.veraison.processed-evidence"]:
        assert decoded["ear.veraison.processed-evidence"][key] == file_data[key]


def verify_bad_swcomp_attestation_results(response, template, key):
    decoded = decode_attestation_result(response, key)

    with open(template,'r+') as file:
        # Load existing data into a dict
        file_data = json.load(file)
        
        if file_data == None:
            return 1

    assert decoded["ear.status"] == "warning"

    assert decoded["ear.trustworthiness-vector"]["executables"] == 33

    for key in decoded["ear.veraison.processed-evidence"]:
        assert decoded["ear.veraison.processed-evidence"][key] == file_data[key]
    