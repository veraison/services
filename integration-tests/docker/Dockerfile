# Copyright 2023-2025 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0
FROM python:3.11 AS veraison-test

# User identify that will be used to build the project. This should be
# overriden at build time to match the host user running the tester who owns
# the git checkout of the source.
ARG TESTER_UID=1000
ARG TESTER_GID=1000

RUN apt-get update \
    && DEBIAN_FRONTEND=noninteractive apt-get install \
        --assume-yes \
        --no-install-recommends \
        less \
        ruby \
        jq \
        vim \
        sqlite3 \
        jose \
    && apt-get clean \
    && apt-get autoremove --assume-yes \
    && rm -rf /var/lib/apt/lists/* /var/tmp/* /tmp/* && \
    gem install cbor-diag

RUN userdel -f $(cat /etc/passwd | awk -F: "\$3 == ${TESTER_UID}" | cut -d: -f1); \
    groupdel -f $(cat /etc/group | awk -F: "\$3 == ${TESTER_GID}" | cut -d: -f1); \
    groupadd -g ${TESTER_GID} tavern && \
    groupadd -g 616 veraison && \
    useradd -m -u ${TESTER_UID} -g tavern -G veraison \
    -s /bin/bash tavern

ADD --chown=root:root rootCA.crt /usr/local/share/ca-certificates/veraison-ca.crt
RUN update-ca-certificates
ENV REQUESTS_CA_BUNDLE=/etc/ssl/certs/ca-certificates.crt

WORKDIR /integration-tests

USER tavern

RUN pip install --upgrade pip &&  \
    pip install jsonschema==4.22.0 \
                tavern==2.11.0 \
                urllib3==2.2.1 \
                python-jose==3.3.0 \
                ipdb==0.13.11

ADD gen-enacttrust-token evcli cocli run-tests /home/tavern/.local/bin/
ADD bashrc /home/tavern/.bashrc

ENTRYPOINT ["/home/tavern/.local/bin/run-tests"]

