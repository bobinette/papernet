package services

import (
	"encoding/base64"
	"math/rand"
)

func randToken(size int) string {
	b := make([]byte, size)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}
