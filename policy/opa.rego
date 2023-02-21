package policy

import future.keywords

NO_CLAIM := 0
UNEXECTED_EVIDENCE := 1

AFFIRMING := 2
WARNING := 32
CONTRAINDICATED := 96

RECOGNIZED_INSTANCE := 2
UNTRUSTWORTHY_INSTANCE := 96
UNRECOGNIZED_INSTANCE := 97

APPROVED_CONFIG := 2
SAFE_CONFIG := 3
UNSAFE_CONFIG := 32
UNSUPPORTABLE_CONFIG := 96

APPROVED_RT := 2
APPROVED_BOOT := 3
UNSAFE_RT := 32
UNRECOGNIZED_RT := 33
CONTRAINDICATED_RT := 96

APPROVED_FS := 2
UNRECOGNIZED_FS := 32
CONTRAINDICATED_FS := 96

GENUINE_HW := 2
UNSAFE_HW := 32
CONTRAINDICATED_HW := 96
UNRECOGNIZED_HW := 97

ENCRYPTED_RT := 2
ISOLATED_RT := 32
VISIBLE_RT := 96

HW_ENCRYPTED_SECRETS := 2
SW_ENCRYPTED_SECRETS := 32
UNENCRYPTED_SECRETES := 96

TRUSTED_SOURCES := 2
UNTRUSTED_SOURCES := 32
CONTRAINDICATED_SOURCES := 96

submod := input.submod
result := input.result
evidence := input.evidence["evidence"]
endorsements := input.endorsements

format_names := [
  "UNKNOWN_FORMAT",
  "PSA_IOT",
  "TCG_DICE",
  "TPM_ENACTTRUST",
]

format := format_names[input.evidence["attestation-format"]]

default status = 0
default instance_identity = 0
default configuration = 0
default executables = 0
default file_system = 0
default hardware = 0
default runtime_opaque = 0
default storage_opaque = 0
default sourced_data = 0
default added_claims = {}

simple_semver_split (s) := res {
  parts :=  split(s, ".")
  res := [to_number(p) | p := parts[_]]
}

cmp (n1, n2) := 1 {n1 > n2} else =  -1 { n1 < n2 } else = 0

# TODO(setrofim): this currently only supports simple SemVer's in the form
# MAJOR.MINOR.PATCH (e.g. "3.1.7"). Pre-release versions (e.g. "3.1.7-alpha")
# are not supported.
semver_cmp (s1, s2) := res {
  n1 := simple_semver_split(s1)
  n2 := simple_semver_split(s2)
  res := cmp(n1, n2)
}

outcome := {
  "status": status,
  "trust-vector": {
    "instance-identity": instance_identity,
    "configuration":     configuration,
    "executables":       executables,
    "file-system":       file_system,
    "hardware":          hardware,
    "runtime-opaque":    runtime_opaque,
    "storage-opaque":    storage_opaque,
    "sourced-data":      sourced_data
  },
  "added-claims": added_claims,
}
