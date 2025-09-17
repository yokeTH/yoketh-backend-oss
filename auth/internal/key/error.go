package key

import "errors"

var (
	ErrNotECDSAKey = errors.New("not an ECDSA key")
)
