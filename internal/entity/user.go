package entity

type User struct {
	ID       int64  `json:"-" db:"id"`
	Username string `json:"username" db:"username"`
	Password string `json:"password" db:"password_hash"`
	Coins    int64  `json:"coins" db:"coins"`
}
