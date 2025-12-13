package helpers

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
)

func CalcSHA256HashBuffer(data bytes.Buffer) string {
	hasher := sha256.New()
	io.Copy(hasher, &data)
	return fmt.Sprintf("%x", hasher.Sum(nil))
}

func CalcSHA256Hash(data []byte) string {
	h := sha256.New()
	h.Write(data)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func IsBadShaRequest(data []byte, hash string) bool {
	return CalcSHA256Hash(data) != hash
}