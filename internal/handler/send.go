package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/senyabanana/shop-service/internal/entity"
)

func (h *Handler) sendCoin(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		return
	}

	var input entity.SendCoinRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		entity.NewErrorResponse(c, h.log, http.StatusBadRequest, "invalid request format")
		return
	}

	err = h.services.Transaction.SendCoin(c.Request.Context(), userID, input.ToUser, input.Amount)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrRecipientNotFound):
			entity.NewErrorResponse(c, h.log, http.StatusBadRequest, "recipient not found")
		case errors.Is(err, entity.ErrSendThemselves):
			entity.NewErrorResponse(c, h.log, http.StatusBadRequest, "cannot send coins to yourself")
		case errors.Is(err, entity.ErrInsufficientBalance):
			entity.NewErrorResponse(c, h.log, http.StatusBadRequest, "insufficient balance")
		default:
			entity.NewErrorResponse(c, h.log, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	c.JSON(http.StatusOK, entity.StatusResponse{
		Status: "coins were successfully sent to the user",
	})
}
