#!/usr/bin/env python
# Copyright 2023-2025 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0
import argparse
import multiprocessing
import os
import subprocess
import time
import sys

import gevent
import json
import locust.stats # must be imported before requests, as it monkey-patches ssl
import requests
from locust import HttpUser, task, events, tag
from locust.env import Environment
from locust.stats import stats_printer, stats_history, RequestStats, StatsCSVFileWriter
from oauthlib.oauth2 import LegacyApplicationClient
from requests_oauthlib import OAuth2Session

locust.stats.CSV_STATS_INTERVAL_SEC = 1
locust.stats.CSV_STATS_FLUSH_INTERVAL_SEC = 10


aws_auth = {
    'token_url': 'https://keycloak.veraison-testing.com:11111/realms/veraison/protocol/openid-connect/token',
    'client_id': 'veraison-client',
    'client_secret': 'YifmabB4cVSPPtFLAmHfq7wKaEHQn10Z',
    'password': 'veraison',
    'username': 'veraison-provisioner',
}


def get_oauth2_token(token_url, username, password, client_id, client_secret):
    oauth = OAuth2Session(client=LegacyApplicationClient(client_id=client_id))
    return oauth.fetch_token(
        token_url=token_url,
        username=username, password=password,
        client_id=client_id, client_secret=client_secret,
    )['access_token']


def get_auth_headers(auth_config):
    ret = {}
    if auth_config:
        token = get_oauth2_token(**auth_config)
        ret['Authorization'] = f'Bearer {token}'
    return ret


class BaseVeraisonUser(HttpUser):
    provisioning_host = ''
    verification_host = ''
    host = ''
    auth_headers = {}

    def on_start(self):
        this_dir = os.path.dirname(__file__)

        evidence_file = os.path.join(this_dir, '../data/psa.good.cbor')
        with open(evidence_file, 'rb') as fh:
            self.evidence_data = fh.read()

        # Make sure necessary endorsements are provisioned even of the
        # runner hasn't run ProvisioningUser yet.
        endorsements_file = os.path.join(this_dir, '../data/corim-psa-full.cbor')
        with open(endorsements_file, 'rb') as fh:
            self.provisioning_data = fh.read()

        headers = {
            'Content-Type': 'application/corim-unsigned+cbor; profile="http://arm.com/psa/iot/1"',
        }
        headers.update(self.auth_headers)

        requests.post(
            self.provisioning_host + "/endorsement-provisioning/v1/submit", # type: ignore
            headers=headers,
            data=self.provisioning_data,
        )

    @tag('well-known')
    @task
    def get_wellkown(self):
        self.client.get(
            self.provisioning_host + '/.well-known/veraison/provisioning',
            headers=self.auth_headers,
        )

    @tag('provision')
    @task
    def provision(self):
        headers = {
            'Content-Type': 'application/corim-unsigned+cbor; profile="http://arm.com/psa/iot/1"',
        }
        headers.update(self.auth_headers)

        self.client.post(
            self.provisioning_host + '/endorsement-provisioning/v1/submit',
            headers=headers,
            data=self.provisioning_data,
        )

    @tag('verify')
    @task
    def verification(self):
        resp = self.client.post(
            self.verification_host + '/challenge-response/v1/newSession?nonce=QUp8F0FBs9DpodKK8xUg8NQimf6sQAfe2J1ormzZLxk=',
            data=self.evidence_data,
            name='/challenge-response/v1/newSession?nonce=[NONCE]',
        )

        location = resp.headers.get('Location')

        self.client.post(
            self.verification_host + f'/challenge-response/v1/{location}',
            headers={
                'Content-Type': 'application/psa-attestation-token',
            },
            data=self.evidence_data,
            name='/challenge-response/v1/[SESSION_ID]',
        )

        self.client.delete(
            self.verification_host + f"/challenge-response/v1/{location}",
            name=f"/challenge-response/v1/[SESSION_ID]",
        )


class LocalVeraisonUser(BaseVeraisonUser):
    verification_host = 'http://localhost:8443'
    provisioning_host = 'http://localhost:9443'


class AWSVeraisonUser(BaseVeraisonUser):
    verification_host = 'https://services.veraison-testing.com:8443'
    provisioning_host = 'https://services.veraison-testing.com:9443'

    def on_start(self):
        self.auth_headers = get_auth_headers(aws_auth)
        super().on_start()


