-- name: CreateJWK :exec
INSERT INTO
  jwk_keys (
    kid,
    alg,
    public_jwk,
    status,
    priv_ciphertext,
    priv_nonce,
    wrapped_dek,
    kek_ref,
    created_at,
    rotated_at
  )
VALUES
  ($1, $2, $3, 'ACTIVE', $4, $5, $6, $7, now(), NULL);

-- name: GetJWK :one
SELECT
  kid,
  alg,
  public_jwk,
  status,
  priv_ciphertext,
  priv_nonce,
  wrapped_dek,
  kek_ref,
  created_at,
  rotated_at
FROM
  jwk_keys
WHERE
  status IN ('ACTIVE', 'RETIRING')
ORDER BY
  CASE status
    WHEN 'ACTIVE' THEN 0
    ELSE 1
  END,
  created_at DESC
LIMIT
  1;

-- name: UpdateJWKToRetiring :exec
UPDATE jwk_keys
SET
  status = 'RETIRING',
  rotated_at = now()
WHERE
  status = 'ACTIVE';

-- name: UpdateJWKToRetired :exec
UPDATE jwk_keys
SET
  status = 'RETIRED'
WHERE
  status = 'RETIRING';

-- name: GetPubJWK :many
SELECT
  kid,
  public_jwk
FROM
  jwk_keys
WHERE
  status IN ('ACTIVE', 'RETIRING')
ORDER BY
  created_at DESC;

-- name: CountJWK :one
SELECT
  COUNT(*)
FROM
  jwk_keys
WHERE
  status IN ('ACTIVE', 'RETIRING');
