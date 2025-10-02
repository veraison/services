# Plugin Naming Scheme Migration Guide

This guide explains how to migrate from the old ad-hoc plugin naming scheme to the new standardized Veraison plugin naming conventions.

## Summary of Changes

### Plugin Names (GetName())

#### Old Naming Patterns:
- Inconsistent formats: `"corim (PSA profile)"`, `"psa-evidence-handler"`
- No vendor prefixing
- Mixed case and separators

#### New Naming Scheme:
- Consistent format: `veraison/<scheme>/<handler-type>`
- Vendor prefix prevents collisions
- Lowercase with hyphens for readability

### Examples:

| Scheme | Handler | Old Name | New Name |
|--------|---------|----------|----------|
| PSA IoT | Evidence | `"psa-evidence-handler"` | `"veraison/psa-iot/evidence"` |
| PSA IoT | Endorsement | `"corim (PSA profile)"` | `"veraison/psa-iot/endorsement"` |
| PSA IoT | Store | `"psa-store-handler"` | `"veraison/psa-iot/store"` |
| ARM CCA | Evidence | `"cca-evidence-handler"` | `"veraison/arm-cca/evidence"` |
| ARM CCA | Endorsement | `"corim (CCA platform profile)"` | `"veraison/arm-cca/endorsement"` |
| ARM CCA | Store | `"cca-store-handler"` | `"veraison/arm-cca/store"` |
| Parsec CCA | Evidence | `"parsec-cca-evidence-handler"` | `"veraison/parsec-cca/evidence"` |
| Parsec CCA | Endorsement | `"corim (Parsec CCA profile)"` | `"veraison/parsec-cca/endorsement"` |
| Parsec CCA | Store | `"parsec-cca-store-handler"` | `"veraison/parsec-cca/store"` |
| Parsec TPM | Evidence | `"parsec-tpm-evidence-handler"` | `"veraison/parsec-tpm/evidence"` |
| Parsec TPM | Endorsement | `"corim (Parsec TPM profile)"` | `"veraison/parsec-tpm/endorsement"` |
| Parsec TPM | Store | `"parsec-tpm-store-handler"` | `"veraison/parsec-tpm/store"` |
| TPM EnactTrust | Evidence | `"tpm-enacttrust-evidence-handler"` | `"veraison/tpm-enacttrust/evidence"` |
| TPM EnactTrust | Endorsement | `"corim (TPM EnactTrust profile)"` | `"veraison/tpm-enacttrust/endorsement"` |
| TPM EnactTrust | Store | `"tpm-enacttrust-store-handler"` | `"veraison/tpm-enacttrust/store"` |
| RIoT | Evidence | `"riot-evidence-handler"` | `"veraison/riot/evidence"` |
| RIoT | Store | `"riot-store-handler"` | `"veraison/riot/store"` |

### Attestation Scheme Names (GetAttestationScheme())

#### Updated for Consistency:
| Scheme | Old Name | New Name |
|--------|----------|----------|
| PSA IoT | `"PSA_IOT"` | `"ARM_PSA_IOT"` |
| RIoT | `"riot"` | `"RIOT"` |
| Others | No change | No change |

## Implementation Details

### Constants Added to scheme.go Files

Each scheme now defines constants for consistent naming:

```go
const (
    // SchemeName follows the format: <VENDOR>_<TECHNOLOGY>_<VARIANT>
    SchemeName = "ARM_PSA_IOT"
    
    // Plugin name constants following the format: veraison/<scheme>/<handler-type>
    EvidenceHandlerName    = "veraison/psa-iot/evidence"
    EndorsementHandlerName = "veraison/psa-iot/endorsement" 
    StoreHandlerName       = "veraison/psa-iot/store"
)
```

### Updated Handler Implementations

All `GetName()` methods now use the constants:

```go
func (s EvidenceHandler) GetName() string {
    return EvidenceHandlerName
}
```

### Test Updates

Test expectations updated to use the constants:

```go
expectedName := EvidenceHandlerName
assert.Equal(t, name, expectedName)
```

## Backward Compatibility

While this change updates the naming scheme, the actual plugin functionality remains unchanged. The new names:

1. Provide better collision avoidance
2. Follow consistent patterns
3. Enable easier plugin discovery
4. Support future extensibility

## Benefits

1. **Collision Avoidance**: Vendor prefixing prevents name conflicts
2. **Consistency**: Uniform format across all schemes
3. **Readability**: Clear hierarchy and purpose identification
4. **Maintainability**: Easier to manage and understand plugin ecosystem
5. **Extensibility**: Format supports future enhancements like versioning

## Migration Impact

- **Breaking Change**: Plugin names have changed
- **API Impact**: GetName() returns different values
- **Test Updates**: Test expectations updated accordingly
- **Documentation**: Updated to reflect new naming conventions

## Future Enhancements

The new naming scheme supports future extensions:

1. **Versioning**: `veraison/psa-iot/evidence/v1.0`
2. **Third-party plugins**: `<vendor>/<scheme>/<handler-type>`
3. **Variants**: `veraison/psa-iot-secure/evidence`