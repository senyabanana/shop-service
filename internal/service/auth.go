package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/dgrijalva/jwt-go"
	"github.com/sirupsen/logrus"

	"github.com/senyabanana/shop-service/internal/entity"
	"github.com/senyabanana/shop-service/internal/repository"
)

const (
	salt     = "random_salt_string"
	tokenTTL = 12 * time.Hour
)

type tokenClaims struct {
	jwt.StandardClaims
	UserID int64 `json:"user_id"`
}

type AuthService struct {
	userRepo     repository.UserRepository
	trManager    *manager.Manager
	jwtSecretKey string
	log          *logrus.Logger
}

func NewAuthService(repo repository.UserRepository, trManager *manager.Manager, jwtSecretKey string, log *logrus.Logger) *AuthService {
	return &AuthService{
		userRepo:     repo,
		trManager:    trManager,
		jwtSecretKey: jwtSecretKey,
		log:          log,
	}
}

func (s *AuthService) GetUser(ctx context.Context, username string) (entity.User, error) {
	user, err := s.userRepo.GetUser(ctx, username)
	if err != nil {
		s.log.Warnf("User not found: %s", username)
		return entity.User{}, entity.ErrUserNotFound
	}

	return user, nil
}

func (s *AuthService) CreateUser(ctx context.Context, username, password string) error {
	hashedPassword := generatePasswordHash(password)

	newUser := entity.User{
		Username: username,
		Password: hashedPassword,
		Coins:    1000,
	}

	_, err := s.userRepo.CreateUser(ctx, newUser)
	if err != nil {
		s.log.Errorf("Failed to create user %s: %v", username, err)
		return err
	}

	s.log.Infof("User %s created successfully", username)
	return nil
}

func (s *AuthService) GenerateToken(ctx context.Context, username, password string) (string, error) {
	user, err := s.userRepo.GetUser(ctx, username)
	if err != nil {
		s.log.Warnf("GenerateToken: User %s not found", username)
		return "", err
	}

	hashedPassword := generatePasswordHash(password)
	if user.Password != hashedPassword {
		s.log.Warnf("GenerateToken: Invalid password for user %s", username)
		return "", entity.ErrIncorrectPassword
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &tokenClaims{
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(tokenTTL).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		user.ID,
	})

	s.log.Infof("Generated token for user %s", username)
	return token.SignedString([]byte(s.jwtSecretKey))
}

func (s *AuthService) ParseToken(accessToken string) (int64, error) {
	token, err := jwt.ParseWithClaims(accessToken, &tokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			s.log.Warn("ParseToken: invalid signing method")
			return nil, entity.ErrInvalidSigningMethod
		}

		return []byte(s.jwtSecretKey), nil
	})
	if err != nil {
		s.log.Warnf("ParseToken: failed to parse token: %s", err.Error())
		return 0, err
	}

	claims, ok := token.Claims.(*tokenClaims)
	if !ok {
		s.log.Warn("ParseToken: token claims are invalid")
		return 0, entity.ErrInvalidTokenClaimsType
	}

	return claims.UserID, nil
}

func generatePasswordHash(password string) string {
	hash := sha256.Sum256([]byte(password + salt))
	return fmt.Sprintf("%x", hash)
}
