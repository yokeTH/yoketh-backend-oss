CREATE TABLE IF NOT EXISTS jwk_keys (
  kid TEXT PRIMARY KEY,
  alg TEXT NOT NULL,
  public_jwk JSONB NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('ACTIVE', 'RETIRING', 'RETIRED')),
  priv_ciphertext BYTEA NOT NULL, -- AES-GCM( DEK, pkcs8, aad="PRIV:"+kid )
  priv_nonce BYTEA NOT NULL,
  wrapped_dek BYTEA NOT NULL, -- AES-GCM( KEK, dek, aad="DEK")
  kek_ref TEXT, -- where KEK comes from (e.g., env://KEK_HEX)
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  rotated_at TIMESTAMPTZ,
  not_before TIMESTAMPTZ,
  not_after TIMESTAMPTZ
);
