package db

import (
	"neuro-reading/config"
	"neuro-reading/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init(cfg *config.Config) error {
	var err error
	DB, err = gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return err
	}

	return autoMigrate()
}

func autoMigrate() error {
	return DB.AutoMigrate(
		&model.User{},
		&model.Author{},
		&model.Book{},
		&model.Chapter{},
		&model.BookshelfItem{},
		&model.Comment{},
		&model.ParagraphComment{},
		&model.Follow{},
		&model.Like{},
		&model.FeedActivity{},
	)
}