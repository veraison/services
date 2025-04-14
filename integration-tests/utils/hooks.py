# Copyright 2023-2024 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0
import os

from generators import *
from util import run_command, get_access_token


def setup_end_to_end(test, variables):
    _set_content_types(test, variables)
    _set_authorization(test, variables, 'provisioner')
    _set_nonce(test, variables)
    generate_endorsements(test)
    generate_evidence_from_test(test)


def setup_bad_session(test, variables):
    _set_authorization(test, variables, 'provisioner')
    generate_endorsements(test)


def setup_no_nonce(test, variables):
    _set_authorization(test, variables, 'provisioner')
    generate_evidence_from_test(test)


def setup_multi_nonce(test, variables):
    _set_content_types(test, variables)
    _set_authorization(test, variables, 'provisioner')
    _set_nonce(test, variables)
    generate_endorsements(test)
    generate_evidence_from_test_no_nonce(test)


def setup_enacttrust_badnode(test, variables):
    _set_authorization(test, variables, 'provisioner')
    _set_content_types(test, variables)
    _set_nonce(test, variables)
    generate_endorsements(test)
    generate_evidence_from_test(test)


def setup_enacttrust_badkey(test, variables):
    _set_authorization(test, variables, 'provisioner')
    _set_content_types(test, variables)


def setup_policy_management(test, variables):
    _set_authorization(test, variables, 'manager')


def setup_provisioning_fail_empty_body(test, variables):
    _set_authorization(test, variables, 'provisioner')


def setup_cca_verify_challenge(test, variables):
    _set_content_types(test, variables)
    _set_authorization(test, variables, 'provisioner')
    _set_alt_authorization(test, variables, 'manager')
    _set_nonce(test, variables)
    generate_endorsements(test)
    generate_evidence_from_test(test)

def setup_cca_end_to_end(test, variables):
    _set_cca_content_types(test, variables)
    _set_authorization(test, variables, 'provisioner')
    _set_alt_authorization(test, variables, 'manager')
    _set_nonce(test, variables)
    generate_cca_end_to_end_endorsements(test)
    generate_evidence_from_test(test)

def setup_freshness_check_fail(test, variables):
    _set_content_types(test, variables)
    _set_authorization(test, variables, 'provisioner')
    _set_nonce(test, variables)
    generate_endorsements(test)
    generate_evidence_from_test(test)

def _set_content_types(test, variables):
    scheme = test.test_vars['scheme']
    profile = test.test_vars['profile']
    ends_content_types = test.common_vars['endorsements-content-types']
    ev_content_types = test.common_vars['evidence-content-types']
    variables['endorsements-content-type'] = ends_content_types[f'{scheme}.{profile}']
    variables['evidence-content-type'] = ev_content_types[f'{scheme}.{profile}']


def _set_authorization(test, variables, role):
    token = get_access_token(test, role)
    variables['authorization'] = f'Bearer {token}'


def _set_alt_authorization(test, variables, role):
    token = get_access_token(test, role)
    variables['alt-authorization'] = f'Bearer {token}'


def _set_nonce(test, variables):
    nonce_config = test.test_vars['nonce']
    variables['nonce-value'] = test.common_vars[nonce_config]['value']
    variables['nonce-bad-value'] = test.common_vars[nonce_config]['bad-value']
    variables['nonce-size'] = test.common_vars[nonce_config]['size']


def _set_cca_content_types(test, variables):
    scheme = test.test_vars['scheme']
    profile = test.test_vars['profile']
    corim_type = test.test_vars.get('corim_type', 'unsigned')
    ev_content_types = test.common_vars['evidence-content-types']
    
    variables['evidence-content-type'] = ev_content_types[f'{scheme}.{profile}']
    
    # Set platform and realm content types
    if corim_type == 'signed':
        # Use signed content types
        variables['platform-en-content-type'] = 'application/rim+cose; profile="http://arm.com/cca/ssd/1"'
        variables['realm-en-content-type'] = 'application/rim+cose; profile="http://arm.com/cca/realm/1"'
    else:
        # Use unsigned content types
        variables['platform-en-content-type'] = 'application/corim-unsigned+cbor; profile="http://arm.com/cca/ssd/1"'
        variables['realm-en-content-type'] = 'application/corim-unsigned+cbor; profile="http://arm.com/cca/realm/1"'
