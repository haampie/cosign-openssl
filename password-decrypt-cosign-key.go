package main

import (
    "encoding/base64"
    "encoding/json"
    "bytes"
	"fmt"
    "os"
    "crypto/x509"
    "crypto/ecdsa"
	"golang.org/x/crypto/nacl/secretbox"
    "golang.org/x/crypto/scrypt"
)

type Params struct {
    N int `json:"N"`
    R int `json:"r"`
    P int `json:"p"`
}

type Kdf struct {
    Name string `json:"name"`
    Params Params `json:"params"`
    Salt string `json:"salt"`
}

type Cipher struct {
    Name string
    Nonce string
}

type Key struct {
    Kdf Kdf `json:"kdf"`
    Cipher Cipher `json:"cipher"`
    Ciphertext string `json:"ciphertext"`
}

func main() {
    if len(os.Args) < 3 {
        fmt.Printf("Usage: %s <cosign.key> <password>", os.Args[0])
        os.Exit(1)
    }

    // Read the file in the first positional argument
    contents, err := os.ReadFile(os.Args[1])

    // Read the password from the command line
    password := os.Args[2]

    if err != nil { panic(err) }
    
    // Split lines and drop first/last line
    lines := bytes.Split(contents, []byte("\n"))
    lines = lines[1:len(lines)-2]

    // Join back to bytes
    data := bytes.Join(lines, []byte(""))

    // Base64 decode this
    decoded, err := base64.StdEncoding.DecodeString(string(data))
    if err != nil { panic(err) }

    var key Key;

    // Then parse it as JSON
    json.Unmarshal(decoded, &key)

    salt, err := base64.StdEncoding.DecodeString(key.Kdf.Salt)

    if err != nil { panic(err) }

    nonce, err := base64.StdEncoding.DecodeString(key.Cipher.Nonce)

    if err != nil { panic(err) }

    dk, err := scrypt.Key([]byte(password), salt, key.Kdf.Params.N, key.Kdf.Params.R, key.Kdf.Params.P, 32)

    if err != nil { panic(err) }

    ciphertext, err := base64.StdEncoding.DecodeString(key.Ciphertext)

    // Create fixed length byte buffers?
    var keyBytes [32]byte;
    var nonceBytes [24]byte;

    copy(keyBytes[:], dk)
    copy(nonceBytes[:], nonce)

    x509Encoded, ok := secretbox.Open(nil, ciphertext, &nonceBytes, &keyBytes) 

    if !ok { panic(err) }

    pk, err := x509.ParsePKCS8PrivateKey(x509Encoded)

	if err != nil { panic(err) }

    _, ok = pk.(*ecdsa.PrivateKey)

    if !ok { panic("oh no") }

    fmt.Println("-----BEGIN PRIVATE KEY-----")
    fmt.Println(base64.StdEncoding.EncodeToString(x509Encoded))
    fmt.Println("-----END PRIVATE KEY-----")

}

