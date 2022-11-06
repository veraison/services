
# TPM key attestation data format

This document describes the data encoding for key attestation tokens produced using a TPM. The format relies on the [Attestation Object](https://www.w3.org/TR/webauthn/#sctn-attestation) format defined in WebAuthN, using the [TPM attestation statement format](https://www.w3.org/TR/webauthn-2/#sctn-tpm-attestation). All data structures below are expected to be encoded in [CTAP canonical CBOR encoding](https://fidoalliance.org/specs/fido-v2.0-ps-20190130/fido-client-to-authenticator-protocol-v2.0-ps-20190130.html#ctap2-canonical-cbor-encoding-form).

## Attestation Object

The high-level attestation object is generic over the attestation provider and offers a "metadata header" for the authenticator to describe itself and the key it is using for attestation.

CDDL representation of an Attestation Object (`attObj`):
```
attObj = {
            authData: bytes,
            $$attStmtType
         }

attStmtTemplate = (
                      fmt: text,
                      attStmt: { * tstr => any } ; Map is filled in by each concrete attStmtType
                  )
```

`authData`: Encodes contextual bindings made by the authenticator. For this use-case, the data is irrelevant and thus the `authData` field should be omitted.

## TPM Attestation Statement Format

The attestation statement format defined for TPM backends is defined below.

CDDL representation of TPM Attestation Statement Format:
```
$$attStmtType // = (
                    fmt: "tpm",
                    attStmt: tpmStmtFormat
                )

tpmStmtFormat = {
                 ver: "2.0",
                 (
                     alg: COSEAlgorithmIdentifier,
                     x5c: [ aikCert: bytes, * (caCert: bytes) ]
                 )
                 sig: bytes,
                 certInfo: bytes,
                 pubArea: bytes
             }
```


- `ver`: The version of the TPM specification to which the signature conforms.
- `alg`: A COSEAlgorithmIdentifier containing the identifier of the algorithm used to generate the attestation signature.
- `x5c`: `aikCert` followed by its certificate chain, in X.509 encoding.
    - `aikCert`: The AIK certificate used for the attestation, in X.509 encoding.
- `sig`: The attestation signature, in the form of a TPMT_SIGNATURE structure as specified in [TPMv2-Part2](https://trustedcomputinggroup.org/wp-content/uploads/TCG_TPM2_r1p59_Part2_Structures_pub.pdf) section 11.3.4.
- `certInfo`: The TPMS_ATTEST structure over which the above signature was computed, as specified in [TPMv2-Part2](https://trustedcomputinggroup.org/wp-content/uploads/TCG_TPM2_r1p59_Part2_Structures_pub.pdf) section 10.12.8.
- `pubArea`: The TPMT_PUBLIC structure (see [TPMv2-Part2](https://trustedcomputinggroup.org/wp-content/uploads/TCG_TPM2_r1p59_Part2_Structures_pub.pdf) section 12.2.4) used by the TPM to represent the credential public key.

The signature (`sig`) described above is that produced using the TPM2_Certify operation. In order to ensure freshness, the WebAuthN spec mandates the use of `extraData` in the operation. This extra data is obtained by hashing `attToBeSigned`, which represents the concatenation of `authData` and `clientDataHash`. Given that neither of these two fields is relevant in our case, `extraData` will instead contain a nonce provided by the relying party - `relyingPartyNonce`.

## Verification procedure

Given the verification procedure inputs `attStmt` and `relyingPartyNonce`, the verification procedure is as follows:

- Verify that `attStmt` is valid CBOR conforming to the syntax defined above and perform CBOR decoding on it to extract the contained fields.

- Verify that `x5c` is present.

- Verify the `sig` is a valid signature over `certInfo` using the attestation public key in `aikCert` with the algorithm specified in `alg`.

- Verify that `aikCert` meets the requirements in § 8.3.1 TPM Attestation Statement Certificate Requirements.

- Validate that `certInfo` is valid:

    * Verify that `magic` is set to TPM_GENERATED_VALUE.

    * Verify that `type` is set to TPM_ST_ATTEST_CERTIFY.

    * Verify that `extraData` is set to `relyingPartyNonce`.

    * Verify that `attested` contains a TPMS_CERTIFY_INFO structure as specified in [TPMv2-Part2](https://trustedcomputinggroup.org/wp-content/uploads/TCG_TPM2_r1p59_Part2_Structures_pub.pdf) section 10.12.3, whose name field contains a valid Name for `pubArea`, as computed using the algorithm in the `nameAlg` field of `pubArea` using the procedure specified in [TPMv2-Part1] section 16.

    * Note that the remaining fields in the "Standard Attestation Structure" [TPMv2 Part1](https://trustedcomputinggroup.org/wp-content/uploads/TCG_TPM2_r1p59_Part1_Architecture_pub.pdf) section 31.2, i.e., `qualifiedSigner`, `clockInfo` and `firmwareVersion` are ignored. These fields MAY be used as an input to risk engines.

- If successful, return implementation-specific values representing attestation type AttCA and attestation trust path `x5c`.

**Note**: The above steps only serve as verification for the key attestation steps, **not** for platform attestation.

## Authenticator Data (**UNUSED**)

This section describes the structure of authenticator data normally included in the attestation object. This is done for completeness, as authenticator data is _not relevant to our use-case_. Structure:

| Name                   | Length (in bytes)     | Description                                            |
|------------------------|-----------------------|--------------------------------------------------------|
| rpIdHash               | 32                    | SHA-256 hash of the RP ID the credential is scoped to. |
| flags                  | 1                     | See flag table below.                                  |
| signCount              | 4                     | Signature counter, 32-bit unsigned big-endian integer. |
| attestedCredentialData | variable (if present) | Attested credential data (if present). <br>Its length depends on the length of the credential ID and credential public key being attested. |
| extensions             | variable (if present) | Extension-defined authenticator data. <br> This is a [CBOR](https://www.rfc-editor.org/rfc/rfc8949.html) map with extension identifiers as keys, and authenticator extension outputs as values. |

Authenticator data flags (bit 0 is the least significant bit):

| Name                                   | Bit position | Description                                                         |
|----------------------------------------|--------------|---------------------------------------------------------------------|
| User Present (UP)                      | 0            | 1 means the user is present                                         |
| Reserved for future use (RFU1)         | 1            | N/A                                                                 |
| User Verified (UV)                     | 2            | 1 means the user is verified                                        |
| Reserved for future use (RFU2)         | 3-5          | N/A                                                                 |
| Attested credential data included (AT) | 6            | Indicates whether the authenticator added attested credential data. |
| Extension data included (ED)           | 7            | Indicates if the authenticator data has extensions.                 |

For our use case, `rpIdHash`, `signCount`, and `extensions` are (probably?) useless.

`attestedCredentialData` needs to be passed along to identify the key that was used during attestation. The bytes in that field have the following structure:

| Name                | Length (in bytes) | Description                                                           |
|---------------------|-------------------|-----------------------------------------------------------------------|
| aaguid              | 16                | The AAGUID of the authenticator.                                      |
| credentialLength    | 2                 | Byte length `L` of Credential ID, 16-bit unsigned big-endian integer. |
| credentialId        | `L`               | Credential ID                                                         |
| credentialPublicKey | variable          | The credential public key encoded in COSE_Key format, as defined in [Section 7 of RFC8152](https://datatracker.ietf.org/doc/html/rfc8152#section-7), using the CTAP2 canonical CBOR encoding form. The COSE_Key-encoded credential public key MUST contain the "alg" parameter and MUST NOT contain any other OPTIONAL parameters. The "alg" parameter MUST contain a COSEAlgorithmIdentifier value. The encoded credential public key MUST also contain any additional REQUIRED parameters stipulated by the relevant key type specification, i.e., REQUIRED for the key type "kty" and algorithm "alg" (see [Section 8 of RFC8152](https://datatracker.ietf.org/doc/html/rfc8152#section-8)). |

The `credentialId` is just a "probabilistically-unique byte sequence identifying a public key credential source and its authentication assertions, [...] at least 16 bytes that include at least 100 bits of entropy." Whether we want to actually have a value for this, or just leave it as `0x00` is up to us - presumably Veraison could identify the key using the `credentialPublicKey` (?). We also get the AIK certificate, as described below, in `x5c`. So not sure what the value of either of these fields is, actually.