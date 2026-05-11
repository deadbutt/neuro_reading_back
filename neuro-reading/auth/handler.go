package auth

import (
	"fmt"
	"neuro-reading/config"
	"neuro-reading/db"
	"neuro-reading/model"
	"neuro-reading/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	cfg *config.Config
}

func NewHandler(cfg *config.Config) *Handler {
	return &Handler{cfg: cfg}
}

func (h *Handler) SendCode(c *gin.Context) {
	var req model.SendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, model.Error(1001, "参数错误"))
		return
	}

	if err := utils.StoreVerifyCode(req.Account, utils.GenerateCode(), req.Type); err != nil {
		if err.Error() == "发送过于频繁" {
			c.JSON(400, model.Error(1001, "发送过于频繁，请稍后再试"))
			return
		}
		c.JSON(500, model.Error(1005, "服务器内部错误"))
		return
	}

	code, _, _ := utils.GetVerifyCode(req.Account)

	if req.Type == "email" && h.cfg.Email.FromEmail != "" {
		emailCfg := &utils.EmailConfig{
			SMTPHost:     h.cfg.Email.SMTPHost,
			SMTPPort:     h.cfg.Email.SMTPPort,
			FromEmail:    h.cfg.Email.FromEmail,
			FromPassword: h.cfg.Email.FromPassword,
		}
		if err := utils.SendEmailCode(emailCfg, req.Account, code); err != nil {
			fmt.Printf("邮件发送失败: %v\n", err)
		}
	} else if req.Type == "qq" && h.cfg.SMS.AccessKeyID != "" {
		sms := utils.NewAliyunSMS(h.cfg.SMS.AccessKeyID, h.cfg.SMS.AccessKeySecret, h.cfg.SMS.SignName, h.cfg.SMS.TemplateCode)
		if err := sms.SendSMS(req.Account, code); err != nil {
			fmt.Printf("短信发送失败: %v\n", err)
		}
	} else {
		fmt.Printf("验证码 [%s] 已发送到 %s (%s)\n", code, req.Account, req.Type)
	}

	c.JSON(200, model.Success(nil))
}

func (h *Handler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, model.Error(1001, "参数错误"))
		return
	}

	var user model.User
	if err := db.DB.Where("account = ?", req.Account).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(400, model.Error(2005, "账号不存在"))
			return
		}
		c.JSON(500, model.Error(1005, "服务器内部错误"))
		return
	}

	if !utils.CheckPassword(req.Password, user.Password) {
		c.JSON(400, model.Error(2002, "账号或密码错误"))
		return
	}

	token, _, err := utils.GenerateToken(user.UserID, "access", &h.cfg.JWT)
	if err != nil {
		c.JSON(500, model.Error(1005, "服务器内部错误"))
		return
	}

	refreshToken, _, err := utils.GenerateToken(user.UserID, "refresh", &h.cfg.JWT)
	if err != nil {
		c.JSON(500, model.Error(1005, "服务器内部错误"))
		return
	}

	c.JSON(200, model.Success(model.LoginResponse{
		UserID:       user.UserID,
		Account:      user.Account,
		Nickname:     user.Nickname,
		Avatar:       user.Avatar,
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresIn:    h.cfg.JWT.AccessExpiry,
	}))
}

func (h *Handler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, model.Error(1001, "参数错误"))
		return
	}

	if req.Password != req.ConfirmPassword {
		c.JSON(400, model.Error(2004, "密码不一致"))
		return
	}

	code, _, err := utils.GetVerifyCode(req.Account)
	if err != nil || code != req.Code {
		c.JSON(400, model.Error(2003, "验证码错误/过期"))
		return
	}

	var existingUser model.User
	if err := db.DB.Where("account = ?", req.Account).First(&existingUser).Error; err == nil {
		c.JSON(400, model.Error(2001, "账号已存在"))
		return
	}

	nickname := req.Nickname
	if nickname == "" {
		nickname = "用户" + req.Account
		if len(req.Account) > 6 {
			nickname = "用户" + req.Account[len(req.Account)-6:]
		}
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(500, model.Error(1005, "服务器内部错误"))
		return
	}

	user := model.User{
		UserID:   utils.GenerateUserID(),
		Account:  req.Account,
		Password: hashedPassword,
		Nickname: nickname,
		Avatar:   "",
	}

	if err := db.DB.Create(&user).Error; err != nil {
		c.JSON(500, model.Error(1005, "服务器内部错误"))
		return
	}

	utils.DeleteVerifyCode(req.Account)

	token, _, err := utils.GenerateToken(user.UserID, "access", &h.cfg.JWT)
	if err != nil {
		c.JSON(500, model.Error(1005, "服务器内部错误"))
		return
	}

	refreshToken, _, err := utils.GenerateToken(user.UserID, "refresh", &h.cfg.JWT)
	if err != nil {
		c.JSON(500, model.Error(1005, "服务器内部错误"))
		return
	}

	c.JSON(200, model.Success(model.LoginResponse{
		UserID:       user.UserID,
		Account:      user.Account,
		Nickname:     user.Nickname,
		Avatar:       user.Avatar,
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresIn:    h.cfg.JWT.AccessExpiry,
	}))
}

func (h *Handler) ForgotPassword(c *gin.Context) {
	var req model.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, model.Error(1001, "参数错误"))
		return
	}

	code, _, err := utils.GetVerifyCode(req.Account)
	if err != nil || code != req.Code {
		c.JSON(400, model.Error(2003, "验证码错误/过期"))
		return
	}

	var user model.User
	if err := db.DB.Where("account = ?", req.Account).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(400, model.Error(2005, "账号不存在"))
			return
		}
		c.JSON(500, model.Error(1005, "服务器内部错误"))
		return
	}

	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(500, model.Error(1005, "服务器内部错误"))
		return
	}

	if err := db.DB.Model(&user).Update("password", hashedPassword).Error; err != nil {
		c.JSON(500, model.Error(1005, "服务器内部错误"))
		return
	}

	utils.DeleteVerifyCode(req.Account)

	c.JSON(200, model.Success(nil))
}

func (h *Handler) RefreshToken(c *gin.Context) {
	var req model.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, model.Error(1001, "参数错误"))
		return
	}

	claims, err := utils.ParseToken(req.RefreshToken, &h.cfg.JWT)
	if err != nil {
		c.JSON(401, model.Error(1002, "未授权/Token过期"))
		return
	}

	if claims.Type != "refresh" {
		c.JSON(401, model.Error(1002, "未授权/Token过期"))
		return
	}

	var user model.User
	if err := db.DB.Where("user_id = ?", claims.UserID).First(&user).Error; err != nil {
		c.JSON(401, model.Error(1002, "未授权/Token过期"))
		return
	}

	token, _, err := utils.GenerateToken(user.UserID, "access", &h.cfg.JWT)
	if err != nil {
		c.JSON(500, model.Error(1005, "服务器内部错误"))
		return
	}

	refreshToken, _, err := utils.GenerateToken(user.UserID, "refresh", &h.cfg.JWT)
	if err != nil {
		c.JSON(500, model.Error(1005, "服务器内部错误"))
		return
	}

	c.JSON(200, model.Success(model.LoginResponse{
		UserID:       user.UserID,
		Account:      user.Account,
		Nickname:     user.Nickname,
		Avatar:       user.Avatar,
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresIn:    h.cfg.JWT.AccessExpiry,
	}))
}
