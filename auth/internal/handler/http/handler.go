package http

import (
	"github.com/gofiber/fiber/v3"
	"github.com/yokeTH/yoketh-backend-oss/auth/internal/keys"
	"github.com/yokeTH/yoketh-backend-oss/pkg/apperror"
)

type Handler struct {
	Mgr *keys.Manager
}

func NewHandler(mgr *keys.Manager) *Handler {
	return &Handler{
		Mgr: mgr,
	}
}

func (h *Handler) HandleJWKS(ctx fiber.Ctx) error {
	jwks, err := h.Mgr.JWKS(ctx)
	if err != nil {
		return apperror.InternalServerError(err, "get jwks error", apperror.StatusJWKError)
	}

	return ctx.JSON(jwks)
}
