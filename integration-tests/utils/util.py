# Copyright 2023-2026 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0
import json
import os
import re
import subprocess
import time

import requests


def to_identifier(text):
     return re.sub(r'\W','_', text)


def update_json(infile, new_data, outfile):
    with open(infile) as fh:
        data = json.load(fh)

    update_dict(data, new_data)

    with open(outfile, 'w') as wfh:
        json.dump(data, wfh)


def update_dict(base, other):
    for k, ov in other.items():
        bv = base.get(k)
        if bv is None:
            continue  # nothing to update

        if isinstance(bv, dict) and isinstance(ov, dict):
            update_dict(bv, ov)
        elif isinstance(bv, dict) or isinstance(ov, dict):
            raise ValueError(f'Value mismatch for "{k}": only one is a dict ({bv}, {ov})')
        else:
            base[k] = ov


def run_command(command: str, action: str) -> int:
    print(command)
    proc = subprocess.run(command, capture_output=True, shell=True)
    print(f'---STDOUT---\n{proc.stdout}')
    print(f'---STDERR---\n{proc.stderr}')
    print(f'------------')
    if proc.returncode:
        executable = command.split()[0]
        raise RuntimeError(f'Could not {action}; {executable} returned {proc.returncode}')


def clear_stores():
    run_command('corim-store db clear', 'clear CoRIM store')
    run_command(
        "sqlite3 /opt/veraison/stores/vts/po-store.sql 'delete from kvstore'",
        'clear policy store',
    )


def get_access_token(test, role):
    kc_host = os.getenv('KEYCLOAK_HOST')
    kc_port = os.getenv('KEYCLOAK_PORT')
    veraison_net = os.getenv('VERAISON_NETWORK')

    # Wait for Keycloak service to come online. This takes a short while, and
    # if the integration tests are run immediately after spinning up the
    # deployment, it may not be there yet.
    for _ in range(10):
        try:
            requests.get(f'https://{kc_host}.{veraison_net}:{kc_port}/')
            break
        except requests.ConnectionError:
            time.sleep(1)
    else:
        raise RuntimeError('Keycloak service does not appear to be online')

    credentials = test.common_vars['credentials'][role]
    data = {
        'client_id': test.common_vars['oauth2']['client-id'],
        'client_secret': test.common_vars['oauth2']['client-secret'],
        'grant_type': 'password',
        'scope': 'openid',
        'username': credentials['username'],
        'password': credentials['password'],

    }
    url = f'https://{kc_host}.{veraison_net}:{kc_port}/realms/veraison/protocol/openid-connect/token'

    r = requests.post(url, data=data)
    resp = r.json()

    return resp['access_token']
