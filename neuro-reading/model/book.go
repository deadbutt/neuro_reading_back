package model

type Book struct {
	ID           uint64 `gorm:"primaryKey;autoIncrement" json:"-"`
	BookID       string `gorm:"uniqueIndex;size:32" json:"bookId"`
	Title        string `gorm:"size:128" json:"title"`
	AuthorID     string `gorm:"size:32;index" json:"-"`
	Cover        string `gorm:"size:255" json:"cover"`
	Description  string `gorm:"type:text" json:"description"`
	HotText      string `gorm:"size:64" json:"hotText"`
	WordCount    int64  `json:"wordCount"`
	ChapterCount int    `json:"chapterCount"`
	Status       string `gorm:"size:16" json:"status"`
	Tags         string `gorm:"size:255" json:"-"`
	LastUpdateTime string `gorm:"size:32" json:"lastUpdateTime"`
	IsVip        bool   `gorm:"default:false" json:"isVip"`
	Rating       float64 `gorm:"default:0" json:"rating"`
	RatingCount  int    `gorm:"default:0" json:"ratingCount"`
	CommentCount int    `gorm:"default:0" json:"commentCount"`
}

type BookResponse struct {
	BookID         string         `json:"bookId"`
	Title          string         `json:"title"`
	Author         AuthorResponse `json:"author"`
	Cover          string         `json:"cover"`
	Description    string         `json:"description"`
	HotText        string         `json:"hotText"`
	WordCount      int64          `json:"wordCount"`
	ChapterCount   int            `json:"chapterCount"`
	Status         string         `json:"status"`
	Tags           []string       `json:"tags"`
	LastUpdateTime string         `json:"lastUpdateTime"`
	IsVip          bool           `json:"isVip"`
}

type BookDetailResponse struct {
	BookID         string            `json:"bookId"`
	Title          string            `json:"title"`
	Author         AuthorResponse    `json:"author"`
	Cover          string            `json:"cover"`
	Description    string            `json:"description"`
	HotText        string            `json:"hotText"`
	WordCount      int64             `json:"wordCount"`
	ChapterCount   int               `json:"chapterCount"`
	Status         string            `json:"status"`
	Tags           []string          `json:"tags"`
	LastUpdateTime string            `json:"lastUpdateTime"`
	IsVip          bool              `json:"isVip"`
	Rating         float64           `json:"rating"`
	RatingCount    int               `json:"ratingCount"`
	CommentCount   int               `json:"commentCount"`
	IsInBookshelf  bool              `json:"isInBookshelf"`
	Chapters       []ChapterResponse `json:"chapters"`
}

type Chapter struct {
	ID         uint64 `gorm:"primaryKey;autoIncrement" json:"-"`
	ChapterID  string `gorm:"uniqueIndex;size:32" json:"chapterId"`
	BookID     string `gorm:"size:32;index" json:"bookId"`
	Title      string `gorm:"size:128" json:"title"`
	Index      int    `json:"index"`
	WordCount  int    `json:"wordCount"`
	IsVip      bool   `gorm:"default:false" json:"isVip"`
	IsTrial    bool   `gorm:"default:false" json:"isTrial"`
	UpdateTime string `gorm:"size:32" json:"updateTime"`
	Content    string `gorm:"type:text" json:"-"`
}

type ChapterResponse struct {
	ChapterID  string `json:"chapterId"`
	BookID     string `json:"bookId"`
	Title      string `json:"title"`
	Index      int    `json:"index"`
	WordCount  int    `json:"wordCount"`
	IsVip      bool   `json:"isVip"`
	IsTrial    bool   `json:"isTrial"`
	UpdateTime string `json:"updateTime"`
}

type ChapterContentResponse struct {
	ChapterID         string         `json:"chapterId"`
	BookID            string         `json:"bookId"`
	Title             string         `json:"title"`
	Content           string         `json:"content"`
	Paragraphs        []string       `json:"paragraphs,omitempty"`
	PrevChapterID     *int           `json:"prevChapterId"`
	NextChapterID     *int           `json:"nextChapterId"`
	ParagraphComments map[string]int `json:"paragraphComments"`
}