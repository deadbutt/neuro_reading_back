package router

import (
	"neuro-reading/article"
	"neuro-reading/auth"
	"neuro-reading/author"
	"neuro-reading/book"
	"neuro-reading/bookshelf"
	"neuro-reading/comment"
	"neuro-reading/config"
	"neuro-reading/creator"
	"neuro-reading/feed"
	"neuro-reading/middleware"
	"neuro-reading/paragraph_comment"
	"neuro-reading/upload"
	"neuro-reading/user"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/gin-gonic/gin"
)

func Setup(cfg *config.Config) *gin.Engine {
	r := gin.Default()

	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.RateLimit())

	authHandler := auth.NewHandler(cfg)
	userHandler := user.NewHandler()
	bookshelfHandler := bookshelf.NewHandler()
	bookHandler := book.NewHandler()
	commentHandler := comment.NewHandler()
	paragraphCommentHandler := paragraph_comment.NewHandler()
	authorHandler := author.NewHandler()
	feedHandler := feed.NewHandler()
	articleHandler := article.NewHandler(cfg)
	creatorHandler := creator.NewHandler(cfg)

	api := r.Group("/api/v1")

	authGroup := api.Group("/auth")
	{
		authGroup.POST("/send-code", authHandler.SendCode)
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/forgot-password", authHandler.ForgotPassword)
		authGroup.POST("/refresh-token", authHandler.RefreshToken)
	}

	userGroup := api.Group("/user")
	userGroup.Use(middleware.AuthMiddleware(&cfg.JWT))
	{
		userGroup.GET("/profile", userHandler.GetProfile)
		userGroup.PUT("/profile", userHandler.UpdateProfile)
		userGroup.POST("/follow/:authorId", userHandler.Follow)
		userGroup.DELETE("/follow/:authorId", userHandler.Unfollow)
		userGroup.GET("/following", userHandler.GetFollowing)
		userGroup.GET("/reading-history", userHandler.GetReadingHistory)
		userGroup.DELETE("/reading-history/:historyId", userHandler.DeleteReadingHistory)
		userGroup.DELETE("/reading-history", userHandler.ClearReadingHistory)
	}

	bookshelfGroup := api.Group("/bookshelf")
	bookshelfGroup.Use(middleware.AuthMiddleware(&cfg.JWT))
	{
		bookshelfGroup.GET("", bookshelfHandler.GetList)
		bookshelfGroup.POST("/:bookId", bookshelfHandler.Add)
		bookshelfGroup.DELETE("/:bookId", bookshelfHandler.Remove)
		bookshelfGroup.PUT("/:bookId/progress", bookshelfHandler.UpdateProgress)
	}

	bookGroup := api.Group("/books")
	{
		bookGroup.GET("/recommend", bookHandler.GetRecommend)
		bookGroup.GET("/hot", bookHandler.GetHot)
		bookGroup.GET("/latest", bookHandler.GetLatest)
		bookGroup.GET("/search", bookHandler.Search)
		bookGroup.GET("/search/hot", bookHandler.GetHotKeywords)
		bookGroup.GET("/:bookId", bookHandler.GetDetail)
		bookGroup.GET("/:bookId/chapters", bookHandler.GetChapters)
		bookGroup.GET("/:bookId/chapters/:chapterId", bookHandler.GetChapterContent)
	}

	commentGroup := api.Group("/books/:bookId/comments")
	commentGroup.Use(middleware.AuthMiddleware(&cfg.JWT))
	{
		commentGroup.GET("", commentHandler.GetComments)
		commentGroup.POST("", commentHandler.CreateComment)
	}

	articleCommentGroup := api.Group("/articles/:articleId/comments")
	articleCommentGroup.Use(middleware.AuthMiddleware(&cfg.JWT))
	{
		articleCommentGroup.GET("", commentHandler.GetComments)
		articleCommentGroup.POST("", commentHandler.CreateComment)
	}

	api.POST("/comments/:commentId/like", middleware.AuthMiddleware(&cfg.JWT), commentHandler.LikeComment)
	api.DELETE("/comments/:commentId/like", middleware.AuthMiddleware(&cfg.JWT), commentHandler.UnlikeComment)

	paragraphCommentGroup := api.Group("/chapters/:chapterId/paragraph-comments")
	paragraphCommentGroup.Use(middleware.AuthMiddleware(&cfg.JWT))
	{
		paragraphCommentGroup.GET("", paragraphCommentHandler.GetList)
		paragraphCommentGroup.POST("", paragraphCommentHandler.Create)
	}

	api.POST("/paragraph-comments/:commentId/like", middleware.AuthMiddleware(&cfg.JWT), paragraphCommentHandler.Like)
	api.DELETE("/paragraph-comments/:commentId/like", middleware.AuthMiddleware(&cfg.JWT), paragraphCommentHandler.Unlike)

	authorGroup := api.Group("/authors")
	{
		authorGroup.GET("/:authorId", authorHandler.GetProfile)
		authorGroup.GET("/:authorId/works", authorHandler.GetWorks)
		authorGroup.GET("/:authorId/activities", authorHandler.GetActivities)
	}

	feedGroup := api.Group("/feed")
	feedGroup.Use(middleware.AuthMiddleware(&cfg.JWT))
	{
		feedGroup.GET("", feedHandler.GetFeed)
		feedGroup.POST("/:feedId/like", feedHandler.LikeFeed)
		feedGroup.DELETE("/:feedId/like", feedHandler.UnlikeFeed)
	}

	uploadHandler := upload.NewHandler(cfg)
	api.POST("/upload/avatar", middleware.AuthMiddleware(&cfg.JWT), uploadHandler.UploadAvatar)

	articleGroup := api.Group("/articles")
	{
		articleGroup.GET("", articleHandler.List)
		articleGroup.GET("/search", articleHandler.Search)
		articleGroup.GET("/:articleId", articleHandler.Detail)
		articleGroup.GET("/:articleId/chapters/:chapterIndex", articleHandler.Chapter)
		articleGroup.POST("/upload", middleware.AuthMiddleware(&cfg.JWT), articleHandler.Upload)
		articleGroup.DELETE("/:articleId", middleware.AuthMiddleware(&cfg.JWT), articleHandler.Delete)
	}

	creatorGroup := api.Group("/creator")
	{
		creatorGroup.POST("/register", creatorHandler.Register)
		creatorGroup.POST("/login", creatorHandler.Login)
		creatorGroup.GET("/profile", middleware.AuthMiddleware(&cfg.JWT), creatorHandler.GetProfile)
		creatorGroup.PUT("/profile", middleware.AuthMiddleware(&cfg.JWT), creatorHandler.UpdateProfile)
		creatorGroup.GET("/works", middleware.AuthMiddleware(&cfg.JWT), creatorHandler.GetMyWorks)
		creatorGroup.POST("/works", middleware.AuthMiddleware(&cfg.JWT), creatorHandler.CreateWork)
		creatorGroup.GET("/works/:workId", middleware.AuthMiddleware(&cfg.JWT), creatorHandler.GetWork)
		creatorGroup.PUT("/works/:workId", middleware.AuthMiddleware(&cfg.JWT), creatorHandler.UpdateWork)
		creatorGroup.DELETE("/works/:workId", middleware.AuthMiddleware(&cfg.JWT), creatorHandler.DeleteWork)
		creatorGroup.POST("/works/:workId/publish", middleware.AuthMiddleware(&cfg.JWT), creatorHandler.PublishWork)
		creatorGroup.POST("/works/:workId/chapters", middleware.AuthMiddleware(&cfg.JWT), creatorHandler.CreateChapter)
		creatorGroup.GET("/works/:workId/chapters/:chapterId", middleware.AuthMiddleware(&cfg.JWT), creatorHandler.GetChapter)
		creatorGroup.PUT("/works/:workId/chapters/:chapterId", middleware.AuthMiddleware(&cfg.JWT), creatorHandler.UpdateChapter)
		creatorGroup.DELETE("/works/:workId/chapters/:chapterId", middleware.AuthMiddleware(&cfg.JWT), creatorHandler.DeleteChapter)
		creatorGroup.POST("/works/upload/docx", middleware.AuthMiddleware(&cfg.JWT), creatorHandler.UploadDocx)
		creatorGroup.POST("/works/upload/txt", middleware.AuthMiddleware(&cfg.JWT), creatorHandler.UploadTxt)
		creatorGroup.POST("/works/upload/cover", middleware.AuthMiddleware(&cfg.JWT), creatorHandler.UploadCover)
	}

	r.GET("/uploads/:filename", func(c *gin.Context) {
		filename := c.Param("filename")
		if filename == "" {
			c.JSON(404, gin.H{"code": 1004, "msg": "文件不存在"})
			return
		}

		filePath := path.Join(cfg.UploadDir, filename)

		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			c.JSON(404, gin.H{"code": 1004, "msg": "文件不存在"})
			return
		}

		c.Header("Cache-Control", "public, max-age=2592000")
		c.Header("Expires", time.Now().AddDate(0, 0, 30).Format(http.TimeFormat))
		c.File(filePath)
	})

	r.GET("/uploads/covers/:filename", func(c *gin.Context) {
		filename := c.Param("filename")
		if filename == "" {
			c.JSON(404, gin.H{"code": 1004, "msg": "文件不存在"})
			return
		}

		filePath := path.Join(cfg.UploadDir, "covers", filename)

		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			c.JSON(404, gin.H{"code": 1004, "msg": "文件不存在"})
			return
		}

		c.Header("Cache-Control", "public, max-age=2592000")
		c.Header("Expires", time.Now().AddDate(0, 0, 30).Format(http.TimeFormat))
		c.File(filePath)
	})

	return r
}
