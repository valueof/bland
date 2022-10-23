package data

import (
	"database/sql"
	"fmt"
)

func fetchBookmarks(q string, args ...any) (bookmarks []Bookmark, err error) {
	rows, err := db.Query(q, args...)
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
			&b.Tags,
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

func fetchTags(where string, args ...any) (tags []Tag, err error) {
	q := fmt.Sprintf(`
	select
		t.id,
		t.name,
		t.is_author,
		count(tb.tag_id) as num_entries
	from tags t
	join tags_bookmarks tb on tb.tag_id = t.id
	join bookmarks b on tb.bookmark_id = b.id
	where %s and b.deleted_at = 0
	group by t.id, t.name, t.is_author;
	`, where)

	rows, err := db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var t Tag
		err = rows.Scan(&t.ID, &t.Name, &t.IsAuthor, &t.NumEntries)
		if err != nil {
			return nil, err
		}
		tags = append(tags, t)
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
		created_at,
		updated_at,
		deleted_at,
		read_at
	from bookmarks
	where deleted_at = 0
	order by created_at desc;
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
		created_at,
		updated_at,
		deleted_at,
		read_at
	from bookmarks
	where read_at = 0 and deleted_at = 0
	order by created_at desc;
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
		created_at,
		updated_at,
		deleted_at,
		read_at
	from bookmarks
	where shortcut <> "" and deleted_at = 0
	order by created_at desc;
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
		created_at,
		updated_at,
		deleted_at,
		read_at
	from bookmarks
	where id = ? and deleted_at = 0
	limit 1
	`
	b := Bookmark{}
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

	return &b, nil
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
		b.created_at,
		b.updated_at,
		b.deleted_at,
		b.read_at
	from bookmarks b
	join tags_bookmarks tb on tb.bookmark_id = b.id
	join tags t on t.id = tb.tag_id
	where t.name = ? and b.deleted_at = 0
	order by b.created_at desc;
	`

	return fetchBookmarks(q, name)
}

func GetShortcutURL(name string) (string, bool) {
	q := `
	select url
	from bookmarks
	where
		shortcut = ? and
		deleted_at = 0
	order by created_at desc
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
	w := `t.is_author = 0`
	return fetchTags(w)
}

func FetchAllAuthors() (tags []Tag, err error) {
	w := `t.is_author = 1`
	return fetchTags(w)
}
