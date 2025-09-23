# Copyright 2023-2024 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0
import ast
import json
import os
import shutil

from util import update_json, run_command


GENDIR = '__generated__'


def generate_endorsements(test):
    os.makedirs(f'{GENDIR}/endorsements', exist_ok=True)

    scheme = test.test_vars['scheme']
    spec = test.test_vars['endorsements']

    if isinstance(spec, str):
        tag = spec
        spec = test.common_vars['endorsements'][spec]
    else:
        tag = spec[0]

    corim_template_name = 'corim-{}-{}.json'.format(scheme, spec[0])
    corim_template = f'data/endorsements/{corim_template_name}'
    comid_templates = ['data/endorsements/comid-{}-{}.json'.format(scheme, c)
                       for c in spec[1:]]
    output_path = f'{GENDIR}/endorsements/corim-{scheme}-{tag}.cbor'

    generate_corim(corim_template, comid_templates, output_path)

def generate_cca_end_to_end_endorsements(test):
    os.makedirs(f'{GENDIR}/endorsements', exist_ok=True)

    scheme = test.test_vars['scheme']
    spec = test.test_vars['endorsements']
    corim_type = test.test_vars.get('corim_type', 'unsigned')
    
    # first construct platform templates
    corim_template_name = 'corim-{}-platform-{}.json'.format(scheme, spec)
    corim_template = f'data/endorsements/{corim_template_name}'
    tag = ["refval", "ta"]
    comid_templates = ['data/endorsements/comid-{}-{}.json'.format(scheme, c)
                       for c in tag[0:]]
    
    # Generate unsigned CoRIM
    unsigned_output_path = f'{GENDIR}/endorsements/corim-{scheme}-platform-{spec}.cbor'
    generate_corim(corim_template, comid_templates, unsigned_output_path)
    
    # If signed CoRIM is requested, sign it
    if corim_type == 'signed':
        # sign the CoRIM
        signed_output_path = f'{GENDIR}/endorsements/corim-{scheme}-platform-{spec}.signed.cbor'
        sign_corim(unsigned_output_path, signed_output_path)
        # Use the signed CoRIM as the output
        shutil.copyfile(signed_output_path, unsigned_output_path)

    # next realm templates
    corim_template_name = 'corim-{}-realm-{}.json'.format(scheme, spec)
    corim_template = f'data/endorsements/{corim_template_name}'
    tag = ["refval"]
    comid_templates = ['data/endorsements/comid-{}-{}.json'.format(scheme, c)
                       for c in tag[0:]]
    
    # Generate unsigned CoRIM
    unsigned_output_path = f'{GENDIR}/endorsements/corim-{scheme}-realm-{spec}.cbor'
    generate_corim(corim_template, comid_templates, unsigned_output_path)

    if corim_type == 'signed':
        # sign the CoRIM
        signed_output_path = f'{GENDIR}/endorsements/corim-{scheme}-realm-{spec}.signed.cbor'
        sign_corim(unsigned_output_path, signed_output_path)
        # Use the signed CoRIM as the output
        shutil.copyfile(signed_output_path, unsigned_output_path)


def generate_artefacts_from_response(response, scheme, evidence, signing, keys, expected):
    generate_evidence_from_response(response, scheme, evidence, signing, keys)
    generate_expected_result_from_response(response, scheme, expected)


def generate_expected_result_from_response(response, scheme, expected):
    os.makedirs(f'{GENDIR}/expected', exist_ok=True)

    infile = f'data/results/{scheme}.{expected}.json'
    outfile = f'{GENDIR}/expected/{scheme}.{expected}.server-nonce.json'
    nonce = response.json()["nonce"]

    if scheme == 'psa' and nonce:
        update_json(
                infile,
                {"PSA_IOT": {'ear.veraison.annotated-evidence': {f'psa-nonce': nonce}}},
                outfile,
                )
    elif scheme == 'cca' and nonce:
        update_json(
                infile,
                {"CCA_REALM": {'ear.veraison.annotated-evidence': {f'cca-realm-challenge': nonce}}},
                outfile,
                )
    else:
        shutil.copyfile(infile, outfile)


def generate_evidence_from_test(test):
    scheme = test.test_vars['scheme']
    evidence = test.test_vars['evidence']
    nonce = test.common_vars[test.test_vars['nonce']]['value']
    signing = test.common_vars['keys'][test.test_vars['signing']]
    outname = f'{scheme}.{evidence}'

    return generate_evidence(scheme, evidence, nonce, signing, outname)


def generate_evidence_from_response(response, scheme, evidence, signing, keys):
    nonce = response.json()["nonce"]
    actual_signing = ast.literal_eval(keys)[signing]
    outname = f'{scheme}.{evidence}.server-nonce'

    return generate_evidence(scheme, evidence, nonce, actual_signing, outname)


