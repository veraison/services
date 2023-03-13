# Copyright 2023 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0
FROM python as veraison-test

# User identify that will be used to build the project. This should be
# overriden at build time to match the host user running the tester who owns
# the git checkout of the source.
ARG TESTER_UID=1000
ARG TESTER_GID=1000

RUN userdel -f $(cat /etc/passwd | awk -F: "\$3 == ${TESTER_UID}" | cut -d: -f1); \
    groupdel -f $(cat /etc/group | awk -F: "\$3 == ${TESTER_GID}" | cut -d: -f1); \
    groupadd -g ${TESTER_GID} tavern && \
    useradd -m -u ${TESTER_UID} -g tavern \
    -s /bin/bash tavern

WORKDIR /integration-tests

USER tavern

RUN pip install --upgrade pip &&  \
    pip install jsonschema==3.2.0 \
                tavern==2.0.5 \
                python-jose==3.3.0 \
                ipdb==0.13.11

ADD gen-enacttrust-token evcli cocli run-tests /home/tavern/.local/bin/
ADD bashrc /home/tavern/.bashrc

ENTRYPOINT ["/home/tavern/.local/bin/run-tests"]

