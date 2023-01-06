#!/usr/bin/env python
# Copyright 2023 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

import os
from jinja2 import Environment, FileSystemLoader

def env_override(default_value, key):
    """ Gets value of the environment variable provided in the key """
    return os.getenv(key, default_value)

def generate_config(path_to_config_templates, path_to_new_config, curr_service):
    # Load templates into Jinja environment
    env = Environment(loader = FileSystemLoader(path_to_config_templates), trim_blocks=True, lstrip_blocks=True)
    env.filters['env_override'] = env_override

    # Renders config templates for each service into a config.yaml file
    template = env.get_template(curr_service + '_tmpl.j2')
    print(template.render())

    file = open(path_to_new_config, 'w')
    file.write(template.render())
    file.close()

# TODO Fix relative directory
services = ["provisioning", "verification", "vts"]
templates = "./config-templates"
for service in services:
    file = "../../" + service + "/cmd/" + service + "-service/config.yaml"
    generate_config(templates, file, service)
