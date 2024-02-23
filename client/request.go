package main

import (
	"crypto/ed25519"
	"crypto/md5"

	"github.com/pbergman/icmp-control/model"
	"golang.org/x/net/icmp"
)

func createRequest(public ed25519.PublicKey, private ed25519.PrivateKey, code uint) ([]byte, error) {

	var message = icmp.Message{
		Type: model.ICMPTypeRequest,
		Code: int(code),
		Body: &model.Request{
			KeyId: md5.Sum(public),
			ID:    1,
			Seq:   1,
		},
	}

	body, err := message.Marshal(nil)

	if err != nil {
		return nil, err
	}

	copy(body[38:], ed25519.Sign(private, body[:38]))

	return body, nil
}
