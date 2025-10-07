# Plugin Naming Scheme

This document describes the standardized naming scheme for Veraison plugins to prevent collisions and ensure consistency.

## Format Specification

### Plugin Names
Plugin names follow the format: `veraison/{technology}/{handler-type}`

Where:
- `technology`: The attestation technology (e.g., `psa-iot`, `cca`, `parsec-cca`, `parsec-tpm`, `tpm-enacttrust`, `riot`)
- `handler-type`: The type of handler with suffix:
  - `.evidence` - Evidence handlers
  - `.endorsement` - Endorsement handlers  
  - `.store` - Store handlers

### Scheme Constants
Scheme constants follow the pattern: `<VENDOR>_<TECHNOLOGY>_<VARIANT>`

Where:
- `VENDOR`: Technology vendor (optional, only when necessary for disambiguation)
- `TECHNOLOGY`: Core technology name
- `VARIANT`: Technology variant or version

## Examples

### PSA IoT
- Scheme constant: `PSA_IOT`
- Plugin names:
  - `veraison/psa-iot/evidence`
  - `veraison/psa-iot/endorsement`
  - `veraison/psa-iot/store`

### ARM CCA
- Scheme constant: `CCA`
- Plugin names:
  - `veraison/cca/evidence`
  - `veraison/cca/endorsement`
  - `veraison/cca/store`

### Parsec CCA
- Scheme constant: `PARSEC_CCA`
- Plugin names:
  - `veraison/parsec-cca/evidence`
  - `veraison/parsec-cca/endorsement`
  - `veraison/parsec-cca/store`

### Parsec TPM
- Scheme constant: `PARSEC_TPM`
- Plugin names:
  - `veraison/parsec-tpm/evidence`
  - `veraison/parsec-tpm/endorsement`
  - `veraison/parsec-tpm/store`

### TPM EnactTrust
- Scheme constant: `TPM_ENACTTRUST`
- Plugin names:
  - `veraison/tpm-enacttrust/evidence`
  - `veraison/tpm-enacttrust/endorsement`
  - `veraison/tpm-enacttrust/store`

### RIoT
- Scheme constant: `RIOT`
- Plugin names:
  - `veraison/riot/evidence`
  - `veraison/riot/endorsement`
  - `veraison/riot/store`

## Benefits

1. **Collision Prevention**: Consistent naming prevents plugin name conflicts
2. **Clarity**: Clear indication of technology and handler type
3. **Consistency**: Uniform pattern across all schemes
4. **Scalability**: Easy to extend for new technologies