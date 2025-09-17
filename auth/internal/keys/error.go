package keys

import "errors"

var (
	ErrNotECDSAKey = errors.New("not an ECDSA key")
)
