package utils

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"neuro-reading/model"
	"strings"
)

type docxDocument struct {
	XMLName xml.Name `xml:"document"`
	Body    struct {
		Paragraphs []struct {
			Runs []struct {
				Text string `xml:"t"`
			} `xml:"r"`
		} `xml:"p"`
	} `xml:"body"`
}

func ParseChapters(text string, ext string) []model.ChapterMeta {
	var chapters []model.ChapterMeta
	lines := strings.Split(text, "\n")

	var currentTitle string
	var currentContent []string
	var chapterIndex int

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		isChapterTitle := false
		if ext == ".md" {
			isChapterTitle = strings.HasPrefix(line, "# ") || strings.HasPrefix(line, "## ")
			if isChapterTitle {
				line = strings.TrimPrefix(line, "# ")
				line = strings.TrimPrefix(line, "## ")
			}
		} else {
			if strings.HasPrefix(line, "第") && (strings.Contains(line, "章") || strings.Contains(line, "节") || strings.Contains(line, "序")) {
				isChapterTitle = true
			}
		}

		if isChapterTitle {
			if currentTitle != "" && len(currentContent) > 0 {
				chapters = append(chapters, model.ChapterMeta{
					Index:     chapterIndex,
					ChapterID: GenerateChapterID(),
					Title:     currentTitle,
					WordCount: 0,
					Content:   strings.Join(currentContent, "\n\n"),
				})
				chapterIndex++
			}
			currentTitle = line
			currentContent = nil
		} else {
			currentContent = append(currentContent, line)
		}
	}

	if currentTitle != "" && len(currentContent) > 0 {
		chapters = append(chapters, model.ChapterMeta{
			Index:     chapterIndex,
			ChapterID: GenerateChapterID(),
			Title:     currentTitle,
			WordCount: 0,
			Content:   strings.Join(currentContent, "\n\n"),
		})
	}

	return chapters
}

func ParseDocx(data []byte) (string, error) {
	reader, err := zip.NewReader(strings.NewReader(string(data)), int64(len(data)))
	if err != nil {
		return "", err
	}

	var documentFile *zip.File
	for _, f := range reader.File {
		if f.Name == "word/document.xml" {
			documentFile = f
			break
		}
	}

	if documentFile == nil {
		return "", fmt.Errorf("invalid docx file: document.xml not found")
	}

	rc, err := documentFile.Open()
	if err != nil {
		return "", err
	}
	defer rc.Close()

	content, err := io.ReadAll(rc)
	if err != nil {
		return "", err
	}

	var doc docxDocument
	if err := xml.Unmarshal(content, &doc); err != nil {
		return "", err
	}

	var paragraphs []string
	for _, p := range doc.Body.Paragraphs {
		var text string
		for _, r := range p.Runs {
			text += r.Text
		}
		if text != "" {
			paragraphs = append(paragraphs, text)
		}
	}

	return strings.Join(paragraphs, "\n\n"), nil
}
