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

func TestHandler_GetInfo(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTransactionService := mocks.NewMockTransaction(ctrl)
	mockService := &service.Service{Transaction: mockTransactionService}
	mockLog := logrus.New()
	handler := &Handler{services: mockService, log: mockLog}

	tests := []struct {
		name         string
		userID       int64
		mockBehavior func()
		wantCode     int
		wantBody     string
	}{
		{
			name:   "Success",
			userID: 1,
			mockBehavior: func() {
				mockTransactionService.EXPECT().GetUserInfo(gomock.Any(), int64(1)).Return(entity.InfoResponse{
					Coins:     500,
					Inventory: []entity.InventoryItem{{Type: "cup", Quantity: 1}},
					CoinHistory: entity.CoinHistory{
						Received: []entity.TransactionDetail{},
						Sent:     []entity.TransactionDetail{},
					},
				}, nil)
			},
			wantCode: http.StatusOK,
			wantBody: `{"coins":500,"inventory":[{"type":"cup","quantity":1}],"coinHistory":{"received":[],"sent":[]}}`,
		},
		{
			name:   "Error fetching user info",
			userID: 2,
			mockBehavior: func() {
				mockTransactionService.EXPECT().GetUserInfo(gomock.Any(), int64(2)).Return(entity.InfoResponse{}, errors.New("db error"))
			},
			wantCode: http.StatusInternalServerError,
			wantBody: `{"errors":"db error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior()

			req := httptest.NewRequest(http.MethodGet, "/api/info", nil)
			w := httptest.NewRecorder()

			c, _ := gin.CreateTestContext(w)
			c.Request = req

			if tt.userID > 0 {
				c.Set(userCtx, tt.userID)
			}

			handler.getInfo(c)

			assert.Equal(t, tt.wantCode, w.Code)
			assert.JSONEq(t, tt.wantBody, w.Body.String())
		})
	}
}
