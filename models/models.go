package models

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func Connect(fp string) (err error) {
	db, err = sql.Open("sqlite3", fp)
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

type Tag struct {
	Name       string `json:"name"`
	IsAuthor   bool   `json:"is_author"`
	NumEntries int64  `json:"num_entries"`
}

// Structs representing DB tables
type dbBookmark struct {
	ID          int64
	URL         string
	Title       string
	Shortcut    string
	Description string
	Tags        string
	CreatedAt   int64
	UpdatedAt   int64
	DeletedAt   int64
	ReadAt      int64
}

type dbTag struct {
	ID         int64
	Name       string
	IsAuthor   bool
	NumEntries int64
}

func toBookmark(in dbBookmark) (out Bookmark) {
	out = Bookmark{
		ID:          in.ID,
		URL:         in.URL,
		Title:       in.Title,
		Shortcut:    in.Shortcut,
		Description: in.Description,
		ToRead:      in.ReadAt == 0,
		Tags:        []string{},
		Authors:     []string{},
	}

	if in.CreatedAt != 0 {
		tm := time.Unix(in.CreatedAt, 0)
		out.CreatedAt = &tm
	}

	if in.UpdatedAt != 0 {
		tm := time.Unix(in.UpdatedAt, 0)
		out.UpdatedAt = &tm
	}

	if in.ReadAt != 0 {
		tm := time.Unix(in.ReadAt, 0)
		out.ReadAt = &tm
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

func toTag(in dbTag) (out Tag) {
	return Tag{
		Name:     in.Name,
		IsAuthor: in.IsAuthor,
	}
}

func fromBookmark(in Bookmark) (out dbBookmark) {
	now := time.Now().Unix()
	out = dbBookmark{
		ID:          in.ID,
		URL:         in.URL,
		Title:       in.Title,
		Shortcut:    in.Shortcut,
		Description: in.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
		ReadAt:      0,
		DeletedAt:   0,
	}

	if in.ToRead {
		out.ReadAt = now
	}

	out.Tags = strings.Join(in.Tags, " ")
	return
}

func fetchBookmarks(q string, args ...any) (bookmarks []Bookmark, err error) {
	rows, err := db.Query(q, args...)
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

func fetchTags(where string, args ...any) (tags []Tag, err error) {
	q := fmt.Sprintf(`
select
	t.id,
    t.name,
    t.is_author,
    count(tb.tag_id) as num_entries
from tags t
join tags_bookmarks tb on tb.tag_id = t.id
%s
group by t.id, t.name, t.is_author;
`, where)

	rows, err := db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var t dbTag
		err = rows.Scan(&t.ID, &t.Name, &t.IsAuthor, &t.NumEntries)
		if err != nil {
			return nil, err
		}
		tags = append(tags, toTag(t))
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

func FetchBookmarkByID(id int64) (bookmark *Bookmark, err error) {
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
where id = ?
limit 1
`
	b := dbBookmark{}
	row := db.QueryRow(q, id)
	err = row.Scan(
		&b.ID,
		&b.URL,
		&b.Title,
		&b.Shortcut,
		&b.Description,
		&b.Tags,
		&b.CreatedAt,
		&b.UpdatedAt,
		&b.DeletedAt,
		&b.ReadAt,
	)

	if err != nil {
		if err != sql.ErrNoRows {
			fmt.Printf("models.FetchBookmarkByID: %v\n", err)
		}
		return nil, err
	}

	out := toBookmark(b)
	return &out, nil
}

func FetchBookmarksByTag(name string) (bookmarks []Bookmark, err error) {
	q := `
select
	b.id,
	b.url,
	b.title,
	b.shortcut,
	b.description,
	b.tags,
	b.createdAt,
	b.updatedAt,
	b.deletedAt,
	b.readAt
from bookmarks b
join tags_bookmarks tb on tb.bookmark_id = b.id
join tags t on t.id = tb.tag_id
where t.name = ?
order by b.createdAt desc;
`

	return fetchBookmarks(q, name)
}

func GetShortcutURL(name string) (string, bool) {
	q := `
select url
from bookmarks
where
	shortcut = ? and
	deletedAt = 0
order by createdAt desc
limit 1`

	var url string
	if err := db.QueryRow(q, name).Scan(&url); err != nil {
		if err != sql.ErrNoRows {
			// TODO: use logger
			fmt.Println(err)
		}
		return "", false
	}
	return url, true
}

func FetchAllTags() (tags []Tag, err error) {
	w := `where t.is_author = 0`
	return fetchTags(w)
}

func FetchAllAuthors() (tags []Tag, err error) {
	w := `where t.is_author = 1`
	return fetchTags(w)
}

type Tx struct {
	ctx   context.Context
	sqlTx *sql.Tx
}

func BeginTx(ctx context.Context) (tx *Tx, err error) {
	sqlTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		fmt.Printf("BeginTX: %v\n", err)
		return nil, err
	}

	tx = &Tx{
		ctx:   ctx,
		sqlTx: sqlTx,
	}

	return tx, nil
}

func (tx *Tx) Insert(q string, args ...any) (id int64, err error) {
	res, err := tx.sqlTx.Exec(q, args...)
	if err != nil {
		return 0, err
	}

	id, err = res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (tx *Tx) Commit() (err error) {
	err = tx.sqlTx.Commit()
	if err != nil {
		fmt.Printf("Tx.Commit: %v\n", err)
	}
	return
}

func (tx *Tx) Rollback() (err error) {
	err = tx.sqlTx.Rollback()
	if err != nil {
		fmt.Printf("Tx.Rollback: %v\n", err)
	}
	return
}

func (tx *Tx) AddBookmark(data Bookmark) (id int64, err error) {
	tags_map := map[string]int64{}
	for _, name := range data.Tags {
		id, err := tx.AddTag(name)
		if err != nil {
			fmt.Printf("Tx.AddTag: %v\n", err)
			return 0, err
		}

		tags_map[name] = id
	}

	b := fromBookmark(data)
	q1 := `insert into bookmarks (url, title, shortcut, description, tags, createdAt, updatedAt, readAt, deletedAt) values(?, ?, ?, ?, ?, ?, ?, ?, ?)`
	id, err = tx.Insert(q1, b.URL, b.Title, b.Shortcut, b.Description, b.Tags, b.CreatedAt, b.UpdatedAt, b.ReadAt, b.DeletedAt)
	if err != nil {
		return 0, err
	}

	q2 := `insert into tags_bookmarks (bookmark_id, tag_id) values (?, ?)`
	for _, tag_id := range tags_map {
		if _, err = tx.sqlTx.Exec(q2, id, tag_id); err != nil {
			return 0, nil
		}
	}

	return id, nil
}

func (tx *Tx) AddTag(name string) (id int64, err error) {
	q1 := `select id from tags where name = ?`
	if err := tx.sqlTx.QueryRow(q1, name).Scan(&id); err != nil {
		if err != sql.ErrNoRows {
			return 0, err
		}
	} else {
		return id, nil
	}

	q2 := `insert into tags (name, is_author) values(?, ?)`
	result, err := tx.sqlTx.Exec(q2, name, strings.HasPrefix(name, "by:"))
	if err != nil {
		return 0, err
	}

	id, err = result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (tx *Tx) MarkAsRead(id int64) (err error) {
	now := time.Now().Unix()
	q := `update bookmarks set readAt = ? where id = ?`
	_, err = tx.sqlTx.Exec(q, now, id)
	return
}
