# Plugin Naming Migration Guide

This guide helps developers migrate to the new standardized plugin naming scheme.

## Migration Overview

The new naming scheme standardizes plugin names to `veraison/{technology}/{handler-type}` format to prevent collisions and improve consistency.

## Changes Required

### 1. Handler GetName() Methods

Update all handler `GetName()` methods to return standardized names:

#### Before:
```go
func (s EvidenceHandler) GetName() string {
    return "psa-evidence-handler"
}
```

#### After:
```go
func (s EvidenceHandler) GetName() string {
    return "veraison/psa-iot/evidence"
}
```

### 2. Scheme Constants

Update scheme constants to follow consistent patterns:

#### Before:
```go
const SchemeName = "ARM_CCA"
```

#### After:
```go
const SchemeName = "CCA"
```

### 3. Test Expectations

Update test cases to expect new naming:

#### Before:
```go
expectedName := "psa-evidence-handler"
```

#### After:
```go
expectedName := "veraison/psa-iot/evidence"
```

## Per-Scheme Migration

### PSA IoT
- Directory: `scheme/psa-iot/`
- Scheme: `PSA_IOT` (no change)
- Handlers:
  - Evidence: `psa-evidence-handler` → `veraison/psa-iot/evidence`
  - Endorsement: `psa-endorsement-handler` → `veraison/psa-iot/endorsement`
  - Store: `psa-store-handler` → `veraison/psa-iot/store`

### ARM CCA (renamed to CCA)
- Directory: `scheme/arm-cca/` → `scheme/cca/`
- Scheme: `ARM_CCA` → `CCA`
- Package: `arm_cca` → `cca`
- Handlers:
  - Evidence: `arm-cca-evidence-handler` → `veraison/cca/evidence`
  - Endorsement: `arm-cca-endorsement-handler` → `veraison/cca/endorsement`
  - Store: `arm-cca-store-handler` → `veraison/cca/store`

### Parsec CCA
- Directory: `scheme/parsec-cca/`
- Scheme: `PARSEC_CCA` (no change)
- Handlers:
  - Evidence: `parsec-cca-evidence-handler` → `veraison/parsec-cca/evidence`
  - Endorsement: `parsec-cca-endorsement-handler` → `veraison/parsec-cca/endorsement`
  - Store: `parsec-cca-store-handler` → `veraison/parsec-cca/store`

### Parsec TPM
- Directory: `scheme/parsec-tpm/`
- Scheme: `PARSEC_TPM` (no change)
- Handlers:
  - Evidence: `parsec-tpm-evidence-handler` → `veraison/parsec-tpm/evidence`
  - Endorsement: `parsec-tpm-endorsement-handler` → `veraison/parsec-tpm/endorsement`
  - Store: `parsec-tpm-store-handler` → `veraison/parsec-tpm/store`

### TPM EnactTrust
- Directory: `scheme/tpm-enacttrust/`
- Scheme: `TPM_ENACTTRUST` (no change)
- Handlers:
  - Evidence: `tpm-enacttrust-evidence-handler` → `veraison/tpm-enacttrust/evidence`
  - Endorsement: `tpm-enacttrust-endorsement-handler` → `veraison/tpm-enacttrust/endorsement`
  - Store: `tpm-enacttrust-store-handler` → `veraison/tpm-enacttrust/store`

### RIoT
- Directory: `scheme/riot/`
- Scheme: `RIOT` (no change)
- Handlers:
  - Evidence: `riot-evidence-handler` → `veraison/riot/evidence`
  - Endorsement: `riot-endorsement-handler` → `veraison/riot/endorsement`
  - Store: `riot-store-handler` → `veraison/riot/store`

## Testing Migration

After making changes:

1. Run scheme-specific tests:
   ```bash
   make test
   ```

2. Run integration tests:
   ```bash
   make integ-test
   ```

3. Verify plugin loading:
   ```bash
   make docker-deploy
   ```

## Breaking Changes

This migration introduces breaking changes:
- Plugin names change format
- ARM CCA scheme renamed to CCA
- Directory structure changes for CCA scheme

Deployments using these plugins will need to update their configurations accordingly.