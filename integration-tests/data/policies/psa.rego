package policy

executables = APPROVED_RT {
  some i

  evidence["psa-software-components"][i]["measurement-type"] == "BL"

  semver_cmp(evidence["psa-software-components"][i].version, "3.5") >= 0
} else =  UNSAFE_RT
