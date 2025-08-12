package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/senyabanana/shop-service/internal/entity"
)

func (h *Handler) buyItem(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		return
	}

	item := c.Param("item")
	if item == "" {
		entity.NewErrorResponse(c, h.log, http.StatusBadRequest, "item param is required")
		return
	}

	err = h.services.Inventory.BuyItem(c.Request.Context(), userID, item)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrItemNotFound):
			entity.NewErrorResponse(c, h.log, http.StatusBadRequest, "item not found")
		case errors.Is(err, entity.ErrInsufficientBalance):
			entity.NewErrorResponse(c, h.log, http.StatusBadRequest, "insufficient balance")
		default:
			entity.NewErrorResponse(c, h.log, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	c.JSON(http.StatusOK, entity.StatusResponse{
		Status: "item was successfully purchased",
	})
}
