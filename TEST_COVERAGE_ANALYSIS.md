# Test Coverage Re-evaluation Report

## Issue #252: Re-evaluate test coverage

**Author**: Abhijit Das (Sukuna0007Abhi)  
**Date**: October 4, 2025  
**Branch**: test-coverage  

## Executive Summary

This report provides a comprehensive analysis of the packages currently excluded from test coverage in the Veraison services repository. The analysis examines each excluded package to determine whether additional unit tests should be added to improve code quality and maintainability.

## Current Coverage Exclusions

The following packages are currently excluded from coverage checks via `IGNORE_COVERAGE` in the top-level Makefile:

### Category 1: Plugin-related packages (justified exclusions)
1. `github.com/veraison/services/plugin` - Tested via plugin/test package
2. `github.com/veraison/services/plugin/test` - Pure test code

### Category 2: Protobuf-generated code (justified exclusion)
3. `github.com/veraison/services/handler` - Contains protobuf-generated code

### Category 3: Packages without tests (Go 1.22+ reports 0% coverage)
4. `github.com/veraison/services/builtin`
5. `github.com/veraison/services/management/api`
6. `github.com/veraison/services/management/cmd/management-service`
7. `github.com/veraison/services/provisioning/cmd/provisioning-service`
8. `github.com/veraison/services/provisioning/provisioner`
9. `github.com/veraison/services/scheme/common`
10. `github.com/veraison/services/scheme/common/arm`
11. `github.com/veraison/services/verification/cmd/verification-service`
12. `github.com/veraison/services/verification/verifier`
13. `github.com/veraison/services/vts/cmd/vts-service`
14. `github.com/veraison/services/vts/trustedservices`
15. `github.com/veraison/services/vtsclient`

## Analysis and Recommendations

### HIGH PRIORITY: Should Add Unit Tests

#### 1. `builtin` package
**Current State**: No unit tests  
**Functionality**: 
- BuiltinManager with generic type support
- BuiltinLoader for plugin discovery
- Media type registration and lookup
- Attestation scheme handling

**Recommendation**: **ADD UNIT TESTS**
- **Reason**: Contains substantial business logic for plugin management
- **Test Areas**:
  - BuiltinManager creation and initialization
  - Media type registration and lookup
  - Attestation scheme registration
  - Error handling in CreateBuiltinManager functions
  - Generic type behavior validation

#### 2. `vtsclient` package  
**Current State**: No unit tests  
**Functionality**:
- gRPC client implementation for VTS communication
- Connection management with credentials
- Error handling with custom error types (NoConnectionError)

**Recommendation**: **ADD UNIT TESTS**
- **Reason**: Critical communication layer with complex error handling
- **Test Areas**:
  - GRPC client creation and configuration
  - Connection establishment with different credential types
  - Custom error type behavior (NoConnectionError)
  - gRPC call handling and error propagation

#### 3. `provisioning/provisioner` package
**Current State**: No unit tests, but tested via API layer  
**Functionality**:
- Media type support validation
- VTS client interaction
- Input parameter validation

**Recommendation**: **ADD UNIT TESTS**
- **Reason**: Contains business logic that should be unit tested independently
- **Test Areas**:
  - IsSupportedMediaType logic
  - Input parameter validation (ErrInputParam cases)
  - VTS client interaction mocking
  - SubmitCoRIM functionality

#### 4. `verification/verifier` package
**Current State**: No unit tests, but tested via API layer  
**Functionality**:
- Media type support validation
- Evidence processing
- VTS state retrieval

**Recommendation**: **ADD UNIT TESTS**
- **Reason**: Core verification logic should be independently testable
- **Test Areas**:
  - IsSupportedMediaType validation
  - ProcessEvidence functionality
  - GetVTSState error handling
  - Input parameter validation

#### 5. `scheme/common` package
**Current State**: No unit tests  
**Functionality**:
- CCA platform/realm claim wrappers
- PSA platform claim wrappers
- Claims to map conversion utilities
- Certificate parsing utilities

**Recommendation**: **ADD UNIT TESTS**
- **Reason**: Contains utility functions with complex JSON marshaling logic
- **Test Areas**:
  - ClaimMapper implementations for CCA and PSA
  - ClaimsToMap conversion function
  - Certificate parsing (ParseCertificates function)
  - JSON marshaling of different claim types

### MEDIUM PRIORITY: Consider Adding Unit Tests

#### 6. `vts/trustedservices` package
**Current State**: No unit tests  
**Functionality**: 
- Large gRPC service implementation (681 lines)
- Contains substantial business logic
- Policy management, attestation processing

**Recommendation**: **CONSIDER UNIT TESTS**
- **Reason**: While primarily integration-tested, some business logic could benefit from unit tests
- **Areas to Consider**:
  - Configuration validation
  - Error handling logic
  - State management functions
  - Individual service method logic (mocked dependencies)

### LOW PRIORITY: Maintain Current Exclusion

#### 7. `management/api` package
**Current State**: No unit tests  
**Recommendation**: **LOW PRIORITY**
- **Reason**: Likely thin API layer, already covered by integration tests

#### 8. All `cmd/*-service` packages
**Current State**: No unit tests  
**Recommendation**: **MAINTAIN EXCLUSION**
- **Reason**: Main entry points - better tested via integration tests
- **Packages**: management/cmd/management-service, provisioning/cmd/provisioning-service, verification/cmd/verification-service, vts/cmd/vts-service

#### 9. `scheme/common/arm` package
**Current State**: No unit tests  
**Recommendation**: **LOW PRIORITY**
- **Reason**: Needs investigation - directory structure suggests it may be scheme-specific utilities

## Current Test Coverage Status

**Coverage check is currently passing** with threshold of 60.0%

The existing test coverage includes:
- Integration tests covering end-to-end workflows
- API handler tests with mocked dependencies
- Unit tests for non-excluded packages

## Implementation Plan

### Phase 1: High Priority Packages
1. **builtin** package - Focus on BuiltinManager and BuiltinLoader
2. **vtsclient** package - Focus on gRPC client and error handling
3. **provisioning/provisioner** package - Focus on business logic
4. **verification/verifier** package - Focus on verification logic
5. **scheme/common** package - Focus on claim processing utilities

### Phase 2: Medium Priority
1. **vts/trustedservices** package - Focus on core business logic functions

### Testing Strategy
- Use dependency injection with interfaces to enable mocking
- Focus on business logic rather than integration scenarios
- Maintain the existing integration test coverage
- Use table-driven tests for validation logic
- Mock external dependencies (VTS clients, gRPC connections)

## Benefits of Adding Unit Tests

1. **Earlier Error Detection**: Catch bugs during development
2. **Refactoring Safety**: Enable confident code changes
3. **Documentation**: Tests serve as usage examples
4. **Code Quality**: Force consideration of edge cases and error conditions
5. **Debugging**: Easier to isolate issues to specific functions

## Conclusion

While the current integration test coverage provides good end-to-end validation, adding unit tests to the **5 high-priority packages** would significantly improve the codebase's maintainability and robustness. These packages contain substantial business logic that would benefit from isolated testing.

The exclusions for plugin-related packages, protobuf-generated code, and main entry points should be maintained as they are appropriately tested through other means.