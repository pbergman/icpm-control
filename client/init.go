package main

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func initConfigDir(path string) error {

	if _, err := os.Stat(path); os.IsNotExist(err) {

		fmt.Printf("config dir not exist, will create dir \"%s\"\n", path)

		_ = os.MkdirAll(path, 0700)

		public, private, err := ed25519.GenerateKey(nil)

		if err != nil {
			log.Fatal(err)
		}

		if err := os.WriteFile(filepath.Join(path, "public.key"), []byte(base64.StdEncoding.EncodeToString(public)), 0400); err != nil {
			log.Fatalf("failed to save public key: %s", err)
		}

		if err := os.WriteFile(filepath.Join(path, "private.key"), []byte(base64.StdEncoding.EncodeToString(private)), 0400); err != nil {
			log.Fatalf("failed to save private key: %s", err)
		}

		log.Println("generated new keys")
		log.Printf("client public key:  %s\n", base64.StdEncoding.EncodeToString(public))
	}

	return nil
}
