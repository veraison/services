#!/usr/bin/env python3

import json
import subprocess
import jwt

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
    
    # Add nonce to json template
    success_nonce = write_json(nonce_val, './integration-tests/verification/psa-claims-profile-2-integ-without-nonce.json')

    # Generate token with evcli using template with nonce and key
    evcli_create = subprocess.run(['evcli psa create --claims=./integration-tests/verification/psa-claims-profile-2-integ-without-nonce.json --key=./integration-tests/verification/ec-p256.jwk --token=./integration-tests/verification/my.cbor'], shell=True)

    assert success_nonce == 0 and evcli_create.returncode == 0

def verify_attestation_results(response, template):
    currJWT = response.json()["result"]
    decoded = jwt.decode(currJWT, options={"verify_signature": False})

    # template = './integration-tests/verification/psa-claims-profile-2-integ.json'
    with open(template,'r+') as file:
        # Load existing data into a dict
        file_data = json.load(file)
        
        if file_data == None:
            return 1

    assert decoded["ear.status"] == "affirming"

    for key in decoded["ear.veraison.processed-evidence"]:
        assert decoded["ear.veraison.processed-evidence"][key] == file_data[key]

    