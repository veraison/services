all: akpub.pem akpub.b64

akpub.pem: akpub.bin
	@openssl ec -pubin -in $< -inform der -out $@

akpub.b64: akpub.bin
	@openssl base64 -in $< -out -A -out $@

clean: ; $(RM) akpub.pem akpub.b64
