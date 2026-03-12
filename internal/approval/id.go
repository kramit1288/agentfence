package approval

import (
	"crypto/rand"
	"encoding/hex"
)

func newID() string {
	var random [6]byte
	_, _ = rand.Read(random[:])
	return "apr_" + hex.EncodeToString(random[:])
}
