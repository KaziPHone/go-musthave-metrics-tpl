package helpers

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"sync"
)

// hasherPool позволяет повторно использовать экземпляр хэшера для избежания выделения памяти
var hasherPool = sync.Pool{
	New: func() interface{} {
		return sha256.New()
	},
}

func CalcSHA256HashBuffer(data bytes.Buffer) string {
	hasher := hasherPool.Get().(interface {
		Reset()
		io.Writer
		Sum([]byte) []byte
	})
	defer hasherPool.Put(hasher)

	hasher.Reset()
	io.Copy(hasher, &data)
	return fmt.Sprintf("%x", hasher.Sum(nil))
}

func CalcSHA256Hash(data []byte) string {
	h := hasherPool.Get().(interface {
		Reset()
		io.Writer
		Sum([]byte) []byte
	})
	defer hasherPool.Put(h)

	h.Reset()
	h.Write(data)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func IsBadShaRequest(data []byte, hash string) bool {
	return CalcSHA256Hash(data) != hash && hash != ""
}
