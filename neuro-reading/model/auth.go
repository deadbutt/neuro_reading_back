package model

type SendCodeRequest struct {
	Account string `json:"account" binding:"required"`
	Type    string `json:"type" binding:"required,oneof=qq email"`
}

type LoginRequest struct {
	Account  string `json:"account" binding:"required"`
	Password string `json:"password" binding:"required,len=32"`
}

type RegisterRequest struct {
	Account         string `json:"account" binding:"required"`
	Code            string `json:"code" binding:"required,len=6"`
	Password        string `json:"password" binding:"required,len=32"`
	ConfirmPassword string `json:"confirmPassword" binding:"required,len=32"`
	Nickname        string `json:"nickname,omitempty"`
}

type ForgotPasswordRequest struct {
	Account     string `json:"account" binding:"required"`
	Code        string `json:"code" binding:"required,len=6"`
	NewPassword string `json:"newPassword" binding:"required,len=32"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type LoginResponse struct {
	UserID       string `json:"userId"`
	Account      string `json:"account"`
	Nickname     string `json:"nickname"`
	Avatar       string `json:"avatar"`
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"`
}