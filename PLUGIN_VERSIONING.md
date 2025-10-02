# Plugin Versioning Implementation (Issue #120)

## Summary

This PR implements versioning support for Veraison plugins and scheme implementations, addressing **Issue #120**.

**Issue Requirements:**
- ✅ Add versioning to pluggable interfaces
- ✅ Log versions during plugin discovery
- ✅ Expose versions via meta query interface (ServiceState API)
- ✅ Allow clients to understand level of support for particular formats

## Implementation Approach

**Scheme-Level Versioning:** All pluggables associated with a scheme (evidence handler, endorsement handler, store handler) share the same version, consistent with how they share the scheme name.

## Key Changes

### 1. Core Plugin Interface
- Added `GetVersion() string` to `IPluggable` interface
- All plugins must now return semantic version (e.g., "1.0.0")

### 2. Version Exposure Methods

**Manager Interface (`plugin/imanager.go`):**
```go
GetPluginVersion(name string) (string, error)
GetSchemeVersion(scheme string) (string, error)
```

**ServiceState API (proto/state.proto):**
```protobuf
map<string, string> scheme_versions = 4;
```

### 3. Logging
Plugins now log version at INFO level during discovery:
```
loaded plugin | name=psa-evidence-handler scheme=PSA_IOT version=1.0.0 path=...
```

### 4. All Schemes Updated

| Scheme | Version |
|--------|---------|
| PSA_IOT | 1.0.0 |
| ARM_CCA | 1.0.0 |
| PARSEC_CCA | 1.0.0 |
| PARSEC_TPM | 1.0.0 |
| RIOT | 1.0.0 |
| TPM_ENACTTRUST | 1.0.0 |

## Files Modified (39 total)

**Core Framework:**
- `plugin/ipluggable.go` - Added GetVersion() method
- `plugin/goplugin_context.go` - Version tracking
- `plugin/goplugin_loader.go` - Version logging
- `plugin/goplugin_manager.go` - Version query methods
- `plugin/imanager.go` - Manager interface extension
- `builtin/builtin_loader.go` - Version logging
- `builtin/builtin_manager.go` - Version query methods

**RPC Support:**
- `handler/evidence_rpc.go`
- `handler/endorsement_rpc.go`
- `handler/store_rpc.go`

**All 6 Scheme Implementations:**
- Added `SchemeVersion` constants
- Implemented `GetVersion()` in all handlers

**Test Infrastructure:**
- Updated test plugins and interfaces

## Usage Examples

### Query via Manager
```go
version, err := pluginManager.GetSchemeVersion("PSA_IOT")
// Returns: "1.0.0"
```

### Query via API
```
GET /verification/v1/state
```
Response includes:
```json
{
  "scheme_versions": {
    "PSA_IOT": "1.0.0",
    "ARM_CCA": "1.0.0",
    ...
  }
}
```

## Build Status
✅ All packages compile successfully
✅ No errors or warnings

## Resolves
Closes #120
