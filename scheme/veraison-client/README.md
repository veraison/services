# Veraison client plugin

This plugin implements a Veraison API client that requests an appraisal of evidence from a verifier which exposes Veraison's challenge-response and discovery APIs.

VTS can use this plugin in a node running in lead verifier mode when connecting to a downstream verifier, or to itself when the node is running in a hybrid lead/sub mode.

The appraisal request is made in RP mode, which means that the client decides on the nonce to be used to determine the freshness of the supplied evidence.

## Business logic

The plugin is tasked with the following actions:

* Receive component evidence and the nonce from the CE handler.
* Get the verifier's public key and C-R session endpoint by querying the well-known interface.
* Initiate a challenge-response session in RP mode with the configured verifier, supplying the component evidence and nonce.
* Obtain an EAR from the verifier.
* Verify the signature of the EAR.
* Return the EAR appraisal to the CE Handler.

## Configuration

The clientCfg parameter supplied by the VTS contains the relevant connectivity and trust settings as a serialised JSON byte string.

When de-serialised, the JSON object contains the following keys:

* "url" (mandatory): the verifier’s discovery URL
* "ca-certs" (optional): one or more files containing the trust anchors used to authenticate server certificates
* "insecure" (optional): whether certificate verification can skip the trust-related settings

Example:
```json
{
  "url": "https://downstream-verifier.example:8443/.well-known/veraison/verification",
  "ca-certs": [ "/path/to/ca1.pem", "/path/to/ca2.pem" ]
}
```

:warning: :construction: :warning:

Please note that this code is currently in an experimental phase and has not yet been tested as part of the integration test suite.
It may stop working at any time without warning.
If you encounter an issue, please report it as a [bug](https://github.com/veraison/services/issues/new?template=bug-report.md).

:warning: :construction: :warning:
