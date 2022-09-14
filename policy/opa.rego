package policy

import future.keywords

NO_CLAIM := 0
CANNOT_PARSE := 1
VERIFIER_ERROR := -1
INVALID := 99
CONF_AFFIRMING := 2
CONF_NOVULN := 3
CONF_UNSAFE := 32
CONF_UNSUPPORTABLE := 96
EXE_AFFIRMING := 2
EXE_BOOT_AFFIRMING := 3
EXE_UNSAFE := 32
EXE_UNRECOGNIZED := 33
EXE_CONTRAINDICATED := 96
FS_AFFIRMING := 2
FS_UNSAFE := 32
FS_COUNTRAINDICATED := 96
HW_AFFIRMING := 2
HW_UNSAFE := 32
HW_CONTRAINDICATED := 96
HW_UNRECOGNIZED := 97
IDENT_AFFIRMING := 2
IDENT_CONTRAINDICATED := 96
IDENT_UNRECOGNIZED := 97
RT_AFFIRMING := 2
RT_ISOLATED := 32
RT_EXPOSED := 96
SOURCED_AFFIRMING := 2
SOURCED_UNSAFE := 32
SOURCED_CONTRAINDICATED := 96
SECRETS_AFFIRMING := 2
SECRETS_NOHWKEYS := 32
SECRETS_EXPOSED := 96


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
default verifier_added_claims = {}

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
  "veraison-verifier-added-claims": verifier_added_claims,
}
