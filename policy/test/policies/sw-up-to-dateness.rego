package policy

# Use the psa_executables rules iff the attestaion format is PSA_IOT, and
# to enacttrust_executables iff the format is TPM_ENACTTRUST, otherwise,
# executables will remain undefined.
executables = psa_executables { format == "PSA_IOT" }
             else = enacttrust_executables { format == "TPM_ENACTTRUST" }

# This sets executables trust verctor value to AFFIRMING iff BL version is
# 3.5 or greater, and to failure otherwise.
psa_executables = EXE_AFFIRMING {
  # there exisists some i such that...
  some i
  # ...the i'th software component has type "BL", and...
  evidence["psa-software-components"][i]["measurement-type"] == "BL"

  # ...the version of this component is greater or equal to 3.5.
  # (semver_cmp is defined by the policy package. It returns 1 if the first
  # parameter is greater than the second, -1 if it is less than the second,
  # and 0 if they are equal.)
  semver_cmp(evidence["psa-software-components"][i].version, "3.5") >= 0
} else =  EXE_UNSAFE # unless the above condition is met, return EXE_UNRECOGNIZED

# Unlike the PSA token, the EnactTrust token does not include information about
# multiple sofware componets and instead has a single "firmware" entry.
enacttrust_executables = EXE_AFFIRMING {
  evidence["firmware"] >= 8
} else = EXE_UNSAFE