user_classes= {
    'local': [
        LocalVeraisonUser,
    ],
    'aws': [
        AWSVeraisonUser,
    ],
}


all_tasks = ['well-known', 'provision', 'verify']


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('-u', '--users', type=int, default=2,
                        help='number of user that will be spawned')
    parser.add_argument('-d', '--duration', type=int, default=30,
                        help='The duration for which the test will run.')
    parser.add_argument('-w', '--workers', type=int, default=multiprocessing.cpu_count(),
                        help='number of worker processes that will be used')
    parser.add_argument('-l', '--local', action='store_true',
                        help='run in a single process (-w will be ignored)')
    parser.add_argument('--no-webui', action='store_false', dest='webui',
                        help='do not launch web UI')
    parser.add_argument('-o', '--output-dir', default='run-loads-results',
                        help='output file the results will be written to')
    parser.add_argument('-f', '--force', action='store_true',
                        help='overwrite existing output')
    parser.add_argument('-s', '--services', default='local', choices=['local', 'aws'],
                        help='Which services deployment to use')
    parser.add_argument('-t', '--task', action='append', dest='tasks',
                        choices=all_tasks,
                        help='The tasks the user will perform')
    args = parser.parse_args()
    if not args.tasks:
        args.tasks = all_tasks

    if os.path.exists(args.output_dir) and not args.force:
        print(f'ERROR: {args.output_dir} already exists')
        sys.exit(1)
    os.makedirs(args.output_dir, exist_ok=True)

    # WebUI, when created, will instantiate its own argument parser (for some
    # reason), so we need to clear args so that it doesn't complain about the
    # flags not meant for it.
    sys.argv = [sys.argv[0]]

    env = Environment(
        user_classes=user_classes[args.services],
        tags=args.tasks,
        events=events,
    )

    csv_prefix = os.path.join(args.output_dir, 'csv/')
    os.makedirs(csv_prefix, exist_ok=True)
    csv_writer = StatsCSVFileWriter(
        env, locust.stats.PERCENTILES_TO_REPORT, csv_prefix, full_history=True,
    )

    if args.local:
        run_local(env, args.users, args.duration, args.webui, csv_writer)
    else:
        run(env, args.users, args.workers, args.duration, args.webui, csv_writer)

    outfile = os.path.join(args.output_dir, 'result.json')
    postamble(env.stats, outfile)


def run(
    env: Environment,

    num_users: int,
    num_workers: int,
    duration: int,
    launch_webui: bool = True,
    csv_writer: StatsCSVFileWriter | None = None,
):
    runner = env.create_master_runner()

    if launch_webui:
        web_ui = env.create_web_ui("127.0.0.1", 8089, stats_csv_writer=csv_writer)

    gevent.spawn(stats_printer(env.stats))
    gevent.spawn(stats_history, env.runner)
    if csv_writer:
        gevent.spawn(csv_writer.stats_writer)

    tags = ['--tag']
    tags.extend(env.tags or [])

    workers = []
    for _ in range(num_workers):
        workers.append(subprocess.Popen(["locust", "--worker", "-f", __file__] + tags))

    time.sleep(2)
    runner.start(num_users, spawn_rate=num_users)

    gevent.spawn_later(duration, lambda: runner.quit())

    runner.greenlet.join()
    for worker in workers:
        worker.wait(timeout=5)

    if launch_webui:
        web_ui.stop() # type: ignore


def run_local(
    env: Environment,
    duration: int,
    num_users: int,
    launch_webui: bool = True,
    csv_writer: StatsCSVFileWriter | None = None,
):
    runner = env.create_local_runner()

    if launch_webui:
        web_ui = env.create_web_ui("127.0.0.1", 8089, stats_csv_writer=csv_writer)

    gevent.spawn(stats_printer(env.stats))
    gevent.spawn(stats_history, env.runner)
    if csv_writer:
        gevent.spawn(csv_writer.stats_writer)

    time.sleep(2)
    runner.start(num_users, spawn_rate=num_users)

    gevent.spawn_later(duration, lambda: runner.quit())

    runner.greenlet.join()

    if launch_webui:
        web_ui.stop()


def postamble(stats: RequestStats, outfile: str):
    results = []
    for entry in stats.entries.values():
        results.append(entry.serialize())
        print(entry.to_string())

    with open(outfile, 'w') as wfh:
        json.dump(results, wfh)


if __name__ == '__main__':
    main()
