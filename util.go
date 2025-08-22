package main

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"time"
)

func uuid() string {
	b := make([]byte, 16)
	binary.BigEndian.PutUint64(b[:8], uint64(time.Now().Unix()))
	rand.Read(b[8:])
	return hex.EncodeToString(b)[6:]
}
