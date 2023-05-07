## Configuration

- `ear-signer`: stanza containing the configuration details about the attestation
  result signing process.  The supported directives are:
  - `alg`: the [JWS algorithm](https://www.iana.org/assignments/jose/jose.xhtml#web-signature-encryption-algorithms)
    used for signing, e.g.: `ES256`, `RS512`.
  - `key` (optional): file containing the private key to be used with `alg`.
    The key is in [JWK format](https://datatracker.ietf.org/doc/rfc7517/).
    A new key-pair is generated on-the-fly if the `key` directive is missing.
  - `attester` (optional): when VTS runs in a TEE (e.g., an AWS Nitro enclave),
    this directive specifies which attester is to be used to produce evidence
    for the EAR signing key.  The supported attesters are: `nitro`.

## Startup

```mermaid
sequenceDiagram
box VTS
participant vts-service
participant earsigner
participant JWT earsigner
end
box Platform RoT
participant Platform Attester
end

vts-service->>earsigner: New(cfg)
activate earsigner
earsigner->>JWT earsigner: Init(cfg)
alt cfg.key is set
  JWT earsigner->>File System: loadKey(cfg.key)
  File System->>JWT earsigner: (sk, pk)
else
  JWT earsigner->>JWT earsigner: (sk, pk) = generateKey(cfg.alg)
end
alt cfg.attest is set
  JWT earsigner->>Platform Attester: Attest(pk)
  Platform Attester->>JWT earsigner: kat (key attestation token)
  Note over JWT earsigner,Platform Attester: In what cases does it<br/>make sense to attest<br/>a pre-provisioned key?
end
JWT earsigner->>JWT earsigner: store pk, sk, alg, kat into context
```

## Run-time via `.well-known/veraison`

```mermaid
sequenceDiagram
actor API client
box VFE
participant verification-service
end
box VTS
participant vts-service
participant JWT earsigner
end

API client->>verification-service: GET /.well-known/veraison
verification-service->>vts-service: GetEARSigningPublicKey()
vts-service->>JWT earsigner: GetPublicKeyInfo()
JWT earsigner->>vts-service: alg, pk, kat
vts-service->>verification-service: alg, pk, kat
verification-service->>API client: JSON{ ..., alg, pk, kat }
```

## Run-time via `challenge-response/v1/newSession`

```mermaid
sequenceDiagram
actor API client
box VFE
participant verification-service
end
box VTS
participant vts-service
participant JWT earsigner
end

API client->>verification-service: POST /challenge-response/v1/newSession?tee-report=full
verification-service->>API client: 201, Location: /challenge-response/v1/session/<ID>
Note over API client,verification-service: time passes
API client->>verification-service: POST /challenge-response/v1/session/<ID>
verification-service->>vts-service: ProcessEvidence(..., session.teeReport)
vts-service->>JWT earsigner: GetPublicKeyInfo()
JWT earsigner->>vts-service: alg, pk, kat
vts-service->>vts-service: EAR claims += pk, hash(kat)
alt teeReport==full
vts-service->>vts-service: EAR claims += kat
end
vts-service->>JWT earsigner: Sign(EAR claims)
JWT earsigner->>vts-service: EAR
vts-service->>verification-service: EAR
verification-service->>API client: JSON{ ..., EAR }
```
