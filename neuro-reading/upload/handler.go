package upload

import (
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"neuro-reading/config"
	"neuro-reading/db"
	"neuro-reading/model"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	cfg       *config.Config
	uploadDir string
}

func NewHandler(cfg *config.Config) *Handler {
	uploadDir := cfg.UploadDir
	if uploadDir == "" {
		uploadDir = "./uploads"
	}

	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		fmt.Printf("创建上传目录失败: %v\n", err)
	}

	return &Handler{
		cfg:       cfg,
		uploadDir: uploadDir,
	}
}

func (h *Handler) UploadAvatar(c *gin.Context) {
	userID := c.GetString("userId")

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(400, model.Error(1001, "请选择要上传的文件"))
		return
	}
	defer file.Close()

	ext := path.Ext(header.Filename)
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" {
		c.JSON(400, model.Error(1001, "仅支持 jpg、png、gif 格式的图片"))
		return
	}

	if header.Size > 5*1024*1024 {
		c.JSON(400, model.Error(1001, "图片大小不能超过 5MB"))
		return
	}

	var user model.User
	var oldAvatar string
	if err := db.DB.Where("user_id = ?", userID).First(&user).Error; err == nil {
		oldAvatar = user.Avatar
	}

	filename := fmt.Sprintf("%d_%s%s", time.Now().Unix(), strconv.Itoa(int(time.Now().UnixNano()%10000)), ext)
	filePath := path.Join(h.uploadDir, filename)

	srcImage, _, err := image.Decode(file)
	if err != nil {
		c.JSON(400, model.Error(1001, "图片解析失败"))
		return
	}

	resizedImage := imaging.Fill(srcImage, 256, 256, imaging.Center, imaging.Lanczos)

	out, err := os.Create(filePath)
	if err != nil {
		c.JSON(500, model.Error(1005, "文件保存失败"))
		return
	}
	defer out.Close()

	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg":
		err = jpeg.Encode(out, resizedImage, &jpeg.Options{Quality: 80})
	case ".png":
		err = png.Encode(out, resizedImage)
	case ".gif":
		err = gif.Encode(out, resizedImage, nil)
	default:
		err = jpeg.Encode(out, resizedImage, &jpeg.Options{Quality: 80})
	}

	if err != nil {
		c.JSON(500, model.Error(1005, "图片压缩保存失败"))
		return
	}

	if oldAvatar != "" && strings.Contains(oldAvatar, "/uploads/") {
		oldFilename := path.Base(oldAvatar)
		oldFilePath := path.Join(h.uploadDir, oldFilename)
		if err := os.Remove(oldFilePath); err != nil {
			fmt.Printf("删除旧头像失败: %v\n", err)
		}
	}

	baseURL := h.cfg.Server.PublicURL
	if baseURL == "" {
		baseURL = fmt.Sprintf("http://%s:%s", h.cfg.Server.Host, h.cfg.Server.Port)
	}
	fileURL := fmt.Sprintf("%s/uploads/%s", baseURL, filename)

	db.DB.Model(&model.User{}).Where("user_id = ?", userID).Update("avatar", fileURL)

	c.JSON(200, model.Success(map[string]string{
		"filename": filename,
		"url":      fileURL,
	}))
}

func (h *Handler) ServeFile(c *gin.Context) {
	filename := c.Param("filename")
	if filename == "" {
		c.JSON(404, model.Error(1004, "文件不存在"))
		return
	}

	filePath := path.Join(h.uploadDir, filename)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(404, model.Error(1004, "文件不存在"))
		return
	}

	c.File(filePath)
}
