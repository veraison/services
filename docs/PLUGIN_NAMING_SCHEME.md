# Veraison Plugin Naming Scheme Specification

## Overview

This document defines the standardized naming conventions for Veraison attestation scheme plugins to ensure consistency, avoid collisions, and provide clear identification of plugin capabilities.

## Naming Scheme Format

### Plugin Names (`GetName()`)

The plugin name follows this pattern:
```
veraison/<scheme>/<handler-type>
```

Where:
- **`veraison/`**: Vendor prefix to avoid collisions with third-party implementations
- **`<scheme>`**: Lowercase scheme identifier (e.g., `psa-iot`, `arm-cca`, `parsec-tpm`)
- **`<handler-type>`**: Handler type (`endorsement`, `evidence`, `store`)

### Examples:
- `veraison/psa-iot/evidence` - PSA IoT evidence handler
- `veraison/arm-cca/endorsement` - ARM CCA endorsement handler  
- `veraison/parsec-tpm/store` - Parsec TPM store handler

### Attestation Scheme Names (`GetAttestationScheme()`)

The attestation scheme name follows this pattern:
```
<VENDOR>_<TECHNOLOGY>_<VARIANT>
```

Where:
- **`<VENDOR>`**: Uppercase vendor/standard identifier (e.g., `ARM`, `PARSEC`)
- **`<TECHNOLOGY>`**: Uppercase technology identifier (e.g., `PSA`, `CCA`, `TPM`)
- **`<VARIANT>`**: Uppercase variant identifier (e.g., `IOT`, `ENACTTRUST`)

### Examples:
- `ARM_PSA_IOT` - ARM PSA IoT profile
- `CCA` - ARM Confidential Compute Architecture
- `PARSEC_TPM` - Parsec TPM-based attestation
- `TPM_ENACTTRUST` - TPM EnactTrust profile
- `RIOT` - RIoT-based DICE attestation

## Scheme Directory Naming

Directory names should use lowercase with hyphens:
- `psa-iot/` (was `psa-iot/`)
- `arm-cca/` (was `arm-cca/`) 
- `parsec-tpm/` (was `parsec-tpm/`)
- `parsec-cca/` (was `parsec-cca/`)
- `tpm-enacttrust/` (was `tmp-enacttrust/`)
- `riot/` (was `riot/`)

## Migration Strategy

### Phase 1: Update Constants
1. Update `SchemeName` constants to follow the new format
2. Update all `GetName()` implementations to use the new format
3. Update corresponding test cases

### Phase 2: Deprecation Support
1. Add backward compatibility for existing names
2. Log warnings for deprecated name usage
3. Update documentation with migration timeline

### Phase 3: Cleanup
1. Remove deprecated name support after transition period
2. Update all references and documentation

## Benefits

1. **Collision Avoidance**: Vendor prefixing prevents conflicts with third-party plugins
2. **Consistency**: Uniform naming patterns across all schemes
3. **Discoverability**: Clear identification of plugin capabilities
4. **Extensibility**: Format supports future versioning if needed
5. **Maintainability**: Easier to track and manage plugin implementations

## Implementation Notes

- All existing functionality remains unchanged except naming
- Plugin lookup by both old and new names supported during transition
- Tests updated to validate new naming scheme
- Documentation updated to reflect new standards

## Future Considerations

### Versioning Support
If versioning becomes required, the format can be extended:
```
veraison/<scheme>/<handler-type>/v<major>.<minor>
```

Example: `veraison/psa-iot/evidence/v1.0`

### Third-Party Plugins
Third-party developers should use their own vendor prefix:
```
<vendor>/<scheme>/<handler-type>
```

Example: `acme/custom-tpm/evidence`