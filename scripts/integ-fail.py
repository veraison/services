# Copyright 2022-2023 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0
import os
import re
import sys

# read output of docker container running integration tests 
failure_lines = sys.stdin.read()

if len(failure_lines) == 0:
    sys.exit(0)
else:
    sys.exit(1)