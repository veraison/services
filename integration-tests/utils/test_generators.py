# Copyright 2026 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

from generators import base64url_to_base64


def test_base64url_to_base64_replaces_characters_and_restores_padding():
    assert base64url_to_base64('-_8') == '+/8='


def test_base64url_to_base64_does_not_add_unnecessary_padding():
    assert base64url_to_base64('YWJj') == 'YWJj'
