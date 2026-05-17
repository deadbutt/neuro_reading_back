package utils

import (
	"fmt"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func GenerateUserID() string {
	return fmt.Sprintf("u_%d", time.Now().UnixNano())
}

func GenerateAuthorID() string {
	return fmt.Sprintf("a_%d", time.Now().UnixNano())
}

func GenerateBookID() string {
	return fmt.Sprintf("b_%d", time.Now().UnixNano())
}

func GenerateChapterID() string {
	return fmt.Sprintf("c_%d", time.Now().UnixNano())
}

func GenerateCommentID() string {
	return fmt.Sprintf("cm_%d", time.Now().UnixNano())
}

func GenerateParagraphCommentID() string {
	return fmt.Sprintf("pc_%d", time.Now().UnixNano())
}

func GenerateFeedID() string {
	return fmt.Sprintf("f_%d", time.Now().UnixNano())
}

func GenerateArticleID() string {
	return fmt.Sprintf("ar_%d", time.Now().UnixNano())
}

func GenerateCreatorID() string {
	return fmt.Sprintf("cr_%d", time.Now().UnixNano())
}

func GenerateWorkID() string {
	return fmt.Sprintf("w_%d", time.Now().UnixNano())
}

func GenerateHistoryID() string {
	return fmt.Sprintf("rh_%d", time.Now().UnixNano())
}

func GenerateNotificationID() string {
	return fmt.Sprintf("nt_%d", time.Now().UnixNano())
}

func GenerateCode() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

