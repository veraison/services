# Copyright 2023 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0
import os

from generators import *
from util import run_command, get_access_token


def setup_end_to_end(test, variables):
    _set_content_types(test, variables)
    _set_authorization(test, variables, 'provisioner')
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
    generate_endorsements(test)
    generate_evidence_from_test_no_nonce(test)


def setup_enacttrust_badnode(test, variables):
    _set_authorization(test, variables, 'provisioner')
    _set_content_types(test, variables)
    generate_endorsements(test)
    generate_evidence_from_test(test)

def setup_enacttrust_badkey(test, variables):
    _set_authorization(test, variables, 'provisioner')
    _set_content_types(test, variables)
    generate_endorsements(test)


def setup_policy_management(test, variables):
    _set_authorization(test, variables, 'manager')


def setup_provisioning_fail_empty_body(test, variables):
    _set_authorization(test, variables, 'provisioner')


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


