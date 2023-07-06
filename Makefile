COSIGN_URL = https://github.com/sigstore/cosign/releases/download/v2.1.1/cosign-linux-amd64
CURL = curl
GO = go
PASSWORD = secret!
OPENSSL = openssl

.PHONY: all distclean clean verify

all: verify

cosign:
	# Install cosign (yes, I'm not validating their binary)
	$(CURL) -LfsS -o $@ https://github.com/sigstore/cosign/releases/download/v2.1.1/cosign-linux-amd64
	chmod +x $@

cosign.key: cosign
	# Generate a key
	COSIGN_PASSWORD="$(PASSWORD)" ./cosign generate-key-pair

cosign.decrypted.key: cosign.key password-decrypt-cosign-key.go
	# Password decrypt the private key
	go run password-decrypt-cosign-key.go $< "$(PASSWORD)" > $@

openssl.key: cosign.decrypted.key
	# Show the key information
	$(OPENSSL) pkey -in $< -text -noout
	# Password protect the key again with openssl
	$(OPENSSL) pkcs8 -in $< -topk8 -scrypt -passout "pass:$(PASSWORD)" -out $@

openssl.pub: openssl.key
	# Extract the public key from the private one
	$(OPENSSL) pkey -in $< -passin "pass:$(PASSWORD)" -pubout -out $@

example:
	# Create a dummy file
	echo "This is a file" > $@

example.sig: openssl.key example
	# Sign the dummy file
	$(OPENSSL) dgst -sign openssl.key -passin "pass:$(PASSWORD)" -out $@ example

verify: example.sig openssl.pub cosign.pub
	# Verify the dummy file using our extracted public key
	$(OPENSSL) dgst -verify openssl.pub -signature example.sig example
	# Verify the dummy file using the public key that cosign generated (should be the same)
	$(OPENSSL) dgst -verify cosign.pub -signature example.sig example

clean:
	rm -f cosign.key cosign.pub cosign.decrypted.key openssl.key openssl.pub example example.sig

distclean: clean
	rm -f cosign

