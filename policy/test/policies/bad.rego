package policy

x := 7

executables = EXE_AFFIRMING {
  x > y  # y undeclared
} else = EXE_UNSAFE
