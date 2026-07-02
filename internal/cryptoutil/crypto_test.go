package cryptoutil

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func generateKeyPair(t *testing.T) (*rsa.PrivateKey, *rsa.PublicKey) {
	t.Helper()

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	return priv, &priv.PublicKey
}

func writePEMFile(t *testing.T, path string, block *pem.Block) {
	t.Helper()

	data := pem.EncodeToMemory(block)
	require.NoError(t, os.WriteFile(path, data, 0o600))
}

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	priv, pub := generateKeyPair(t)

	plaintext := []byte("hello, this is a metrics payload that is definitely longer than one RSA-OAEP chunk of about 190 bytes, so it must be split into multiple chunks and reassembled correctly on decryption")

	ciphertext, err := Encrypt(pub, plaintext)
	require.NoError(t, err)
	require.NotEqual(t, plaintext, ciphertext)

	decrypted, err := Decrypt(priv, ciphertext)
	require.NoError(t, err)
	require.Equal(t, plaintext, decrypted)
}

func TestEncryptDecrypt_Empty(t *testing.T) {
	priv, pub := generateKeyPair(t)

	ciphertext, err := Encrypt(pub, nil)
	require.NoError(t, err)

	decrypted, err := Decrypt(priv, ciphertext)
	require.NoError(t, err)
	require.Empty(t, decrypted)
}

func TestDecrypt_InvalidLength(t *testing.T) {
	priv, _ := generateKeyPair(t)

	_, err := Decrypt(priv, []byte("not a valid multiple of key size"))
	require.Error(t, err)
}

func TestLoadPublicKey_PKIX(t *testing.T) {
	_, pub := generateKeyPair(t)

	der, err := x509.MarshalPKIXPublicKey(pub)
	require.NoError(t, err)

	path := filepath.Join(t.TempDir(), "public.pem")
	writePEMFile(t, path, &pem.Block{Type: "PUBLIC KEY", Bytes: der})

	loaded, err := LoadPublicKey(path)
	require.NoError(t, err)
	require.Equal(t, pub.N, loaded.N)
	require.Equal(t, pub.E, loaded.E)
}

func TestLoadPrivateKey_PKCS1(t *testing.T) {
	priv, _ := generateKeyPair(t)

	der := x509.MarshalPKCS1PrivateKey(priv)

	path := filepath.Join(t.TempDir(), "private.pem")
	writePEMFile(t, path, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: der})

	loaded, err := LoadPrivateKey(path)
	require.NoError(t, err)
	require.Equal(t, priv.D, loaded.D)
}

func TestLoadPrivateKey_PKCS8(t *testing.T) {
	priv, _ := generateKeyPair(t)

	der, err := x509.MarshalPKCS8PrivateKey(priv)
	require.NoError(t, err)

	path := filepath.Join(t.TempDir(), "private.pem")
	writePEMFile(t, path, &pem.Block{Type: "PRIVATE KEY", Bytes: der})

	loaded, err := LoadPrivateKey(path)
	require.NoError(t, err)
	require.Equal(t, priv.D, loaded.D)
}

func TestLoadPublicKey_FileNotFound(t *testing.T) {
	_, err := LoadPublicKey(filepath.Join(t.TempDir(), "missing.pem"))
	require.Error(t, err)
}

func TestLoadPrivateKey_FileNotFound(t *testing.T) {
	_, err := LoadPrivateKey(filepath.Join(t.TempDir(), "missing.pem"))
	require.Error(t, err)
}

func TestParsePublicKey_InvalidPEM(t *testing.T) {
	_, err := ParsePublicKey([]byte("not a pem"))
	require.Error(t, err)
}

func TestParsePrivateKey_InvalidPEM(t *testing.T) {
	_, err := ParsePrivateKey([]byte("not a pem"))
	require.Error(t, err)
}
