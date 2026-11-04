The code in this directory represents coserv broker which can be used to fetch endorsements or reference values 
from a remote coserv service.

Configuration to setup the broker is provided in:
Docker Based Deployment: `services/deployments/docker/src/config.yaml.template`
Native Deployment: `services/deployments/native/config/services.yaml.template`

The setup is as follows:
```yml
endorsement-store:
  active-plugins:
    # Mention coserv-broker as an active store plugin for vts to search
    - coserv-broker
  plugin-parameters:
    coserv-broker:
      # Enable endorsement broker
      enable-broker: true
      # Enable caching for coserv fetch
      client-caching: true
      # Scheme based configuration
      schemes:
        ARM_CCA:
          # Remote coserv service URL
          coserv-url: "https://veraison.test.linaro.org:11443"
          # Enable insecure TLS
          client-tls-insecure: true
          # Certificates with each path inside docker container
          client-certs:
            - "cert1"
          # If true then only accepts unsigned coserv
          unsigned-coserv: false
          # CoSERV profile
          profile: "tag:arm.com,2025:cca_platform#1.0.0"
```

>Note: active-plugins and enable-broker need to be kept in sync, if coserv-broker is removed from active-plugins
>in order to deactivate broker process, enable-broker should be false. Similarly if coserv-broker is added as an
>active plugin then enable-broker should be true.