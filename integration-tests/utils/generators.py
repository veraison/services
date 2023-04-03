import ast
import os
import shutil

from util import update_json, run_command


GENDIR = '__generated__'


def generate_endorsements(test):
    os.makedirs(f'{GENDIR}/endorsements', exist_ok=True)

    scheme = test.test_vars['scheme']
    spec = test.test_vars['endorsements']

    if isinstance(spec, str):
        spec = test.common_vars['endorsements'][spec]

    corim_template_name = 'corim-{}-{}.json'.format(scheme, spec[0])
    corim_template = f'data/endorsements/{corim_template_name}'
    comid_templates = ['data/endorsements/comid-{}-{}.json'.format(scheme, c)
                       for c in spec[1:]]
    output_path = f'{GENDIR}/endorsements/endorsements.cbor'

    generate_corim(corim_template, comid_templates, output_path)


def generate_artefacts_from_response(response, scheme, evidence, signing, keys, expected):
    generate_evidence_from_response(response, scheme, evidence, signing, keys)
    generate_expecte_result_from_response(response, scheme, expected)


def generate_expecte_result_from_response(response, scheme, expected):
    os.makedirs(f'{GENDIR}/expected', exist_ok=True)

    infile = f'data/results/{scheme}.{expected}.json'
    outfile = f'{GENDIR}/expected/{scheme}.{expected}.server-nonce.json'
    nonce = response.json()["nonce"]

    if scheme in ['psa'] and nonce:
        update_json(
                infile,
                {'ear.veraison.annotated-evidence': {f'{scheme}-nonce': nonce}},
                outfile,
                )
    else:
        shutil.copyfile(infile, outfile)


def generate_evidence_from_test(test):
    scheme = test.test_vars['scheme']
    evidence = test.test_vars['evidence']
    nonce = test.common_vars['good-nonce']
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

    if scheme in ['psa', 'cca'] and nonce:
        claims_file = f'{GENDIR}/claims/{scheme}.{evidence}.json'
        update_json(
                f'data/claims/{scheme}.{evidence}.json',
                {f'{scheme}-nonce': nonce},
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
        pak, rak = signing
        generate_cca_evidence_token(
                claims_file,
                f'data/keys/{pak}.jwk',
                f'data/keys/{rak}.jwk',
                f'{GENDIR}/evidence/{outname}.cbor',
                )
    elif scheme == 'enacttrust':
        key = signing
        generate_eancttrust_evidence_token(
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


def generate_psa_evidence_token(claims_file, key_file, token_file):
    evcli_command = f"evcli psa create --allow-invalid --claims={claims_file} --key={key_file} --token={token_file}"
    run_command(evcli_command, 'generate PSA token')


def generate_cca_evidence_token(claims_file, pak_file, rak_file, token_file):
    evcli_command = f"evcli cca create --claims={claims_file} " + \
                    f"--pak={pak_file} --rak={rak_file} --token={token_file}"
    run_command(evcli_command, 'generate CCA token')

def generate_eancttrust_evidence_token(claims_file, key_file, token_file):
    gentoken_command = f"gen-enacttrust-token -key {key_file} -out {token_file} {claims_file}"
    run_command(gentoken_command, 'generate EnactTrust token')
