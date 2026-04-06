package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"os"
)

// LoadPublicKeyFromFile загружает RSA публичный ключ из PEM файла.
func LoadPublicKeyFromFile(path string) (*rsa.PublicKey, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read public key file: %w", err)
	}
	block, _ := pem.Decode(b)
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block from public key file")
	}
	pubIfc, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		// try parsing as x509 cert
		cert, err2 := x509.ParseCertificate(block.Bytes)
		if err2 == nil {
			if pk, ok := cert.PublicKey.(*rsa.PublicKey); ok {
				return pk, nil
			}
		}
		return nil, fmt.Errorf("parse public key: %w", err)
	}
	pub, ok := pubIfc.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("public key is not RSA")
	}
	return pub, nil
}

// LoadPrivateKeyFromFile загружает RSA приватный ключ из PEM файла.
func LoadPrivateKeyFromFile(path string) (*rsa.PrivateKey, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read private key file: %w", err)
	}
	block, _ := pem.Decode(b)
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block from private key file")
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err == nil {
		return priv, nil
	}
	pkIfc, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err == nil {
		if p, ok := pkIfc.(*rsa.PrivateKey); ok {
			return p, nil
		}
	}
	return nil, fmt.Errorf("parse private key: %v", err)
}

// Encrypt шифрует данные с помощью RSA
func Encrypt(pub *rsa.PublicKey, data []byte) ([]byte, error) {
	if pub == nil {
		return nil, fmt.Errorf("public key is nil")
	}
	ciphertext, err := rsa.EncryptPKCS1v15(rand.Reader, pub, data)
	if err != nil {
		return nil, fmt.Errorf("encrypt: %w", err)
	}
	return ciphertext, nil
}

// Decrypt расшифровывает данные с помощью RSA приватного ключа.
func Decrypt(priv *rsa.PrivateKey, data []byte) ([]byte, error) {
	if priv == nil {
		return nil, fmt.Errorf("private key is nil")
	}
	plain, err := rsa.DecryptPKCS1v15(rand.Reader, priv, data)
	if err != nil {
		return nil, fmt.Errorf("decrypt: %w", err)
	}
	return plain, nil
}
