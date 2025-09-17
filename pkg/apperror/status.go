package apperror

type ErrorStatus string

var (
	StatusJWKError ErrorStatus = "JWKS_ERROR"

	StatusFiberError ErrorStatus = "FIBER_ERROR"

	StatusInternalServerError ErrorStatus = "INTERNAL_SERVER_ERROR"
)
