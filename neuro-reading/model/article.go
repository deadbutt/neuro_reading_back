package model

type ArticleIndex struct {
	ArticleID      string   `json:"articleId"`
	Title          string   `json:"title"`
	Author         string   `json:"author"`
	Summary        string   `json:"summary"`
	Cover          string   `json:"cover,omitempty"`
	WordCount      int      `json:"wordCount"`
	ChapterCount   int      `json:"chapterCount"`
	Tags           []string `json:"tags,omitempty"`
	LastUpdateTime string   `json:"lastUpdateTime"`
}

type ArticleMeta struct {
	ArticleID      string        `json:"articleId"`
	Title          string        `json:"title"`
	Author         string        `json:"author"`
	Summary        string        `json:"summary"`
	Cover          string        `json:"cover,omitempty"`
	Tags           []string      `json:"tags,omitempty"`
	WordCount      int           `json:"wordCount"`
	ChapterCount   int           `json:"chapterCount"`
	Status         string        `json:"status"`
	PublishTime    string        `json:"publishTime"`
	LastUpdateTime string        `json:"lastUpdateTime"`
	Chapters       []ChapterMeta `json:"chapters"`
}

type ChapterMeta struct {
	Index     int    `json:"index"`
	ChapterID string `json:"chapterId"`
	Title     string `json:"title"`
	WordCount int    `json:"wordCount"`
	Content   string `json:"-"`
}
