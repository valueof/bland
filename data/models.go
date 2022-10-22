package data

import (
	"net/http"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Bookmark struct {
	ID          int64  `json:"id"`
	URL         string `json:"url"`
	Title       string `json:"title"`
	Shortcut    string `json:"shortcut"`
	Description string `json:"description"`
	Tags        string `json:"tags"`
	CreatedAt   int64  `json:"createdAt"`
	UpdatedAt   int64  `json:"updatedAt"`
	DeletedAt   int64  `json:"deletedAt"`
	ReadAt      int64  `json:"readAt"`
}

func BookmarkFromRequest(r *http.Request) (b *Bookmark) {
	b = &Bookmark{
		URL:         r.FormValue("url"),
		Title:       r.FormValue("title"),
		Description: r.FormValue("description"),
		Shortcut:    r.FormValue("shortcut"),
		ReadAt:      0,
	}

	if r.FormValue("toread") == "" {
		b.ReadAt = time.Now().Unix()
	}

	tags := []string{}
	for _, t := range strings.Split(strings.TrimSpace(r.FormValue("tags")), " ") {
		tags = append(tags, strings.TrimSpace(t))
	}

	b.Tags = strings.Join(tags, " ")
	return
}

func (b *Bookmark) ToRead() bool {
	return b.ReadAt == 0
}

func (b *Bookmark) TimeCreated() *time.Time {
	tm := time.Unix(b.CreatedAt, 0)
	return &tm
}

func (b *Bookmark) TimeUpdated() *time.Time {
	tm := time.Unix(b.UpdatedAt, 0)
	return &tm
}

func (b *Bookmark) TimeRead() *time.Time {
	tm := time.Unix(b.ReadAt, 0)
	return &tm
}

func (b *Bookmark) ParseTagsFunc(f func(string) bool) (tags []string) {
	tags = []string{}
	if strings.TrimSpace(b.Tags) == "" {
		return
	}

	for _, t := range strings.Split(b.Tags, " ") {
		t = strings.TrimSpace(t)
		if f(t) {
			tags = append(tags, t)
		}
	}

	return
}

func (b *Bookmark) ParseTags() (tags []string) {
	return b.ParseTagsFunc(func(t string) bool {
		return !strings.HasPrefix(t, "by:")
	})
}

func (b *Bookmark) ParseAuthors() (authors []string) {
	authors = b.ParseTagsFunc(func(t string) bool {
		return strings.HasPrefix(t, "by:")
	})

	for i, t := range authors {
		authors[i] = strings.TrimPrefix(t, "by:")
	}

	return
}

type Tag struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	IsAuthor   bool   `json:"is_author"`
	NumEntries int64  `json:"num_entries"`
}
