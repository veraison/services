# Services

This repository contains attestation services assembled using Veraison components.

## Provisioning

Provisioning service provides a REST-based API for external trusted supply chain actors (for example, Endorsers) to provision Reference Values, Endorsed Values (known as Endorsements), and Trust Anchors into Veraison Trusted Services.

It uses Attestation Format (e.g. PSA) specific decoders  to extract the Endorsements from received payload. On the back end it communicates with [VTS](#Veraison-Trusted-Services) to store, retrieve, and manage the Endorsements and Trust Anchors. The API details are documented under [Endorsement Provisioning Interface](https://github.com/veraison/docs/tree/main/api/endorsement-provisioning). Provisioning service acts as a front end to accept a variety of Endorsement Formats. For now, only PSA (Profile 1 & Profile 2) and TPM based Endorsements are supported.

Refer to [scope](https://github.com/veraison/docs/blob/main/project-overview.md#scope---provisioning) for full set of services that shall be provided by the provisioning service.

## Verification

Verification service provides a REST-based API for external Attesters or Relying Parties (known as Challengers) to submit Attestation token, containing Attestation Evidence claims. On the back end, it communicates with [VTS](#Veraison-Trusted-Services) to appraise the received Evidence and receive the Attestation Verification Results, which are then passed to the challenger.

This service acts as a frontend for accepting a variety of attestation token formats. For now, only PSA (Profile 1 & Profile 2) and TPM-based attestation tokens are supported.

The API is based on the Challenge/Response Interaction Models as documented in [challenge-response](https://github.com/veraison/docs/tree/main/api/challenge-response)


## Veraison Trusted Services

Veraison Trusted Services (VTS) backend provides core services to the Verification and Provisioning frontends. On the Provisioning path, it receives Endorsements and Trust Anchors, from the [Provisioning](#Provisioning) frontend, invokes Attestation Format (e.g. PSA) specific logic (a "scheme") to generate Endorsements and Trust Anchors, and saves them in the corresponding stores. On the Verification path, it invokes Attestation Format specific logic (a "scheme") to parse attestation token and extract evidence claims. From these claims, it constructs a logical index/key to fetch the trust anchors required to verify the token signature, and the associated Endorsements (`golden values`) used to appraise the evidence claims. Upon evaluation, it populates the Attestation Result and communicates it to the [Verification](#Verification) frontend.


More details about the VTS can be found under [VTS](https://github.com/veraison/docs/tree/main/architecture/verifier#vts)

## KVStore

The key-values store is the Veraison Storage Layer. It is used to store both Endorsements and Trust Anchors.

KV Store details can be found under [kvstore](https://github.com/veraison/services/tree/migration/kvstore#kv-store)
