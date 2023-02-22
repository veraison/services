# Integration tests with Tavern

Useful aspects to note before getting started:

- Tests are defined in YAML files, and each new test must have a file name with
  the following format: `test_<test_name>.tavern.yaml`
- Tavern tests are run in a separate container (called `tavern`)
- Tests will only run successfully once the full docker environment of all
  services and test containers are setup

## Structure of Tavern test cases

- Each Tavern test has a test name and can consist of several stages
- Each stage has request information such as a HTTP method, URL, headers etc,
  that you submit a request with
- Each stage has a response field that is populated with the response you
  expect for that specific request

    - For example to validate that the key-value pair `val: none` is present in
      the response the `json` field needs to be checked:

        ```yaml
        response:
            status_code: 200
            json:
                nonce: "{nonce:s}"
        ```

- Values from a response can be saved for use in subsequent stages. For example
  the below saves the Location field in the `header`, into the variable
  `my-var` which can be used in subsequent stages within the same test:

    ```yaml
    response:
        ....
        save:
            headers:
                my-var: Location
    ```

- External functions can be written and used to customise validation:

    ```yaml
    response:
        ....
        verify_response_with:
            - function: <file_name_with_function>:<func_name>
    ```

## Setup integration test environment to run tests locally
Below are the steps to setup the environment and run the integration tests:

1. Open `integration-tests/docker-compose-integration-tests.yml` file and paste
   the `command` entry into the tavern service configuration, inline with build
   and volume configurations:

    ```yaml
    build:
        ....
    volumes:
        ....
    command: sleep infinity
    ```

    Explanation of volumes: this configuration mounts your local
    `integration-tests` directory into the container at the location
    `/integration-tests` in the container. Each time you edit a test file
    locally, the changes will update in the container. This means you wont have
    to rebuild the docker containers each time you want edit your test.

    Explanation of command: The purpose of the command `sleep infinity` is to
    allow the container to stay alive so you can shell into the container and
    run the tests manually.

    > **Warning**: Remember to remove the `command: sleep infinity` line before
    > commiting your changes!

2. Change to the integration-tests directory locally and setup the docker
   containers by running the below make command.

    ```bash
    cd services/integration-tests
    make integration-tests-up
    ```

3. Use the following command to attach to the logs of the running services to
   see the live logs.

    ```
    docker compose \
        --env-file=../deployments/docker/default.env \
        --file=docker-compose-integration-tests.yml \
        logs -f -t
    ```

4. In another terminal, use this command to shell into the tavern test
   container (this will give you access to an interactive bash shell inside the
   tavern docker container).

    ```bash
    docker exec -it tavern bash
    ```

5. Run the tests using the following command in the shell of the container.
   Make sure you are in the root directory of the container when running the
   command.

    ```bash
    PYTHONPATH=$PYTHONPATH:integration-tests \
        py.test integration-tests/ -vv
    ```

