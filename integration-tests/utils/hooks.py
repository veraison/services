import os

from generators import *


def setup_end_to_end(test, variables):
    _set_content_types(test, variables)
    generate_endorsements(test)
    generate_evidence_from_test(test)


def setup_bad_session(test, veriables):
    generate_endorsements(test)


def setup_no_nonce(test, veriables):
    generate_evidence_from_test(test)

def _set_content_types(test, variables):
    scheme = test.test_vars['scheme']
    profile = test.test_vars['profile']
    ends_content_types = test.common_vars['endorsements-content-types']
    ev_content_types = test.common_vars['evidence-content-types']
    variables['endorsements-content-type'] = ends_content_types[f'{scheme}.{profile}']
    variables['evidence-content-type'] = ev_content_types[f'{scheme}.{profile}']