def generate_evidence(scheme, evidence, nonce, signing, outname):
    os.makedirs(f'{GENDIR}/evidence', exist_ok=True)
    os.makedirs(f'{GENDIR}/claims', exist_ok=True)

    if scheme == 'psa' and nonce:
        claims_file = f'{GENDIR}/claims/{scheme}.{evidence}.json'
        # convert nonce from base64url to base64
        translated_nonce = nonce.replace('-', '+').replace('_', '/')
        update_json(
                f'data/claims/{scheme}.{evidence}.json',
                {f'{scheme}-nonce': translated_nonce},
                claims_file,
                )
    elif scheme == 'cca' and nonce:
        claims_file = f'{GENDIR}/claims/{scheme}.{evidence}.json'
        # convert nonce from base64url to base64
        translated_nonce = nonce.replace('-', '+').replace('_', '/')
        update_json(
                f'data/claims/{scheme}.{evidence}.json',
                {'cca-realm-delegated-token': {f'cca-realm-challenge': translated_nonce}},
                claims_file,
                )
    else:
        claims_file = f'data/claims/{scheme}.{evidence}.json'

    if scheme == 'psa':
        iak = signing
        generate_psa_evidence_token(
                claims_file,
                f'data/keys/{iak}.jwk',
                f'{GENDIR}/evidence/{outname}.cbor',
                )
    elif scheme == 'cca':
        iak, rak = signing
        generate_cca_evidence_token(
                claims_file,
                f'data/keys/{iak}.jwk',
                f'data/keys/{rak}.jwk',
                f'{GENDIR}/evidence/{outname}.cbor',
                )
    elif scheme == 'enacttrust':
        key = signing
        badnode = True if 'badnode' in evidence else False
        generate_enacttrust_evidence_token(
                claims_file,
                f'data/keys/{key}.pem',
                f'{GENDIR}/evidence/{outname}.cbor',
                badnode,
                )
    else:
        raise ValueError(f'Unexpected scheme: {scheme}')


def generate_evidence_from_test_no_nonce(test):
    scheme = test.test_vars['scheme']
    evidence = test.test_vars['evidence']
    signing = test.common_vars['keys'][test.test_vars['signing']]
    outname = f'{scheme}.{evidence}'

    return generate_evidence_no_nonce(scheme, evidence, signing, outname)


def generate_evidence_no_nonce(scheme, evidence, signing, outname):
    os.makedirs(f'{GENDIR}/evidence', exist_ok=True)

    claims_file = f'data/claims/{scheme}.{evidence}.json'

    if scheme == 'psa':
        iak = signing
        generate_psa_evidence_token(
                claims_file,
                f'data/keys/{iak}.jwk',
                f'{GENDIR}/evidence/{outname}.cbor',
                )
    elif scheme == 'cca':
        iak, rak = signing
        generate_cca_evidence_token(
                claims_file,
                f'data/keys/{iak}.jwk',
                f'data/keys/{rak}.jwk',
                f'{GENDIR}/evidence/{outname}.cbor',
                )
    elif scheme == 'enacttrust':
        key = signing
        generate_enacttrust_evidence_token(
                claims_file,
                f'data/keys/{key}.pem',
                f'{GENDIR}/evidence/{outname}.cbor',
                )
    else:
        raise ValueError(f'Unexpected scheme: {scheme}')


def generate_corim(corim_template, comid_templates, output_path):
    output_dir = os.path.dirname(output_path)

    comid_create_cmd = ' '.join(
        [f'cocli comid create --output-dir={output_dir}'] +
        [f'--template={t}' for t in comid_templates]
    )
    run_command(comid_create_cmd, 'generate CoMID(s)')

    comid_files = [os.path.join(output_dir, '.'.join([os.path.splitext(name)[0], 'cbor']))
                   for name in map(os.path.basename, comid_templates)]

    corim_create_cmd = ' '.join(
            [f'cocli corim create --output {output_path} --template={corim_template}'] +
            [f'--comid={cf}' for cf in comid_files]
    )
    run_command(corim_create_cmd, 'generate CoRIM')


def sign_corim(unsigned_corim_path, signed_corim_path):
    meta_file = f'{GENDIR}/meta.json'
    meta_content = {
        "signer": {
            "name": "Veraison Test Signer",
            "uri": "https://veraison.example/test-signer",
            "id": "Veraison Test Signer"
        }
    }
    
    with open(meta_file, 'w') as f:
        json.dump(meta_content, f, indent=2)
    
    key_file = 'data/keys/certs/endEntity.jwk'
    cert_file = 'data/keys/certs/endEntity.der'
    int_cert_file = 'data/keys/certs/intermediateCA.der'
    
    # Sign the CoRIM using cocli
    sign_cmd = (
        f'cocli corim sign '
        f'--file={unsigned_corim_path} '
        f'--key={key_file} '
        f'--cert={cert_file} '
        f'--intermediates={int_cert_file} '
        f'--meta={meta_file} '
        f'--output={signed_corim_path}'
    )
    run_command(sign_cmd, 'sign CoRIM')


def generate_psa_evidence_token(claims_file, key_file, token_file):
    evcli_command = f"evcli psa create --allow-invalid --claims={claims_file} --key={key_file} --token={token_file}"
    run_command(evcli_command, 'generate PSA token')


def generate_cca_evidence_token(claims_file, iak_file, rak_file, token_file):
    evcli_command = f"evcli cca create --claims={claims_file} " + \
                    f"--iak={iak_file} --rak={rak_file} --token={token_file}"
    run_command(evcli_command, 'generate CCA token')

def generate_enacttrust_evidence_token(claims_file, key_file, token_file, badnode):
    bn_flag = '-bad-node' if badnode else ''
    gentoken_command = f"gen-enacttrust-token {bn_flag} -key {key_file} -out {token_file} {claims_file}"
    run_command(gentoken_command, 'generate EnactTrust token')
