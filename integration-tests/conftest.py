#!/usr/bin/env python3

import json
import subprocess
import os.path
import pytest


@pytest.fixture(name='remove_cbor')
def psa_clean_up_cbor_generated_token():
    # Check if my.cbor exists and delete it
    my_cbor = "/test-vectors/provisioning/cbor/psa-good-evidence.cbor"
    exists = os.path.isfile(my_cbor)
    if exists:
        del_success = subprocess.run(["rm /test-vectors/provisioning/cbor/psa-good-evidence.cbor"], shell=True)
        assert del_success.returncode == 0


@pytest.fixture
def psa_generate_good_provisioning_cbor():
    cocli_cmd = """
                cocli comid create --template=$COCLI_TEMPLATES/data/templates/comid-psa-integ-iakpub.json &&
                cocli comid create --template=$COCLI_TEMPLATES/data/templates/comid-psa-refval.json && 
                cocli corim create --template=$COCLI_TEMPLATES/data/templates/corim-full.json --comid=comid-psa-integ-iakpub.cbor --comid=comid-psa-refval.cbor &&
                mv corim-full.cbor /test-vectors/provisioning/cbor &&
                mv comid-psa-integ-iakpub.cbor /test-vectors/provisioning/cbor &&
                mv comid-psa-refval.cbor /test-vectors/provisioning/cbor
    """
    success = subprocess.run(cocli_cmd, shell=True)
    assert success.returncode == 0

@pytest.fixture
def psa_generate_good_evidence():
    evcli_cmd = """
                evcli psa create -c /test-vectors/verification/json/psa-claims-profile-2-integ.json -k /test-vectors/verification/keys/ec-p256.jwk --token=psa-good-evidence.cbor &&
                mv psa-good-evidence.cbor /test-vectors/verification/cbor
    """
    success = subprocess.run(evcli_cmd, shell=True)
    assert success.returncode == 0

@pytest.fixture
def psa_generate_bad_signature_evidence():
    go_cmds = """
              go-psa -p 2 < /test-vectors/verification/json/psa-claims-profile-2-integ.json | xxd -p -r > psa-token.cbor &&
              go-cose-cli -k $EVCLI_TEMPLATES/ec256.json -a ES256 < ./psa-token.cbor
    """

    signature = subprocess.run(go_cmds, shell=True, stdout=subprocess.PIPE)
    assert signature.returncode == 0
    
    curr_signature = bytearray(signature.stdout)
    last_byte = len(curr_signature)-2
    curr_signature[last_byte] =  curr_signature[last_byte] + 1
    string = curr_signature.decode("utf").replace("\n", "")
   
    bad_signage_cmd = "echo " + string + " | xxd -p -r > /psa-token-bad-signage.cbor"
    final_cbor = subprocess.run(bad_signage_cmd, shell=True)
    assert final_cbor.returncode == 0

    clean_cmd = """
                mv /psa-token-bad-signage.cbor /test-vectors/verification/cbor
                rm ./psa-token.cbor
    """
    clean_up = subprocess.run(clean_cmd, shell=True)
    assert clean_up.returncode == 0


@pytest.fixture
def psa_generate_bad_swcomp_evidence():
    template = "/test-vectors/verification/json/psa-claims-profile-2-integ.json"
    bad_swcomp_template = "/test-vectors/verification/json/psa-claims-profile-2-integ-bad-swcomp.json"
    create_new_file = subprocess.run(["cp " + template + " " + bad_swcomp_template], shell=True)
    assert create_new_file.returncode == 0

    with open(bad_swcomp_template,'r+') as file:
        # Load existing data into a dict
        file_data = json.load(file)
        
        if file_data == None:
            return 1

        # Extract and distort BL software component measurement
        BL_measurement = list(file_data["psa-software-components"][0]["measurement-value"])
        BL_measurement[0] = "H"
        final_BL = "".join(BL_measurement)
        file_data["psa-software-components"][0]["measurement-value"] = final_BL

        # Set file's current position at offset.
        file.seek(0)
        # Convert back to json
        json.dump(file_data, file, indent = 4) 

    evcli_cmd = """
                evcli psa create -c /test-vectors/verification/json/psa-claims-profile-2-integ-bad-swcomp.json -k /test-vectors/verification/keys/ec-p256.jwk --token=psa-bad-swcomp-evidence.cbor &&
                mv psa-bad-swcomp-evidence.cbor /test-vectors/verification/cbor
    """

    success = subprocess.run(evcli_cmd, shell=True)
    assert success.returncode == 0