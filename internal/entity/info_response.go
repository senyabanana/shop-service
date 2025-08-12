package entity

type InfoResponse struct {
	Coins       int64           `json:"coins"`
	Inventory   []InventoryItem `json:"inventory"`
	CoinHistory CoinHistory     `json:"coinHistory"`
}

type InventoryItem struct {
	Type     string `json:"type"`
	Quantity int    `json:"quantity"`
}

type CoinHistory struct {
	Received []TransactionDetail `json:"received"`
	Sent     []TransactionDetail `json:"sent"`
}

type TransactionDetail struct {
	FromUser string `json:"fromUser,omitempty" db:"from_user"`
	ToUser   string `json:"toUser,omitempty" db:"to_user"`
	Amount   int64  `json:"amount" db:"amount"`
}
