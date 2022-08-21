package policy

x := 7

sw_up_to_dateness = "SUCCESS" {
  x > y  # y undeclared
} else = "FAILURE"
