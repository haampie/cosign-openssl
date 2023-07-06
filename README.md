Run `make`, it will do the following:

1. Generate a keypair with `sigstore/cosign`
2. Compile and run some Go code to password decrypt the generated private key
3. Re-password-encrypt it with the `openssl` utility (also using scrypt, but probably a different symmetric encryption strategy than what cosign does...)
4. Extract the public part from the `openssl`-version of the keypair
5. Sign some file with the `openssl`-version of the private key
6. Verify the signature using (a) the original public key that cosign generated, and (b) the extract public key from the `openssl`-version of it.
