package identity

import (
	"crypto/rand"
	"encoding/hex"
	"sync/atomic"
	"time"
)

type Generator struct{ sequence atomic.Uint64 }

func (g *Generator) New(prefix string) string {
	bytes := make([]byte, 6)
	if _, err := rand.Read(bytes); err != nil {
		value := g.sequence.Add(1)
		return prefix + "_" + time.Now().UTC().Format("20060102150405") + "_" + encodeUint(value)
	}
	return prefix + "_" + hex.EncodeToString(bytes)
}

func encodeUint(value uint64) string {
	buffer := make([]byte, 8)
	for index := len(buffer) - 1; index >= 0; index-- {
		buffer[index] = byte(value)
		value >>= 8
	}
	return hex.EncodeToString(buffer)
}
