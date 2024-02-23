package model

import (
	"crypto/ed25519"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"time"
)

type Request struct {
	ID        uint16
	Seq       uint16
	Time      time.Time
	KeyId     [md5.Size]byte
	Args      [6]byte
	Signature []byte
}

func (r *Request) Len(proto int) int {
	return 98
}

func (r *Request) Unmarshal(data []byte) error {
	if len(data) != r.Len(0) {
		return fmt.Errorf("invalid payload length (%d)", len(data))
	}
	r.ID = binary.BigEndian.Uint16(data[0:2])
	r.Seq = binary.BigEndian.Uint16(data[2:4])
	r.Time = time.Unix(int64(binary.LittleEndian.Uint64(data[4:12])), 0)
	r.Signature = make([]byte, ed25519.SignatureSize)
	copy(r.KeyId[:], data[12:28])
	copy(r.Args[:], data[28:34])
	copy(r.Signature, data[34:98])
	return nil
}

func (r *Request) Marshal(proto int) ([]byte, error) {

	var payload = make([]byte, r.Len(0))
	var now = r.Time

	if now.IsZero() {
		now = time.Now().UTC()
	}

	binary.BigEndian.PutUint16(payload[0:2], r.ID)
	binary.BigEndian.PutUint16(payload[2:4], r.Seq)

	// 8  bytes - TIMESTAMP
	binary.LittleEndian.PutUint64(payload[4:12], uint64(now.Unix()))

	// 16 bytes - KEY ID
	copy(payload[12:28], r.KeyId[:])
	// 6 bytes  - Args
	copy(payload[28:34], r.Args[:])

	// 64 bytes - Signature
	if len(r.Signature) == ed25519.SignatureSize {
		copy(payload[34:98], r.Signature)
	}

	return payload, nil
}
