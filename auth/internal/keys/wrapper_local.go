package keys

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
)

type LocalWrapper struct {
	kek []byte
	ref string
}

func NewLocalWrapperFromEnv() (*LocalWrapper, error) {
	hexKey := os.Getenv("KEK_HEX")
	if len(hexKey) == 0 {
		return nil, errors.New("KEK_HEX not set")
	}
	raw, err := hex.DecodeString(hexKey)
	if err != nil || (len(raw) != 16 && len(raw) != 24 && len(raw) != 32) {
		return nil, errors.New("KEK_HEX must be 16/24/32 bytes hex")
	}
	return &LocalWrapper{kek: raw, ref: "env://KEK_HEX"}, nil
}

func (w *LocalWrapper) Wrap(ctx context.Context, dek []byte) ([]byte, string, error) {
	block, err := aes.NewCipher(w.kek)
	if err != nil {
		return nil, "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, "", err
	}
	ct := gcm.Seal(nil, nonce, dek, []byte("DEK"))
	wrapped := append(nonce, ct...)
	return wrapped, w.ref, nil
}

func (w *LocalWrapper) Unwrap(ctx context.Context, wrapped []byte, kekRef string) ([]byte, error) {
	block, err := aes.NewCipher(w.kek)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(wrapped) < gcm.NonceSize() {
		return nil, errors.New("wrapped too short")
	}
	nonce := wrapped[:gcm.NonceSize()]
	ct := wrapped[gcm.NonceSize():]
	return gcm.Open(nil, nonce, ct, []byte("DEK"))
}
