package entity

type SendCoinRequest struct {
	ToUser string `json:"toUser" binding:"required"`
	Amount int64  `json:"amount" binding:"required,gt=0"`
}
