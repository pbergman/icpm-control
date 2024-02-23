package main

import (
	"crypto/ed25519"

	"github.com/pbergman/icmp-control/model"
)

func getPublicKey(request *model.Request, config *Config) ed25519.PublicKey {

	for _, key := range config.Clients {
		if key.Id == request.KeyId {
			return key.Key
		}
	}

	return nil
}
