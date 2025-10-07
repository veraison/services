# Plugin Naming Scheme Standardization (Clean Implementation)

This PR implements the standardized plugin naming scheme to prevent collisions as requested in issue #123.

## 🎯 **What This PR Does**

This is a **clean implementation** that addresses all feedback from the original PR #346, specifically:
- ✅ Addresses @setrofim's consistency feedback about ARM CCA naming
- ✅ Contains **ONLY** plugin naming changes (no unrelated commits)
- ✅ Clean git history focused on the feature

## 📋 **Changes Summary**

### 1. **Standardized Plugin Naming Format**
Implemented format: `veraison/{technology}/{handler-type}`

**Before:**
- `psa-evidence-handler`
- `arm-cca-store-handler` 
- `parsec-tpm-endorsement-handler`

**After:**
- `veraison/psa-iot/evidence`
- `veraison/cca/store`
- `veraison/parsec-tpm/endorsement`

### 2. **Scheme Consistency (Addresses Feedback)**
- **ARM CCA** → **CCA** (removed vendor prefix for consistency)
- Directory renamed: `scheme/arm-cca` → `scheme/cca`
- Scheme constant: `ARM_CCA` → `CCA`

### 3. **All 6 Attestation Schemes Updated**

| Scheme | Constant | Handler Examples |
|--------|----------|------------------|
| PSA IoT | `PSA_IOT` | `veraison/psa-iot/{evidence,endorsement,store}` |
| CCA | `CCA` | `veraison/cca/{evidence,endorsement,store}` |
| Parsec CCA | `PARSEC_CCA` | `veraison/parsec-cca/{evidence,endorsement,store}` |
| Parsec TPM | `PARSEC_TPM` | `veraison/parsec-tpm/{evidence,endorsement,store}` |
| TPM EnactTrust | `TPM_ENACTTRUST` | `veraison/tpm-enacttrust/{evidence,endorsement,store}` |
| RIoT | `RIOT` | `veraison/riot/{evidence,store}` |

## 🔧 **Technical Changes**

- ✅ Updated all `GetName()` implementations in handler files
- ✅ Updated test expectations to match new naming scheme  
- ✅ Updated builtin schemes.gen.go
- ✅ Updated Makefiles and plugin configurations
- ✅ Updated integration test data
- ✅ Fixed all package names and import paths

## 📚 **Documentation Added**

- `docs/PLUGIN_NAMING_SCHEME.md` - Complete naming specification
- `docs/PLUGIN_NAMING_MIGRATION.md` - Developer migration guide

## ✅ **Testing & Verification**

- **Build**: All packages compile successfully
- **Unit Tests**: All `GetName()` tests pass with new naming
- **Scheme Tests**: Individual scheme tests pass
- **Integration Tests**: Updated and passing

**Verified working naming scheme:**
```
PSA IoT - Evidence: veraison/psa-iot/evidence
PSA IoT - Endorsement: veraison/psa-iot/endorsement
PSA IoT - Store: veraison/psa-iot/store

CCA - Evidence: veraison/cca/evidence
CCA - Endorsement: veraison/cca/endorsement
CCA - Store: veraison/cca/store

... (all schemes follow this pattern)
```

## 🔄 **Addresses Original PR Issues**

This clean implementation fixes the problems mentioned in the original PR #346:
- ❌ **OLD PR**: "picked up a bunch of commits that don't belong in this PR"
- ✅ **THIS PR**: Contains only plugin naming scheme changes
- ❌ **OLD PR**: Had merge conflicts and inconsistent naming
- ✅ **THIS PR**: Clean merge, addresses consistency feedback

## 🚀 **Ready for Review**

This PR is ready for review and merge. It:
- Implements the complete plugin naming standardization
- Addresses all reviewer feedback
- Has a clean git history
- Passes all tests
- Includes comprehensive documentation

**Fixes #123**