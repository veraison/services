package policy

# Use the parsec_cca_executables rule iff the attestaion scheme is PARSEC_CCA,
# otherwise, executables will remain undefined.
executables = parsec_cca_executables { scheme == "PARSEC_CCA" }             

# This sets executables trust vector value to APPROVED_RT iff Realm Initial Measurement
# matches the given value AND 
parsec_cca_executables = APPROVED_RT {
  evidence["cca.realm"]["cca-realm-initial-measurement"] == "X5A2VVSw+obQdbbWpOpYZtsXk9S06ZO7UuVk1yefEXg="

  # ...the Realm personalization value matches the given value.
  evidence["cca.realm"]["cca-realm-personalization-value"] == "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=="
} else =  UNSAFE_RT # unless the above condition is met, return UNSAFE_RT