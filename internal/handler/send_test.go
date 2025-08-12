package handler

import (
	"bytes"
	"encoding/json"
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

func TestHandler_SendCoin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTransactionService := mocks.NewMockTransaction(ctrl)
	mockLog := logrus.New()
	handler := &Handler{services: &service.Service{Transaction: mockTransactionService}, log: mockLog}

	tests := []struct {
		name         string
		userID       int64
		requestBody  entity.SendCoinRequest
		mockBehavior func()
		wantStatus   int
		wantBody     string
	}{
		{
			name:        "Success",
			userID:      1,
			requestBody: entity.SendCoinRequest{ToUser: "recipient", Amount: 50},
			mockBehavior: func() {
				mockTransactionService.EXPECT().
					SendCoin(gomock.Any(), int64(1), "recipient", int64(50)).
					Return(nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   `{"status":"coins were successfully sent to the user"}`,
		},
		{
			name:         "Invalid request format",
			userID:       1,
			requestBody:  entity.SendCoinRequest{},
			mockBehavior: func() {},
			wantStatus:   http.StatusBadRequest,
			wantBody:     `{"errors":"invalid request format"}`,
		},
		{
			name:        "Recipient not found",
			userID:      1,
			requestBody: entity.SendCoinRequest{ToUser: "unknown", Amount: 50},
			mockBehavior: func() {
				mockTransactionService.EXPECT().
					SendCoin(gomock.Any(), int64(1), "unknown", int64(50)).
					Return(entity.ErrRecipientNotFound)
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"errors":"recipient not found"}`,
		},
		{
			name:        "Insufficient balance",
			userID:      1,
			requestBody: entity.SendCoinRequest{ToUser: "recipient", Amount: 1000},
			mockBehavior: func() {
				mockTransactionService.EXPECT().
					SendCoin(gomock.Any(), int64(1), "recipient", int64(1000)).
					Return(entity.ErrInsufficientBalance)
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"errors":"insufficient balance"}`,
		},
		{
			name:        "Cannot send to self",
			userID:      1,
			requestBody: entity.SendCoinRequest{ToUser: "sender", Amount: 50},
			mockBehavior: func() {
				mockTransactionService.EXPECT().
					SendCoin(gomock.Any(), int64(1), "sender", int64(50)).
					Return(entity.ErrSendThemselves)
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"errors":"cannot send coins to yourself"}`,
		},
		{
			name:        "Transaction failure",
			userID:      1,
			requestBody: entity.SendCoinRequest{ToUser: "recipient", Amount: 50},
			mockBehavior: func() {
				mockTransactionService.EXPECT().
					SendCoin(gomock.Any(), int64(1), "recipient", int64(50)).
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

			body, _ := json.Marshal(tt.requestBody)
			c.Request = httptest.NewRequest(http.MethodPost, "/sendCoin", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")

			handler.sendCoin(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			assert.JSONEq(t, tt.wantBody, w.Body.String())
		})
	}
}
