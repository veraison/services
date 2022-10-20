Policies allow tenants to perform additional evaluation of attestation
evidence that is not covered by a particular attestation scheme. The policy can
be used to override the overall attestation status and/or the trust vector
values in the result (e.g. rejecting a token considered valid by the scheme if
the more stringent constraints described in the policy are not met).

> **Note**
> Policy administration framework is to be determined in the future. The
> short-term plan is to make this a part of the deployment flow, but a more
> complete policy admin flow may follow.

The syntax of the policy depends on the agent used to evaluate it. At the
moment, the following policy agents are supported:

"opa" -- [Open Policy Agent](https://www.openpolicyagent.org/) is a flexible,
generic Open Source policy agent that utilizes its own policy language called
Rego. See [README.opa.md](README.opa.md).


## Configuration

The following policy agent configuration directives are currently supported:

- `backend`: specified which policy backend will be used. Currently supported
  backends: `opa`.
- `<backend name>`: an entry with the name of a backend is used to specify
  configuration for that backend. Multiple such entries may exist in a single
  config, but only the one for the backend specified by the `backend` directive
  will be used.

### `opa` backend configuration

Currently, `opa` backend does not support any configuration.
