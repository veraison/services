package policy

x := 7

executables = APPROVED_RT {
  x > y  # y undeclared
} else = UNSAFE_RT
