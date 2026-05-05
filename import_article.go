package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"
	"time"
)

type ArticleJSON struct {
	Title   string   `json:"title"`
	Date    string   `json:"date"`
	Tags    []string `json:"tags"`
	URL     string   `json:"url"`
	Content string   `json:"content"`
}

type ArticleIndex struct {
	ArticleID      string   `json:"articleId"`
	Title          string   `json:"title"`
	Author         string   `json:"author"`
	Summary        string   `json:"summary"`
	WordCount      int      `json:"wordCount"`
	ChapterCount   int      `json:"chapterCount"`
	Tags           []string `json:"tags,omitempty"`
	LastUpdateTime string   `json:"lastUpdateTime"`
}

type ChapterMeta struct {
	Index     int    `json:"index"`
	ChapterID string `json:"chapterId"`
	Title     string `json:"title"`
	WordCount int    `json:"wordCount"`
}

type ArticleMeta struct {
	ArticleID      string        `json:"articleId"`
	Title          string        `json:"title"`
	Author         string        `json:"author"`
	Summary        string        `json:"summary"`
	Tags           []string      `json:"tags,omitempty"`
	WordCount      int           `json:"wordCount"`
	ChapterCount   int           `json:"chapterCount"`
	Status         string        `json:"status"`
	PublishTime    string        `json:"publishTime"`
	LastUpdateTime string        `json:"lastUpdateTime"`
	Chapters       []ChapterMeta `json:"chapters"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: go run import_article.go <article.json>")
		os.Exit(1)
	}

	jsonPath := os.Args[1]
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		fmt.Printf("读取文件失败: %v\n", err)
		os.Exit(1)
	}

	var article ArticleJSON
	if err := json.Unmarshal(data, &article); err != nil {
		fmt.Printf("解析JSON失败: %v\n", err)
		os.Exit(1)
	}

	articleDir := "./articles"
	if err := os.MkdirAll(articleDir, 0755); err != nil {
		fmt.Printf("创建目录失败: %v\n", err)
		os.Exit(1)
	}

	articleID := fmt.Sprintf("ar_%d", time.Now().UnixNano())
	articlePath := path.Join(articleDir, articleID)
	chaptersDir := path.Join(articlePath, "chapters")
	if err := os.MkdirAll(chaptersDir, 0755); err != nil {
		fmt.Printf("创建目录失败: %v\n", err)
		os.Exit(1)
	}

	now := time.Now().Format("2006-01-02 15:04:05")

	chapters := parseChapters(article.Content, article.Title)
	if len(chapters) == 0 {
		chapters = []ParsedChapter{
			{Title: article.Title, Content: article.Content},
		}
	}

	totalWordCount := 0
	chapterMetas := make([]ChapterMeta, len(chapters))
	for i, ch := range chapters {
		wordCount := len([]rune(ch.Content))
		totalWordCount += wordCount

		chapterMetas[i] = ChapterMeta{
			Index:     i,
			ChapterID: fmt.Sprintf("c_%d_%d", time.Now().UnixNano(), i),
			Title:     ch.Title,
			WordCount: wordCount,
		}

		chapterPath := path.Join(chaptersDir, fmt.Sprintf("%d.txt", i))
		if err := os.WriteFile(chapterPath, []byte(ch.Content), 0644); err != nil {
			fmt.Printf("保存章节失败: %v\n", err)
			os.Exit(1)
		}
	}

	meta := ArticleMeta{
		ArticleID:      articleID,
		Title:          article.Title,
		Author:         "未知作者",
		Summary:        article.Content,
		Tags:           article.Tags,
		WordCount:      totalWordCount,
		ChapterCount:   len(chapters),
		Status:         "published",
		PublishTime:    now,
		LastUpdateTime: now,
		Chapters:       chapterMetas,
	}

	if len(meta.Summary) > 200 {
		meta.Summary = meta.Summary[:200] + "..."
	}

	metaData, _ := json.MarshalIndent(meta, "", "  ")
	if err := os.WriteFile(path.Join(articlePath, "meta.json"), metaData, 0644); err != nil {
		fmt.Printf("保存meta失败: %v\n", err)
		os.Exit(1)
	}

	indexPath := path.Join(articleDir, "index.json")
	var index []ArticleIndex
	if indexData, err := os.ReadFile(indexPath); err == nil {
		json.Unmarshal(indexData, &index)
	}

	index = append([]ArticleIndex{{
		ArticleID:      articleID,
		Title:          article.Title,
		Author:         meta.Author,
		Summary:        meta.Summary,
		WordCount:      totalWordCount,
		ChapterCount:   len(chapters),
		Tags:           article.Tags,
		LastUpdateTime: now,
	}}, index...)

	indexData, _ := json.MarshalIndent(index, "", "  ")
	if err := os.WriteFile(indexPath, indexData, 0644); err != nil {
		fmt.Printf("保存索引失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("导入成功! 文章ID: %s\n", articleID)
	fmt.Printf("标题: %s\n", article.Title)
	fmt.Printf("章节数: %d\n", len(chapters))
	fmt.Printf("总字数: %d\n", totalWordCount)
}

type ParsedChapter struct {
	Title   string
	Content string
}

func parseChapters(content string, articleTitle string) []ParsedChapter {
	var chapters []ParsedChapter

	// 先尝试找【第X章】这种明确的章节标记
	chapterPatterns := []string{
		`(?m)^【第[一二三四五六七八九十百千万零\d]+章[\s:：]*(.+?)】\s*$`,
		`(?m)^第[一二三四五六七八九十百千万零\d]+章[\s:：]*(.+?)$`,
		`(?m)^【第[一二三四五六七八九十百千万零\d]+章】\s*$`,
		`(?m)^第[一二三四五六七八九十百千万零\d]+章\s*$`,
		`(?m)^Chapter\s+\d+[\s:：]*(.+?)$`,
		`(?m)^CHAPTER\s+\d+[\s:：]*(.+?)$`,
	}

	var allMatches [][]int
	var matchTitles []string

	for _, pattern := range chapterPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatchIndex(content, -1)
		for _, match := range matches {
			start := match[0]
			end := match[1]
			var title string
			if len(match) >= 4 && match[2] >= 0 {
				title = strings.TrimSpace(content[match[2]:match[3]])
			}

			if title == "" {
				title = strings.TrimSpace(content[start:end])
				title = strings.Trim(title, "【】")
			}

			if title == "" {
				continue
			}

			allMatches = append(allMatches, match[:2])
			matchTitles = append(matchTitles, title)
		}
	}

	// 如果没找到明确的章节标记，再找【xxx】这种
	if len(allMatches) == 0 {
		bracketPattern := regexp.MustCompile(`(?m)^【(.+?)】\s*$`)
		matches := bracketPattern.FindAllStringSubmatchIndex(content, -1)
		for i, match := range matches {
			title := strings.TrimSpace(content[match[2]:match[3]])

			if title == "" {
				continue
			}

			// 第一个匹配如果是文章标题，作为第一章
			// 后面的匹配如果还是文章标题，则跳过
			if title == articleTitle && i > 0 {
				continue
			}

			allMatches = append(allMatches, match[:2])
			matchTitles = append(matchTitles, title)
		}
	}

	fmt.Printf("找到 %d 个匹配:\n", len(allMatches))
	for i, title := range matchTitles {
		fmt.Printf("  %d: '%s' (位置: %d-%d)\n", i, title, allMatches[i][0], allMatches[i][1])
	}

	if len(allMatches) == 0 {
		return nil
	}

	for i := 0; i < len(allMatches)-1; i++ {
		for j := i + 1; j < len(allMatches); j++ {
			if allMatches[i][0] > allMatches[j][0] {
				allMatches[i], allMatches[j] = allMatches[j], allMatches[i]
				matchTitles[i], matchTitles[j] = matchTitles[j], matchTitles[i]
			}
		}
	}

	for i := 0; i < len(allMatches); i++ {
		end := allMatches[i][1]
		title := matchTitles[i]

		if title == "" {
			title = articleTitle
		}

		var chapterContent string
		if i < len(allMatches)-1 {
			nextStart := allMatches[i+1][0]
			chapterContent = strings.TrimSpace(content[end:nextStart])
		} else {
			chapterContent = strings.TrimSpace(content[end:])
		}

		if chapterContent != "" {
			chapters = append(chapters, ParsedChapter{
				Title:   title,
				Content: chapterContent,
			})
		}
	}

	return chapters
}
