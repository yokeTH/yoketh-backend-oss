package key

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Signer struct {
	KID  string
	Priv *ecdsa.PrivateKey
	Iss  string
	Aud  string
	TTL  time.Duration
}

func NewSigner(ctx context.Context, m *Manager, aud, iss string, ttl time.Duration) (*Signer, error) {
	kid, priv, err := m.LoadActiveSigner(ctx)
	if err != nil {
		return nil, err
	}
	return &Signer{KID: kid, Priv: priv, Iss: iss, Aud: aud, TTL: ttl}, nil
}

type CustomClaims struct {
	Scopes      []string `json:"scopes"`
	Permissions []string `json:"permissions,omitempty"`
	jwt.RegisteredClaims
}

func (s *Signer) Sign(sub string, scopes []string) (string, error) {
	now := time.Now()

	// todo: get permissions from polices service
	rc := CustomClaims{
		Scopes: scopes,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.Iss,
			Audience:  jwt.ClaimStrings{s.Aud},
			Subject:   sub,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.TTL)),
			ID:        genJTI(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, rc)
	token.Header["kid"] = s.KID

	return token.SignedString(s.Priv)
}

func genJTI() string {
	return randomBase64URL(16)
}

func randomBase64URL(i int) string {
	if i <= 0 {
		return ""
	}
	n := (i*3 + 3) / 4

	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}

	s := base64.RawURLEncoding.EncodeToString(b)
	if len(s) > i {
		return s[:i]
	}
	return s
}
