# Installation Metadata

## Overview

Veraison services support reading installation metadata to provide transparency about how an instance was deployed. This metadata is generated during installation/deployment and read at runtime.

## Metadata File Format

The installation metadata is a JSON file with the following structure:

```json
{
  "version": "1.0.0",
  "deployment_method": "deb",
  "install_time": "2025-10-02T14:30:00Z",
  "attestation_path": "/usr/share/doc/veraison/attestation.intoto.jsonl",
  "artifact_digest": "sha256:abc123...",
  "metadata": {
    "package": "veraison-services",
    "architecture": "amd64",
    "custom_field": "value"
  }
}
```

### Fields

- **version** (required): The version of Veraison being installed
- **deployment_method** (required): A string describing how the service was deployed (e.g., "deb", "rpm", "docker", "native", "source", or any custom method)
- **install_time** (optional): ISO 8601 timestamp of when the installation occurred
- **attestation_path** (optional): Path to the artifact attestation file (can be relative to metadata file location)
- **artifact_digest** (optional): SHA256 digest of the installation artifact
- **metadata** (optional): Additional key-value pairs specific to the deployment method

## Metadata File Locations

The system searches for metadata files in the following locations (in order):

1. `/etc/veraison/installation.json` - System-wide installation
2. `/usr/share/veraison/installation.json` - Package manager installations
3. `/opt/veraison/installation.json` - Alternative installation location
4. `./installation.json` - Local/development installations

## Generating Metadata for Each Deployment Method

### Debian Package (deb)

The Debian package should generate metadata during postinst:

```bash
#!/bin/bash
# In debian/postinst

METADATA_FILE="/usr/share/veraison/installation.json"
VERSION="$(dpkg-query -W -f='${Version}' veraison-services)"
INSTALL_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

cat > "$METADATA_FILE" <<EOF
{
  "version": "${VERSION}",
  "deployment_method": "deb",
  "install_time": "${INSTALL_TIME}",
  "attestation_path": "/usr/share/doc/veraison/attestation.intoto.jsonl",
  "metadata": {
    "package": "veraison-services",
    "architecture": "$(dpkg --print-architecture)"
  }
}
EOF

chmod 644 "$METADATA_FILE"
```

### RPM Package

The RPM spec file should generate metadata during %post:

```spec
%post
METADATA_FILE="/usr/share/veraison/installation.json"
VERSION="%{version}"
INSTALL_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

cat > "$METADATA_FILE" <<EOF
{
  "version": "${VERSION}",
  "deployment_method": "rpm",
  "install_time": "${INSTALL_TIME}",
  "attestation_path": "/usr/share/doc/veraison/attestation.intoto.jsonl",
  "metadata": {
    "package": "veraison-services",
    "architecture": "%{_arch}"
  }
}
EOF

chmod 644 "$METADATA_FILE"
```

### Docker Container

The Dockerfile should generate metadata during build:

```dockerfile
ARG VERSION=latest
ARG BUILD_TIME
ARG ATTESTATION_DIGEST

RUN echo "{\
  \"version\": \"${VERSION}\",\
  \"deployment_method\": \"docker\",\
  \"install_time\": \"${BUILD_TIME}\",\
  \"attestation_path\": \"/opt/veraison/attestation.intoto.jsonl\",\
  \"artifact_digest\": \"${ATTESTATION_DIGEST}\",\
  \"metadata\": {\
    \"image\": \"veraison/services:${VERSION}\"\
  }\
}" > /opt/veraison/installation.json
```

### Native Installation

The installation script should generate metadata:

```bash
#!/bin/bash
# In install.sh

METADATA_FILE="/opt/veraison/installation.json"
VERSION="${VERAISON_VERSION:-unknown}"
INSTALL_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

mkdir -p "$(dirname "$METADATA_FILE")"

cat > "$METADATA_FILE" <<EOF
{
  "version": "${VERSION}",
  "deployment_method": "native",
  "install_time": "${INSTALL_TIME}",
  "metadata": {
    "install_script": "install.sh",
    "prefix": "${INSTALL_PREFIX}"
  }
}
EOF

chmod 644 "$METADATA_FILE"
```

### Source Installation (Development)

For development builds from source:

```bash
#!/bin/bash
# During build or first run

METADATA_FILE="./installation.json"
VERSION=$(git describe --tags --always)
INSTALL_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

cat > "$METADATA_FILE" <<EOF
{
  "version": "${VERSION}",
  "deployment_method": "source",
  "install_time": "${INSTALL_TIME}",
  "metadata": {
    "git_commit": "$(git rev-parse HEAD)",
    "build_host": "$(hostname)"
  }
}
EOF
```

## Adding Artifact Attestations

When generating attestations (e.g., using SLSA or in-toto), update the metadata with the attestation details:

```bash
# After generating attestation
ATTESTATION_FILE="/usr/share/veraison/attestation.intoto.jsonl"
ATTESTATION_DIGEST=$(sha256sum "$ATTESTATION_FILE" | awk '{print "sha256:"$1}')

# Update metadata to include attestation info
jq --arg path "$ATTESTATION_FILE" \
   --arg digest "$ATTESTATION_DIGEST" \
   '.attestation_path = $path | .artifact_digest = $digest' \
   installation.json > installation.json.tmp && \
   mv installation.json.tmp installation.json
```

## Custom Deployment Methods

For custom or third-party deployment methods, follow the same pattern:

1. Choose a descriptive `deployment_method` name
2. Generate the JSON metadata file
3. Place it in one of the standard locations (or a custom location if using `GetInstallationInfoFromPaths()`)
4. Include any deployment-specific information in the `metadata` field

Example for Flatpak:

```json
{
  "version": "1.0.0",
  "deployment_method": "flatpak",
  "install_time": "2025-10-02T14:30:00Z",
  "metadata": {
    "flatpak_id": "io.veraison.Services",
    "runtime": "org.freedesktop.Platform"
  }
}
```

## Testing

To test metadata generation, use the `WriteInstallationMetadata` function:

```go
info := &api.InstallationInfo{
    Version:          "1.0.0",
    DeploymentMethod: "test",
    InstallTime:      time.Now().UTC().Format(time.RFC3339),
}

err := api.WriteInstallationMetadata(info, "/tmp/installation.json")
```

## API Usage

The installation information is included in API responses automatically when available. If no metadata file is found, the installation info will be nil and no error is returned.

```go
info, err := api.GetInstallationInfo()
if err != nil {
    // Handle error reading metadata
}
if info != nil {
    // Use installation information
    log.Printf("Running version %s deployed via %s", info.Version, info.DeploymentMethod)
}
```
