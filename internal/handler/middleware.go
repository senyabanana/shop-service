package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/senyabanana/shop-service/internal/entity"
)

const (
	authorizationHeader = "Authorization"
	userCtx             = "userID"
)

func (h *Handler) userIdentity(c *gin.Context) {
	header := c.GetHeader(authorizationHeader)
	if header == "" {
		entity.NewErrorResponse(c, h.log, http.StatusUnauthorized, "empty auth header")
		return
	}

	headerParts := strings.Split(header, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		entity.NewErrorResponse(c, h.log, http.StatusUnauthorized, "invalid auth header format")
		return
	}

	userID, err := h.services.Authorization.ParseToken(headerParts[1])
	if err != nil {
		entity.NewErrorResponse(c, h.log, http.StatusUnauthorized, "invalid or expired token")
		return
	}

	c.Set(userCtx, userID)
}

func (h *Handler) getUserID(c *gin.Context) (int64, error) {
	id, ok := c.Get(userCtx)
	if !ok {
		h.log.Warn("getUserID: user id not found in context")
		return 0, entity.ErrUserNotFound
	}

	idInt, ok := id.(int64)
	if !ok {
		h.log.Warn("getUserID: user id is of invalid type")
		return 0, entity.ErrInvalidUserIDType
	}

	return idInt, nil
}
