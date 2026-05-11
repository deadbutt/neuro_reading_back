package utils

import (
	"fmt"
	"neuro-reading/model"

	"github.com/gin-gonic/gin"
)

func SendError(c *gin.Context, code int, errCode int, message string) {
	c.JSON(code, model.Error(errCode, message))
}

func SendSuccess(c *gin.Context, data interface{}) {
	c.JSON(200, model.Success(data))
}

func SendPageSuccess(c *gin.Context, data interface{}, total int64, page, pageSize int) {
	c.JSON(200, model.PageSuccess(data, total, page, pageSize))
}

func HandleDBError(c *gin.Context, err error, notFoundMsg string) bool {
	if err != nil {
		if err.Error() == "record not found" {
			SendError(c, 404, 1004, notFoundMsg)
		} else {
			SendError(c, 500, 1005, "服务器内部错误")
		}
		return true
	}
	return false
}

func LogError(format string, args ...interface{}) {
	fmt.Printf("[ERROR] "+format+"\n", args...)
}

func LogInfo(format string, args ...interface{}) {
	fmt.Printf("[INFO] "+format+"\n", args...)
}
