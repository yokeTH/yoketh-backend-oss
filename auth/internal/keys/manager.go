package keys

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"database/sql"
	"encoding/base64"
	"encoding/json"

	"github.com/yokeTH/yoketh-backend-oss/auth/internal/db"
)

type ecPublicJWK struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	Crv string `json:"crv"`
	Alg string `json:"alg"`
	Kid string `json:"kid"`
	X   string `json:"x"`
	Y   string `json:"y"`
}

type jwks struct {
	Keys []ecPublicJWK `json:"keys"`
}

type KeyWrapper interface {
	Wrap(ctx context.Context, dek []byte) (wrapped []byte, kekRef string, err error)
	Unwrap(ctx context.Context, wrapped []byte, kekRef string) (dek []byte, err error)
}

type Manager struct {
	DB      *sql.DB
	queries *db.Queries
	Wrapper KeyWrapper
	Issuer  string
}

func NewManager(db *sql.DB, q *db.Queries, wrapper KeyWrapper, iss string) *Manager {
	return &Manager{
		DB:      db,
		queries: q,
		Wrapper: wrapper,
		Issuer:  iss,
	}
}

func (m *Manager) Init(ctx context.Context) error {
	n, err := m.queries.CountJWK(ctx)
	if err != nil {
		return err
	}

	if n == 0 {
		_, err := m.GenerateAndStore(context.Background(), m.genKID())
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Manager) GenerateAndStore(ctx context.Context, kid string) (string, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", err
	}
	pub := priv.Public().(*ecdsa.PublicKey)

	pkcs8, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return "", err
	}

	dek := make([]byte, 32)
	if _, err := rand.Read(dek); err != nil {
		return "", err
	}

	privNonce, privCT, err := aesGCMEncrypt(dek, pkcs8, []byte("PRIV:"+kid))
	if err != nil {
		return "", err
	}

	wrapped, kekRef, err := m.Wrapper.Wrap(ctx, dek)
	if err != nil {
		return "", err
	}

	pubJWK, err := buildECPublicJWK(kid, pub)
	if err != nil {
		return "", err
	}
	pubJWKJSON, _ := json.Marshal(pubJWK)

	m.queries.CreateJWK(ctx, db.CreateJWKParams{
		KID:            kid,
		ALG:            "ES256",
		PublicJWK:      pubJWKJSON,
		PrivCiphertext: privCT,
		PrivNonce:      privNonce,
		WrappedDEK:     wrapped,
		KEKRef: sql.NullString{
			String: kekRef,
			Valid:  kekRef != "",
		},
	})

	return kid, err
}

func (m *Manager) LoadActiveSigner(ctx context.Context) (kid string, priv *ecdsa.PrivateKey, err error) {
	r, err := m.queries.GetJWK(ctx)
	if err != nil {
		return "", nil, err
	}

	dek, err := m.Wrapper.Unwrap(ctx, r.WrappedDEK, r.KEKRef.String)
	if err != nil {
		return "", nil, err
	}

	pkcs8, err := aesGCMDecrypt(dek, r.PrivNonce, r.PrivCiphertext, []byte("PRIV:"+r.KID))
	if err != nil {
		return "", nil, err
	}
	key, err := x509.ParsePKCS8PrivateKey(pkcs8)
	if err != nil {
		return "", nil, err
	}
	ecdsaKey, ok := key.(*ecdsa.PrivateKey)
	if !ok {
		return "", nil, ErrNotECDSAKey
	}
	return r.KID, ecdsaKey, nil
}

func (m *Manager) Rotate(ctx context.Context, newKID string) (string, error) {
	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback() }()

	qtx := m.queries.WithTx(tx)
	err = qtx.UpdateJWKToRetiring(ctx)
	err = qtx.UpdateJWKToRetired(ctx)

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", err
	}
	pkcs8, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return "", err
	}

	dek := make([]byte, 32)
	_, _ = rand.Read(dek)
	privNonce, privCT, err := aesGCMEncrypt(dek, pkcs8, []byte("PRIV:"+newKID))
	if err != nil {
		return "", err
	}

	wrapped, kekRef, err := m.Wrapper.Wrap(ctx, dek)
	if err != nil {
		return "", err
	}

	pubJWK, _ := buildECPublicJWK(newKID, &priv.PublicKey)
	pubJWKJSON, _ := json.Marshal(pubJWK)

	err = qtx.CreateJWK(ctx, db.CreateJWKParams{
		KID:            newKID,
		ALG:            "ES256",
		PublicJWK:      pubJWKJSON,
		PrivCiphertext: privCT,
		PrivNonce:      privNonce,
		WrappedDEK:     wrapped,
		KEKRef: sql.NullString{
			String: kekRef,
			Valid:  kekRef != "",
		},
	})
	if err != nil {
		return "", err
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}
	return newKID, nil
}

func (m *Manager) JWKS(ctx context.Context) ([]byte, error) {
	rawPub, err := m.queries.GetPubJWK(ctx)
	if err != nil {
		return nil, err
	}

	keys := make([]ecPublicJWK, len(rawPub))
	for i, p := range rawPub {
		if err := json.Unmarshal(p.PublicJWK, &keys[i]); err != nil {
			return nil, err
		}
	}

	return json.Marshal(jwks{Keys: keys})
}

func buildECPublicJWK(kid string, pub *ecdsa.PublicKey) (ecPublicJWK, error) {
	x := pub.X.Bytes()
	y := pub.Y.Bytes()
	xb := leftPad(x, 32)
	yb := leftPad(y, 32)
	return ecPublicJWK{
		Kty: "EC",
		Use: "sig",
		Crv: "P-256",
		Alg: "ES256",
		Kid: kid,
		X:   b64u(xb),
		Y:   b64u(yb),
	}, nil
}

func leftPad(b []byte, n int) []byte {
	if len(b) >= n {
		return b
	}
	out := make([]byte, n)
	copy(out[n-len(b):], b)
	return out
}

func b64u(b []byte) string { return base64.RawURLEncoding.EncodeToString(b) }

func aesGCMEncrypt(key, plaintext, aad []byte) (nonce, ct []byte, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}
	nonce = make([]byte, gcm.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return nil, nil, err
	}
	ct = gcm.Seal(nil, nonce, plaintext, aad)
	return nonce, ct, nil
}

func aesGCMDecrypt(key, nonce, ct, aad []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return gcm.Open(nil, nonce, ct, aad)
}

func (m *Manager) genKID() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}
