# Docker Deployment

The structure of the docker deployment is as follows:
- There are 3 containers; one for each service (provisioning, verification and vts)
- There are two networks:
    - `provisioning-network`: This network allows communication between VTS and
      the provisioning service
    - `verification-network`: This network allows communication between VTS and
      the verification service

## Dockerfile

A [Dockerfile](./Dockerfile) is used to perform a multi-stage build that
outputs the final image for each container:
1. `build-base`: In this stage, the working directory is setup and project
   dependencies and libraries are installed
2. `common-build`:
    - In this stage, build arguments that are passed in from the
      `docker-compose.yml` file, are used to establish the appropriate
      environment variables and the configs for each service are generated.
    - Files for each service are installed into the directory denoted by the
      environment variable `<service>_DEPLOY_PREFIX` using `make install`
    - Individual service directories are bundled into an individual tar file.
3. `provisioning-run`: In this stage the `provisioning.tar` file is copied from
   the common-build image and extracted
4. `verification-run`: In this stage the `verification.tar` file is copied from
   the common-build image and extracted
5. `vts-run`: In this stage the `vts.tar` file is copied from the common-build
   image and extracted

## Docker-compose

The [docker-compose](./docker-compose.yml) functionality defines the
configuration for each container and networks between each container.

For dynamic configuration we pass in a `.env` file with the `docker compose up`
command:


-  Builds the images for the 3 containers
```bash
docker compose --env-file default.env build
```

-  Runs all 3 containers using the images built in the previous step
```bash
docker compose --env-file default.env up
```

- Tears down all 3 containers (this only removes the containers, not the images)
```bash
docker compose --env-file default.env down
```
NOTE: Ensure the above commands are run in the same directory as the
      `docker-compose.yml` and `Dockerfile`

NOTE: The command in 3. is only needed if you need to rebuild the images
      (possibly because new input file, plugins, or code has been added)

## Environment variables

A `default.env` file which provides docker service configurations in the
following way:
- The `default.env` file is read into docker compose as the `--env-file`
  argument is specified in the 'build' and 'up' command
- These environment variables are passed in as build arguments (variable
  substitution happens here) in the `docker-compose.yml` file
- The `Dockerfile` takes the build arguments and initialises them as
  environment variables, ready for the build and installation process

The `default.env` aims to provide a single source of configuration for
individual services and docker configuration. The documentation on the current
environment variables is provided in the [.env](default.env) file.

## Configuration templating

The [Jinja2](https://jinja.palletsprojects.com/en/3.1.x/templates/) templating
engine is used to generate the configuration file for each individual service
(provisioning, verification and vts). The templates are parsed and populated
using the python script [generate-config.py](./generate-config.py). The
templates take in environment variables to configure the service's settings. If
the environment variable is unset or empty, the template is populated with its
default value.

