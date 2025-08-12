package entity

type AuthResponse struct {
	Token string `json:"token"`
}

type AuthRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}
