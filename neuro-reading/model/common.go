package model

type PageRequest struct {
	Page     int `form:"page" binding:"min=1"`
	PageSize int `form:"pageSize" binding:"min=1,max=100"`
}

func (p *PageRequest) GetOffset() int {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 {
		p.PageSize = 20
	}
	return (p.Page - 1) * p.PageSize
}

func (p *PageRequest) GetLimit() int {
	if p.PageSize < 1 {
		p.PageSize = 20
	}
	return p.PageSize
}

type PageResponse struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
	HasMore  bool        `json:"hasMore"`
}

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func NewResponse(code int, message string, data interface{}) Response {
	return Response{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

func Success(data interface{}) Response {
	return NewResponse(0, "success", data)
}

func Error(code int, message string) Response {
	return NewResponse(code, message, nil)
}

func PageSuccess(list interface{}, total int64, page, pageSize int) Response {
	hasMore := int64(page*pageSize) < total
	return Success(PageResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		HasMore:  hasMore,
	})
}

type SendCodeRequest struct {
	Account string `json:"account" binding:"required"`
	Type    string `json:"type" binding:"required,oneof=email qq"`
}

type LoginRequest struct {
	Account  string `json:"account" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Account         string `json:"account" binding:"required"`
	Password        string `json:"password" binding:"required"`
	ConfirmPassword string `json:"confirmPassword" binding:"required"`
	Code            string `json:"code" binding:"required"`
	Nickname        string `json:"nickname"`
}

type ForgotPasswordRequest struct {
	Account     string `json:"account" binding:"required"`
	Code        string `json:"code" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required"`
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
