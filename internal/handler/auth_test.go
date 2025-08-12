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

func TestHandler_Authenticate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := mocks.NewMockAuthorization(ctrl)
	mockService := &service.Service{Authorization: mockAuthService}
	mockLog := logrus.New()
	handler := &Handler{services: mockService, log: mockLog}

	tests := []struct {
		name         string
		requestBody  entity.AuthRequest
		mockBehavior func()
		wantCode     int
		wantBody     string
	}{
		{
			name: "User Exists",
			requestBody: entity.AuthRequest{
				Username: "testuser",
				Password: "testpass",
			},
			mockBehavior: func() {
				mockAuthService.EXPECT().GetUser(gomock.Any(), "testuser").Return(entity.User{ID: 1, Username: "testuser"}, nil)
				mockAuthService.EXPECT().GenerateToken(gomock.Any(), "testuser", "testpass").Return("jwt-token", nil)
			},
			wantCode: http.StatusOK,
			wantBody: `{"token":"jwt-token"}`,
		},
		{
			name: "User Created",
			requestBody: entity.AuthRequest{
				Username: "newuser",
				Password: "newpass",
			},
			mockBehavior: func() {
				mockAuthService.EXPECT().GetUser(gomock.Any(), "newuser").Return(entity.User{}, errors.New("user not found"))
				mockAuthService.EXPECT().CreateUser(gomock.Any(), "newuser", "newpass").Return(nil)
				mockAuthService.EXPECT().GenerateToken(gomock.Any(), "newuser", "newpass").Return("jwt-token", nil)
			},
			wantCode: http.StatusOK,
			wantBody: `{"token":"jwt-token"}`,
		},
		{
			name:         "Invalid Request Format",
			requestBody:  entity.AuthRequest{},
			mockBehavior: func() {},
			wantCode:     http.StatusBadRequest,
			wantBody:     `{"errors":"invalid request format"}`,
		},
		{
			name: "Error Creating User",
			requestBody: entity.AuthRequest{
				Username: "failuser",
				Password: "failpass",
			},
			mockBehavior: func() {
				mockAuthService.EXPECT().GetUser(gomock.Any(), "failuser").Return(entity.User{}, errors.New("user not found"))
				mockAuthService.EXPECT().CreateUser(gomock.Any(), "failuser", "failpass").Return(errors.New("creation failed"))
				mockAuthService.EXPECT().GenerateToken(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			wantCode: http.StatusInternalServerError,
			wantBody: `{"errors":"creation failed"}`,
		},
		{
			name: "Error Generating Token",
			requestBody: entity.AuthRequest{
				Username: "testuser",
				Password: "testpass",
			},
			mockBehavior: func() {
				mockAuthService.EXPECT().GetUser(gomock.Any(), "testuser").Return(entity.User{ID: 1, Username: "testuser"}, nil)
				mockAuthService.EXPECT().GenerateToken(gomock.Any(), "testuser", "testpass").Return("", errors.New("token error"))
			},
			wantCode: http.StatusUnauthorized,
			wantBody: `{"errors":"token error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior()

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/auth", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			c, _ := gin.CreateTestContext(w)
			c.Request = req

			handler.authenticate(c)

			assert.Equal(t, tt.wantCode, w.Code)
			assert.JSONEq(t, tt.wantBody, w.Body.String())
		})
	}
}
