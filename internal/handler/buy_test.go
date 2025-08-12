package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/senyabanana/shop-service/internal/entity"
	"github.com/senyabanana/shop-service/internal/service"
	mocks "github.com/senyabanana/shop-service/internal/service/mocks"
)

func TestHandler_BuyItem(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockInventoryService := mocks.NewMockInventory(ctrl)
	mockLog := logrus.New()
	handler := &Handler{services: &service.Service{Inventory: mockInventoryService}, log: mockLog}

	tests := []struct {
		name         string
		userID       int64
		itemParam    string
		mockBehavior func()
		wantStatus   int
		wantBody     string
	}{
		{
			name:      "Success",
			userID:    1,
			itemParam: "cup",
			mockBehavior: func() {
				mockInventoryService.EXPECT().
					BuyItem(gomock.Any(), int64(1), "cup").
					Return(nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   `{"status":"item was successfully purchased"}`,
		},
		{
			name:         "Item parameter missing",
			userID:       1,
			itemParam:    "",
			mockBehavior: func() {},
			wantStatus:   http.StatusBadRequest,
			wantBody:     `{"errors":"item param is required"}`,
		},
		{
			name:      "Item not found",
			userID:    1,
			itemParam: "UnknownItem",
			mockBehavior: func() {
				mockInventoryService.EXPECT().
					BuyItem(gomock.Any(), int64(1), "UnknownItem").
					Return(entity.ErrItemNotFound)
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"errors":"item not found"}`,
		},
		{
			name:      "Insufficient balance",
			userID:    1,
			itemParam: "cup",
			mockBehavior: func() {
				mockInventoryService.EXPECT().
					BuyItem(gomock.Any(), int64(1), "cup").
					Return(entity.ErrInsufficientBalance)
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"errors":"insufficient balance"}`,
		},
		{
			name:      "Transaction failure",
			userID:    1,
			itemParam: "cup",
			mockBehavior: func() {
				mockInventoryService.EXPECT().
					BuyItem(gomock.Any(), int64(1), "cup").
					Return(errors.New("internal server error"))
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"errors":"internal server error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior()

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			if tt.userID != int64(0) {
				c.Set(userCtx, tt.userID)
			}

			c.Request = httptest.NewRequest(http.MethodPost, "/buy/"+tt.itemParam, nil)
			c.Params = append(c.Params, gin.Param{Key: "item", Value: tt.itemParam})

			handler.buyItem(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			assert.JSONEq(t, tt.wantBody, w.Body.String())
		})
	}
}
