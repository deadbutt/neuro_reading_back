package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Server     ServerConfig
	Database   DatabaseConfig
	JWT        JWTConfig
	SMS        SMSConfig
	Email      EmailConfig
	UploadDir  string
	ArticleDir string
}

type ServerConfig struct {
	Host      string
	Port      string
	PublicURL string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	Charset  string
}

type JWTConfig struct {
	Secret        string
	AccessExpiry  int64
	RefreshExpiry int64
}

type SMSConfig struct {
	AccessKeyID     string
	AccessKeySecret string
	SignName        string
	TemplateCode    string
}

type EmailConfig struct {
	SMTPHost     string
	SMTPPort     string
	FromEmail    string
	FromPassword string
}

func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		Server: ServerConfig{
			Host:      getEnv("SERVER_HOST", "0.0.0.0"),
			Port:      getEnv("SERVER_PORT", "9091"),
			PublicURL: getEnv("SERVER_PUBLIC_URL", ""),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "3306"),
			User:     getEnv("DB_USER", "root"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "neuro_reading"),
			Charset:  "utf8mb4",
		},
		JWT: JWTConfig{
			Secret:        getEnv("JWT_SECRET", "neuro-reading-secret-key-change-in-production"),
			AccessExpiry:  7 * 24 * 3600,
			RefreshExpiry: 30 * 24 * 3600,
		},
		SMS: SMSConfig{
			AccessKeyID:     getEnv("SMS_ACCESS_KEY_ID", ""),
			AccessKeySecret: getEnv("SMS_ACCESS_KEY_SECRET", ""),
			SignName:        getEnv("SMS_SIGN_NAME", ""),
			TemplateCode:    getEnv("SMS_TEMPLATE_CODE", ""),
		},
		Email: EmailConfig{
			SMTPHost:     getEnv("EMAIL_SMTP_HOST", "smtp.qq.com"),
			SMTPPort:     getEnv("EMAIL_SMTP_PORT", "587"),
			FromEmail:    getEnv("EMAIL_FROM", ""),
			FromPassword: getEnv("EMAIL_PASSWORD", ""),
		},
		UploadDir:  getEnv("UPLOAD_DIR", "./uploads"),
		ArticleDir: getEnv("ARTICLE_DIR", "./articles"),
	}
}

func (c *Config) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local",
		c.Database.User,
		c.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.Name,
		c.Database.Charset,
	)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}