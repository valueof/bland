package models

import (
	"database/sql"
	"os"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

var db *sql.DB

func Connect() (err error) {
	cfg := mysql.Config{
		User:      os.Getenv("DBUSER"),
		Passwd:    os.Getenv("DBPASS"),
		Net:       "tcp",
		Addr:      os.Getenv("DBADDR"),
		DBName:    os.Getenv("DBNAME"),
		ParseTime: true,
	}

	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return
	}

	return db.Ping()
}

// Publicly available structs
type Bookmark struct {
	ID          int64      `json:"id"`
	URL         string     `json:"url"`
	Title       string     `json:"title"`
	Shortcut    string     `json:"shortcut,omitempty"`
	Description string     `json:"description,omitempty"`
	ToRead      bool       `json:"toRead"`
	CreatedAt   *time.Time `json:"createdAt,omitempty"`
	UpdatedAt   *time.Time `json:"updatedAt,omitempty"`
	ReadAt      *time.Time `json:"readAt,omitempty"`
	Tags        []string   `json:"tags"`
	Authors     []string   `json:"authors"`
}

// Structs representing DB tables
type dbBookmark struct {
	ID          int64
	URL         string
	Title       string
	Shortcut    string
	Description string
	Tags        string
	CreatedAt   sql.NullTime
	UpdatedAt   sql.NullTime
	DeletedAt   sql.NullTime
	ReadAt      sql.NullTime
}

func toBookmark(in dbBookmark) (out Bookmark) {
	out = Bookmark{
		ID:          in.ID,
		URL:         in.URL,
		Title:       in.Title,
		Shortcut:    in.Shortcut,
		Description: in.Description,
		ToRead:      !in.ReadAt.Valid,
		Tags:        []string{},
		Authors:     []string{},
	}

	if in.CreatedAt.Valid {
		out.CreatedAt = &in.CreatedAt.Time
	}

	if in.UpdatedAt.Valid {
		out.UpdatedAt = &in.UpdatedAt.Time
	}

	if in.ReadAt.Valid {
		out.ReadAt = &in.ReadAt.Time
	}

	if len(strings.TrimSpace(in.Tags)) > 0 {
		tags := strings.Split(in.Tags, " ")
		for _, t := range tags {
			t = strings.TrimSpace(t)
			if strings.HasPrefix(t, "by:") {
				out.Authors = append(out.Authors, strings.TrimPrefix(t, "by:"))
			} else {
				out.Tags = append(out.Tags, t)
			}
		}
	}

	return
}

func fromBookmark(in Bookmark) (out dbBookmark) {
	now := time.Now()
	out = dbBookmark{
		ID:          in.ID,
		URL:         in.URL,
		Title:       in.Title,
		Shortcut:    in.Shortcut,
		Description: in.Description,
		CreatedAt:   sql.NullTime{Valid: true, Time: now},
		UpdatedAt:   sql.NullTime{Valid: true, Time: now},
		ReadAt:      sql.NullTime{Valid: true, Time: now},
	}

	if in.ToRead {
		out.ReadAt = sql.NullTime{Valid: false, Time: now}
	}

	tags := []string{}
	tags = append(tags, in.Tags...)

	for _, t := range in.Authors {
		tags = append(tags, "by:"+t)
	}

	out.Tags = strings.Join(tags, " ")

	return
}

func fetchBookmarks(q string) (bookmarks []Bookmark, err error) {
	rows, err := db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var b dbBookmark
		err = rows.Scan(
			&b.ID,
			&b.URL,
			&b.Title,
			&b.Shortcut,
			&b.Description,
			&b.Tags,
			&b.CreatedAt,
			&b.UpdatedAt,
			&b.DeletedAt,
			&b.ReadAt)

		if err != nil {
			return nil, err
		}

		bookmarks = append(bookmarks, toBookmark(b))
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return
}

func FetchAllBookmarks() (bookmarks []Bookmark, err error) {
	q := `
select
	id,
	url,
	title,
	shortcut,
	description,
	tags,
	createdAt,
	updatedAt,
	deletedAt,
	readAt
from bookmarks
order by createdAt desc;
`

	return fetchBookmarks(q)
}

func FetchUnreadBookmarks() (bookmarks []Bookmark, err error) {
	q := `
select
	id,
	url,
	title,
	shortcut,
	description,
	tags,
	createdAt,
	updatedAt,
	deletedAt,
	readAt
from bookmarks
where readAt is null
order by createdAt desc;
`

	return fetchBookmarks(q)
}

func FetchShortcuts() (bookmarks []Bookmark, err error) {
	q := `
select
	id,
	url,
	title,
	shortcut,
	description,
	tags,
	createdAt,
	updatedAt,
	deletedAt,
	readAt
from bookmarks
where shortcut <> ""
order by createdAt desc;
`

	return fetchBookmarks(q)
}

func AddBookmark(data Bookmark) (id int64, err error) {
	query := `
insert into bookmarks (url, title, shortcut, description, tags, createdAt, updatedAt, readAt) values(?, ?, ?, ?, ?, ?, ?, ?)`

	b := fromBookmark(data)
	result, err := db.Exec(query, b.URL, b.Title, b.Shortcut, b.Description, b.Tags, b.CreatedAt, b.UpdatedAt, b.ReadAt)
	if err != nil {
		return 0, err
	}

	id, err = result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return
}
