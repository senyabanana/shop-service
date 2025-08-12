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

func TestHandler_UserIdentity(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := mocks.NewMockAuthorization(ctrl)
	mockLog := logrus.New()

	handler := &Handler{services: &service.Service{Authorization: mockAuthService}, log: mockLog}

	tests := []struct {
		name         string
		authHeader   string
		mockBehavior func()
		wantStatus   int
		wantBody     string
	}{
		{
			name:       "Success",
			authHeader: "Bearer valid_token",
			mockBehavior: func() {
				mockAuthService.EXPECT().ParseToken("valid_token").Return(int64(1), nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   ``,
		},
		{
			name:         "Empty auth header",
			authHeader:   "",
			mockBehavior: func() {},
			wantStatus:   http.StatusUnauthorized,
			wantBody:     `{"errors":"empty auth header"}`,
		},
		{
			name:         "Invalid auth header format",
			authHeader:   "InvalidToken",
			mockBehavior: func() {},
			wantStatus:   http.StatusUnauthorized,
			wantBody:     `{"errors":"invalid auth header format"}`,
		},
		{
			name:       "Invalid token",
			authHeader: "Bearer invalid_token",
			mockBehavior: func() {
				mockAuthService.EXPECT().ParseToken("invalid_token").Return(int64(0), errors.New("invalid or expired token"))
			},
			wantStatus: http.StatusUnauthorized,
			wantBody:   `{"errors":"invalid or expired token"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior()

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
			c.Request.Header.Set(authorizationHeader, tt.authHeader)

			handler.userIdentity(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.JSONEq(t, tt.wantBody, w.Body.String())
			}
		})
	}
}

func TestHandler_GetUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockLog := logrus.New()
	handler := &Handler{log: mockLog}

	tests := []struct {
		name        string
		userContext interface{}
		wantUserID  int64
		wantErr     error
	}{
		{
			name:        "Success",
			userContext: int64(1),
			wantUserID:  1,
			wantErr:     nil,
		},
		{
			name:        "User ID not in context",
			userContext: nil,
			wantUserID:  0,
			wantErr:     entity.ErrUserNotFound,
		},
		{
			name:        "User ID has wrong type",
			userContext: "string",
			wantUserID:  0,
			wantErr:     entity.ErrInvalidUserIDType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())

			if tt.userContext != nil {
				c.Set(userCtx, tt.userContext)
			}

			userID, err := handler.getUserID(c)

			assert.Equal(t, tt.wantUserID, userID)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}
