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

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(0)

	if err := InitRedis(cfg); err != nil {
		return err
	}

	return autoMigrate()
}

func autoMigrate() error {
	return DB.AutoMigrate(
		&model.User{},
		&model.Book{},
		&model.Chapter{},
		&model.Article{},
		&model.BookshelfItem{},
		&model.Comment{},
		&model.ParagraphComment{},
		&model.Follow{},
		&model.Like{},
		&model.FeedActivity{},
		&model.Work{},
		&model.WorkChapter{},
		&model.ArticleStats{},
		&model.ReadingHistory{},
		&model.Notification{},
	)
}
