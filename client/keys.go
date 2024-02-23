package main

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
)

func readKeys(path string) (ed25519.PublicKey, ed25519.PrivateKey, error) {

	var public = make([]byte, ed25519.PublicKeySize)
	var private = make([]byte, ed25519.PrivateKeySize)

	pub, err := os.ReadFile(filepath.Join(path, "public.key"))

	if err != nil {
		return nil, nil, fmt.Errorf("could not read public key file " + filepath.Join(path, "public.key"))
	}

	if _, err := base64.StdEncoding.Decode(public, pub); err != nil {
		return nil, nil, fmt.Errorf("could not load public key")
	}

	pri, err := os.ReadFile(filepath.Join(path, "private.key"))

	if err != nil {
		return nil, nil, fmt.Errorf("could not read private key file " + filepath.Join(path, "private.key"))
	}

	if _, err := base64.StdEncoding.Decode(private, pri); err != nil {
		return nil, nil, fmt.Errorf("could not load private key")
	}

	return ed25519.PublicKey(public), ed25519.PrivateKey(private), nil
}
