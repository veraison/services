The code in this directory represents coserv broker which can be used to fetch endorsements or reference values 
from a remote coserv service.

Configuration to setup the broker is provided in:
`services/deployments/docker/src/config.yaml.template`

The setup is as follows:
```yml
endorsement-store:
  plugin-parameters:
    coserv-broker:
      enable-broker: true                                   # Enable endorsement broker to call remote coserv service
      client-caching: true                                  # Enable caching for coserv fetch
      schemes:
        ARM_CCA:                                                # Scheme based configuration
          coserv-url: "https://veraison.test.linaro.org:11443"  # Remote coserv service URL which follows draft-ietf-rats-coserv-06 IETF draft
          client-tls-insecure: true                             # Enable insecure TLS, no certificate validation
          client-certs:                                         # If secure TLS, certs to verify, with each path inside docker container
            - "cert1"
          unsigned-coserv: false                                # If true then only accepts unsigned coserv as reponse from remote service
          profile: "tag:arm.com,2025:cca_platform#1.0.0"        # CoSERV profile
```