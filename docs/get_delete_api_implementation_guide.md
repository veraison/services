# GET & DELETE APIs for Veraison Provisioning Interface

## Summary

This implementation adds GET and DELETE functionality to the Veraison Provisioning Interface, allowing users to:
1. **Retrieve endorsements** (reference values and trust anchors) that have been submitted
2. **Delete endorsements** when they are no longer needed

Previously, the interface only supported submitting endorsements via POST.

## Implementation Details

### 1. Protocol Buffer Definitions

Added new message types and RPC methods to `proto/vts.proto`:

- **GetEndorsementsRequest**: Request to retrieve endorsements with optional filtering
  - `key_prefix`: Filter by key prefix (optional)
  - `endorsement_type`: Type filter ("trust-anchor", "reference-value", or "all")

- **GetEndorsementsResponse**: Response containing matching endorsements
  - `endorsements`: List of endorsement entries
  - `status`: Operation status

- **DeleteEndorsementsRequest**: Request to delete endorsements
  - `key`: Key or key prefix to delete (required)
  - `endorsement_type`: Type filter ("trust-anchor", "reference-value", or "all")

- **DeleteEndorsementsResponse**: Response with deletion results
  - `deleted_count`: Number of endorsements deleted
  - `status`: Operation status

### 2. VTS Service Layer

**File**: `vts/trustedservices/trustedservices_grpc.go`

Implemented two new gRPC methods:

- **GetEndorsements**: Retrieves endorsements from trust anchor and/or reference value stores
  - Supports filtering by key prefix
  - Supports filtering by endorsement type
  - Returns all matching endorsements with their keys and values

- **DeleteEndorsements**: Deletes endorsements from stores
  - Supports exact key match or prefix-based deletion
  - Supports filtering by endorsement type
  - Returns count of deleted entries

### 3. Provisioner Layer

**Files**: 
- `provisioning/provisioner/iprovisioner.go`
- `provisioning/provisioner/provisioner.go`

Added two new methods to the provisioner interface and implementation that call the VTS gRPC methods.

### 4. Provisioning API Layer

**Files**:
- `provisioning/api/handler.go`
- `provisioning/api/router.go`

Added two new REST API endpoints:

#### GET /endorsement-provisioning/v1/endorsements

Query parameters:
- `key-prefix` (optional): Filter endorsements by key prefix
- `type` (optional): Filter by type ("all", "trust-anchor", "reference-value"). Default: "all"

Response: JSON with endorsements list

Example:
```bash
curl -H "Accept: application/json" \
  "http://localhost:8888/endorsement-provisioning/v1/endorsements?type=trust-anchor"
```

#### DELETE /endorsement-provisioning/v1/endorsements

Query parameters:
- `key` (required): Key or key prefix to delete
- `type` (optional): Filter by type ("all", "trust-anchor", "reference-value"). Default: "all"

Response: JSON with deletion count

Example:
```bash
curl -X DELETE \
  "http://localhost:8888/endorsement-provisioning/v1/endorsements?key=my-endorsement-key"
```

### 5. VTS Client Layer

**File**: `vtsclient/vtsclient_grpc.go`

Added wrapper methods in the GRPC client to call the new VTS service methods.

### 6. Tests

**File**: `provisioning/api/handler_test.go`

Added comprehensive unit tests:
- `TestHandler_GetEndorsements_Success`: Test successful retrieval
- `TestHandler_GetEndorsements_WithFilter`: Test with filtering
- `TestHandler_GetEndorsements_InvalidType`: Test invalid type parameter
- `TestHandler_DeleteEndorsements_Success`: Test successful deletion
- `TestHandler_DeleteEndorsements_MissingKey`: Test missing key parameter
- `TestHandler_DeleteEndorsements_Error`: Test error handling

All tests pass successfully.

## API Usage Examples

### Retrieve All Endorsements

```bash
curl -H "Accept: application/json" \
  "http://localhost:8888/endorsement-provisioning/v1/endorsements"
```

### Retrieve Trust Anchors Only

```bash
curl -H "Accept: application/json" \
  "http://localhost:8888/endorsement-provisioning/v1/endorsements?type=trust-anchor"
```

### Retrieve Endorsements with Key Prefix

```bash
curl -H "Accept: application/json" \
  "http://localhost:8888/endorsement-provisioning/v1/endorsements?key-prefix=arm-cca"
```

### Delete Specific Endorsement

```bash
curl -X DELETE \
  "http://localhost:8888/endorsement-provisioning/v1/endorsements?key=my-key"
```

### Delete All Trust Anchors with Prefix

```bash
curl -X DELETE \
  "http://localhost:8888/endorsement-provisioning/v1/endorsements?key=prefix-&type=trust-anchor"
```

## Security Considerations

- All new endpoints are protected by the same authorization middleware as the existing submit endpoint
- Users must have the `ProvisionerRole` to access these endpoints
- Deletion is permanent and cannot be undone
- Key prefix matching allows batch deletion - use with caution

## Files Modified

1. `proto/vts.proto` - Proto definitions
2. `proto/vts_grpc.pb.go` - Generated gRPC code (manually updated)
3. `proto/endorsement_query.go` - New proto message structures
4. `vts/trustedservices/trustedservices_grpc.go` - VTS service implementation
5. `provisioning/provisioner/iprovisioner.go` - Provisioner interface
6. `provisioning/provisioner/provisioner.go` - Provisioner implementation
7. `provisioning/api/handler.go` - API handlers
8. `provisioning/api/router.go` - Route registration
9. `provisioning/api/handler_test.go` - Unit tests
10. `provisioning/api/mocks/iprovisioner.go` - Mock updates
11. `vtsclient/vtsclient_grpc.go` - VTS client wrapper

## Testing

All unit tests pass:
```bash
cd provisioning/api && go test -v
```

Build verification:
```bash
go build ./provisioning/...
go build ./vts/...
```
