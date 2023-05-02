#!/usr/bin/env python
import argparse
import multiprocessing
import os
import subprocess
import time

import gevent
import json
import locust.stats # must be imported before requests, as it monkey-patches ssl
import requests
from locust import HttpUser, task, events
from locust.env import Environment
from locust.stats import stats_printer, stats_history, StatsEntry, RequestStats

locust.stats.CSV_STATS_INTERVAL_SEC = 5 # default is 1 second
locust.stats.CSV_STATS_FLUSH_INTERVAL_SEC = 60 # Determines how often the data is flushed to disk, default is 10 seconds


class ProvisioningUser(HttpUser):
    host = 'http://localhost:8888'

    def on_start(self):
        this_dir = os.path.dirname(__file__)

        endorsements_file = os.path.join(this_dir, '../data/corim-psa-full.cbor')
        with open(endorsements_file, 'rb') as fh:
            self.provisioning_data = fh.read()

    @task
    def get_wellkown(self):
        self.client.get("/.well-known/veraison/provisioning")

    @task
    def provision(self):
        self.client.post(
            "/endorsement-provisioning/v1/submit",
            headers={
                'Content-Type': 'application/corim-unsigned+cbor; profile=http://arm.com/psa/iot/1',
            },
            data=self.provisioning_data,
        )


class VerificationUser(HttpUser):
    host = 'http://localhost:8080'

    def on_start(self):
        this_dir = os.path.dirname(__file__)

        evidence_file = os.path.join(this_dir, '../data/psa.good.cbor')
        with open(evidence_file, 'rb') as fh:
            self.evidence_data = fh.read()

        # Make sure necessary endorsements are provisioned even of the
        # runner hasn't run ProvisioningUser yet.
        endorsements_file = os.path.join(this_dir, '../data/corim-psa-full.cbor')
        with open(endorsements_file, 'rb') as fh:
            provisioning_data = fh.read()

        requests.post(
            ProvisioningUser.host + "/endorsement-provisioning/v1/submit",
            headers={
                'Content-Type': 'application/corim-unsigned+cbor; profile=http://arm.com/psa/iot/1',
            },
            data=provisioning_data,
        )

    @task
    def verification(self):
        resp = self.client.post(
            "/challenge-response/v1/newSession?nonce=QUp8F0FBs9DpodKK8xUg8NQimf6sQAfe2J1ormzZLxk=",
            data=self.evidence_data,
        )

        location = resp.headers.get('Location')

        self.client.post(
            f"/challenge-response/v1/{location}",
            headers={
                'Content-Type': 'application/psa-attestation-token',
            },
            data=self.evidence_data,
        )

        self.client.delete(f"/challenge-response/v1/{location}")


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
    parser.add_argument('-o', '--outfile', default='results.json',
                        help='output file the results will be written to')
    args = parser.parse_args()

    # WebUI, when created, will instantiate its own argument parser (for some
    # reason), so we need to clear args so that it doesn't complain about the
    # flags not meant for it.
    import sys
    sys.argv = [sys.argv[0]]

    env = Environment(user_classes=[
        ProvisioningUser,
        VerificationUser,
    ], events=events)

    if args.local:
        run_local(env, args.users, args.duration, args.webui)
    else:
        run(env, args.users, args.workers, args.duration, args.webui)

    post_process(env.stats, args.outfile)


def run(
    env: Environment,
    num_users: int,
    num_workers: int,
    duration: int,
    launch_webui: bool = True,
):
    runner = env.create_master_runner()

    if launch_webui:
        web_ui = env.create_web_ui("127.0.0.1", 8089)

    gevent.spawn(stats_printer(env.stats))
    gevent.spawn(stats_history, env.runner)

    workers = []
    for i in range(num_workers):
        workers.append(subprocess.Popen(["locust", "--worker", "-f", __file__]))

    time.sleep(2)
    runner.start(num_users, spawn_rate=num_users)

    gevent.spawn_later(duration, lambda: runner.quit())

    runner.greenlet.join()
    for worker in workers:
        worker.wait(timeout=5)

    if launch_webui:
        web_ui.stop()


def run_local(env: Environment, num_users: int, launch_webui: bool = True):
    runner = env.create_local_runner()

    if launch_webui:
        web_ui = env.create_web_ui("127.0.0.1", 8089)

    gevent.spawn(stats_printer(env.stats))
    gevent.spawn(stats_history, env.runner)

    time.sleep(2)
    runner.start(num_users, spawn_rate=num_users)

    gevent.spawn_later(30, lambda: runner.quit())

    runner.greenlet.join()

    if launch_webui:
        web_ui.stop()


def post_process(stats: RequestStats, outfile: str):
    # Stats are kept based on the URL path and on the request method. Since the
    # session is part of the challenge-response path, there will be a different
    # stats object for each request; we should combine those
    combined_entries = {}
    for se in stats.entries.values():
        parts = se.name.split('/')
        if parts[1] == "challenge-response" and parts[-1].count('-') == 4:
            name = '/'.join(parts[:-1] + ['SESSION_ID'])
        else:
            name = se.name

        if len(name) > 70:
            name = name[:65] + '[...]'

        key = (name, se.method)
        if key not in combined_entries:
            combined_entries[key] = StatsEntry(
                    stats=se.stats,
                    name=name,
                    method=se.method,
            )
        combined_entries[key].extend(se)

    stats.entries = combined_entries

    results = []
    for entry in combined_entries.values():
        results.append(entry.serialize())
        print(entry.to_string())

    with open(outfile, 'w') as wfh:
        json.dump(results, wfh)


if __name__ == '__main__':
    main()
