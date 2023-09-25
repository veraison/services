package policy

status = AFFIRMING {
  base64url.decode(session["nonce"]) == base64.decode(evidence["realm"]["cca-realm-challenge"])
} else =  CONTRAINDICATED
