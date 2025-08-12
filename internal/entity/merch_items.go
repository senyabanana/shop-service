package entity

type MerchItems struct {
	ID       int64  `json:"id" db:"id"`
	ItemType string `json:"item_type" db:"item_type"`
	Price    int64  `json:"price" db:"price"`
}
