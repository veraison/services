package policy

import future.keywords

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

default status = ""
default hw_authenticity = ""
default sw_integrity = ""
default sw_up_to_dateness = ""
default config_integrity = ""
default runtime_integrity = ""
default certification_status = ""

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
    "hw-authenticity": hw_authenticity,
    "sw-integrity": sw_integrity,
    "sw-up-to-dateness": sw_up_to_dateness,
    "config-integrity": config_integrity,
    "runtime-integrity": runtime_integrity,
    "certification-status": certification_status
 }
}
