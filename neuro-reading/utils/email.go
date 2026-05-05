package utils

import (
	"fmt"
	"net/smtp"
)

type EmailConfig struct {
	SMTPHost     string
	SMTPPort     string
	FromEmail    string
	FromPassword string
}

func SendEmailCode(cfg *EmailConfig, toEmail, code string) error {
	subject := "Neuro阅读 - 验证码"
	body := fmt.Sprintf("您的验证码是：%s，5分钟内有效。如非本人操作，请忽略此邮件。", code)

	msg := []byte(fmt.Sprintf(
		"To: %s\r\n"+
			"From: %s\r\n"+
			"Subject: %s\r\n"+
			"Content-Type: text/plain; charset=UTF-8\r\n"+
			"\r\n"+
			"%s",
		toEmail,
		cfg.FromEmail,
		subject,
		body,
	))

	addr := cfg.SMTPHost + ":" + cfg.SMTPPort
	auth := smtp.PlainAuth("", cfg.FromEmail, cfg.FromPassword, cfg.SMTPHost)

	err := smtp.SendMail(addr, auth, cfg.FromEmail, []string{toEmail}, msg)
	if err != nil {
		fmt.Printf("SMTP错误详情: %v\n", err)
		return fmt.Errorf("邮件发送失败: %w", err)
	}

	fmt.Printf("验证码邮件已发送到 %s\n", toEmail)
	return nil
}