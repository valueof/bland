package models

import (
	"database/sql"
	"os"

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

type Bookmark struct {
	ID          int64
	URL         string
	Title       string
	Shortcut    string
	Description string
	CreatedAt   sql.NullTime
	UpdatedAt   sql.NullTime
	DeletedAt   sql.NullTime
	ReadAt      sql.NullTime
}

func FetchAllBookmarks() (bookmarks []Bookmark, err error) {
	query := `
select id, url, title, shortcut, description, createdAt, updatedAt, deletedAt, readAt
from bookmarks`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var b Bookmark
		err = rows.Scan(
			&b.ID,
			&b.URL,
			&b.Title,
			&b.Shortcut,
			&b.Description,
			&b.CreatedAt,
			&b.UpdatedAt,
			&b.DeletedAt,
			&b.ReadAt)

		if err != nil {
			return nil, err
		}
		bookmarks = append(bookmarks, b)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return
}

func AddBookmark(b Bookmark) (id int64, err error) {
	query := `
insert into bookmarks (url, title, shortcut, description, createdAt) values(?, ?, ?, ?, now())`

	result, err := db.Exec(query, b.URL, b.Title, b.Shortcut, b.Description)
	if err != nil {
		return 0, err
	}

	id, err = result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return
}
