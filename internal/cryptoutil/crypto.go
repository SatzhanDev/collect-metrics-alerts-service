// Package cryptoutil предоставляет утилиты для асимметричного шифрования
// сообщений между агентом и сервером (RSA-OAEP).
package cryptoutil

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
)

// LoadPublicKey читает PEM-файл по пути path и возвращает RSA публичный ключ.
// Используется агентом для шифрования отправляемых на сервер сообщений.
func LoadPublicKey(path string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read public key file: %w", err)
	}
	return ParsePublicKey(data)
}

// ParsePublicKey разбирает PEM-encoded публичный ключ в формате
// PKIX (SubjectPublicKeyInfo) или PKCS1.
func ParsePublicKey(data []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("cryptoutil: failed to decode PEM block with public key")
	}

	if pub, err := x509.ParsePKIXPublicKey(block.Bytes); err == nil {
		rsaPub, ok := pub.(*rsa.PublicKey)
		if !ok {
			return nil, errors.New("cryptoutil: public key is not an RSA key")
		}
		return rsaPub, nil
	}

	return x509.ParsePKCS1PublicKey(block.Bytes)
}

// LoadPrivateKey читает PEM-файл по пути path и возвращает RSA приватный ключ.
// Используется сервером для дешифрования сообщений, полученных от агента.
func LoadPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read private key file: %w", err)
	}
	return ParsePrivateKey(data)
}

// ParsePrivateKey разбирает PEM-encoded приватный ключ в формате PKCS1 или PKCS8.
func ParsePrivateKey(data []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("cryptoutil: failed to decode PEM block with private key")
	}

	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("cryptoutil: parse private key: %w", err)
	}

	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("cryptoutil: private key is not an RSA key")
	}

	return rsaKey, nil
}

// Encrypt шифрует plaintext публичным ключом pub с помощью RSA-OAEP (SHA-256).
// RSA способен зашифровать за один вызов ограниченный объём данных,
// поэтому plaintext разбивается на блоки размером pub.Size()-2*hashSize-2 байт,
// каждый блок шифруется отдельно, а зашифрованные блоки склеиваются.
func Encrypt(pub *rsa.PublicKey, plaintext []byte) ([]byte, error) {
	hash := sha256.New()
	chunkSize := pub.Size() - 2*hash.Size() - 2
	if chunkSize <= 0 {
		return nil, errors.New("cryptoutil: public key is too small for OAEP encryption")
	}

	result := make([]byte, 0, len(plaintext)+pub.Size())

	for start := 0; start < len(plaintext); start += chunkSize {
		end := start + chunkSize
		if end > len(plaintext) {
			end = len(plaintext)
		}

		encrypted, err := rsa.EncryptOAEP(hash, rand.Reader, pub, plaintext[start:end], nil)
		if err != nil {
			return nil, fmt.Errorf("cryptoutil: encrypt chunk: %w", err)
		}

		result = append(result, encrypted...)
	}

	return result, nil
}

// Decrypt дешифрует ciphertext, зашифрованный функцией Encrypt, приватным
// ключом priv. ciphertext должен состоять из блоков размером priv.Size() байт.
func Decrypt(priv *rsa.PrivateKey, ciphertext []byte) ([]byte, error) {
	hash := sha256.New()
	chunkSize := priv.Size()
	if chunkSize == 0 || len(ciphertext)%chunkSize != 0 {
		return nil, errors.New("cryptoutil: invalid ciphertext length")
	}

	result := make([]byte, 0, len(ciphertext))

	for start := 0; start < len(ciphertext); start += chunkSize {
		end := start + chunkSize

		decrypted, err := rsa.DecryptOAEP(hash, rand.Reader, priv, ciphertext[start:end], nil)
		if err != nil {
			return nil, fmt.Errorf("cryptoutil: decrypt chunk: %w", err)
		}

		result = append(result, decrypted...)
	}

	return result, nil
}
