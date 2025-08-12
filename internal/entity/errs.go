package entity

import "errors"

var (
	ErrUserNotFound           = errors.New("user not found")
	ErrInvalidUserIDType      = errors.New("user id is of invalid type")
	ErrIncorrectPassword      = errors.New("incorrect password")
	ErrInvalidSigningMethod   = errors.New("invalid signing method")
	ErrInvalidTokenClaimsType = errors.New("token claims are not of type *tokenClaims")
	ErrRecipientNotFound      = errors.New("recipient not found")
	ErrSendThemselves         = errors.New("cannot send coins to yourself")
	ErrInsufficientBalance    = errors.New("insufficient balance")
	ErrItemNotFound           = errors.New("item not found")
)
