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

## Policy Identification

There are three different ways of identifying a policy:

### appraisal policy ID

This is an identifier used in the attestation result as defined by [EAR
Internet draft](https://www.ietf.org/archive/id/draft-fv-rats-ear-01.html#name-ear-appraisal-claims).
Note that in this case, "policy" is used in the sense of "Appraisal Policy for
Evidence" as per [RATS architecture](https://www.rfc-editor.org/rfc/rfc9334).
In Veraison, this encompasses both, the scheme, and the policy applied from the
policy store by a policy engine.

An appraisal policy ID is a [URI](https://www.rfc-editor.org/rfc/rfc3986) with
the scheme `policy` followed by a rootless path indicating the (RATS) policy
using which the appraisal has been generated. The first segment of the path is
the name of the scheme used to create the appraisal. The second segment, if
present, is the individual policy ID (see below) of the policy that has been
applied to the appraisal created by the scheme.

For example:

- `policy:TPM_ENACTTRUST`: the appraisal has been created using "TPM_ENACTTRUST" scheme, with
  no additional policy applied.
- `policy:PSA_IOT/340d22f7-9eda-499f-9aa2-5af295d6d812`: the appraisal has been
  created using "PSA_IOT" scheme and has subsequently been updated by the
  policy with unique policy ID "ae19cc27-a449-1fb8-6c10-00f47ad1c55c".

#### Potential future extensions

These indicate potential future enhancements, and are **not** supported by the
current implementation.

##### Cascading policies

In the future we may support applying multiple individual policies to a single
appraisal. In that case, each path segment after the first (the scheme) is the
individual policy ID of a policy that has been applied. The ordering of the
segments matches the order in which the policies were applied.

For example:

- `policy:PSA_IOT/340d22f7-9eda-499f-9aa2-5af295d6d812/ae19cc27-a449-1fb8-6c10-00f47ad1c55c`:
  the appraisal has been created using "PSA_IOT" scheme, it was then updated by
  a policy with the individual policy id
  `340d22f7-9eda-499f-9aa2-5af295d6d812`, followed by a policy with the
  individual policy ID `ae19cc27-a449-1fb8-6c10-00f47ad1c55c`.

### policy store key

Policies are stored, retrieved from, and updated in the policy store using a key.
The key is a string consisting of the tenant id, scheme, and policy name
delimited by colons.

For example:

- `0:PSA_IOT:opa`: the key for tenant "0"'s policy for scheme "PSA_IOT" with
  name "opa".

#### policy name

The name exists to support cascading policies in the future (see above). At the
moment, as there is only one policy allowed per appraisal, the name is not
necessary and is always set to the name of the policy engine ("opa"). While
this unnecessarily increases the key size and is somewhat wasteful, given that
the number of the policies a typical deployment is expected to be, at most, in
the hundreds, and the relatively negligible overhead compared to the size of the
polices themselves, this is not deemed to be a major concern.

### individual policy ID

The individual policy ID identifies the specific policy that was applied to an
appraisal. It forms a component of the appraisal policy ID (which also includes
the scheme, and possibly, in the future, individual IDs from multiple
policies). It differs from the policy store key in that it also incorporates
versioning information.

The individual policy id is the UUID of the specific policy instance.

For example:

- `340d22f7-9eda-499f-9aa2-5af295d6d812`: policy for tenant with id "0", named
  "opa", at version 1.
