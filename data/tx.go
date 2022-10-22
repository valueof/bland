package data

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

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

func (tx *Tx) AddBookmark(b Bookmark) (id int64, err error) {
	tags_map := map[string]int64{}
	for _, name := range b.ParseTagsFunc(func(t string) bool { return true }) {
		id, err := tx.AddTag(name)
		if err != nil {
			fmt.Printf("Tx.AddTag: %v\n", err)
			return 0, err
		}

		tags_map[name] = id
	}

	now := time.Now().Unix()
	b.CreatedAt = now
	b.UpdatedAt = now
	b.DeletedAt = 0

	q1 := `
	insert into bookmarks (url, title, shortcut, description, tags, created_at, updated_at, read_at, deleted_at)
	values(?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
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

func (tx *Tx) UpdateBookmark(data Bookmark) (err error) {
	b, err := FetchBookmarkByID(data.ID)
	if err != nil {
		return err
	}

	tags_map := map[string]int64{}
	if data.Tags != b.Tags {
		for _, name := range data.ParseTagsFunc(func(t string) bool { return true }) {
			id, err := tx.AddTag(name)
			if err != nil {
				fmt.Printf("Tx.AddTag: %v\n", err)
				return err
			}

			tags_map[name] = id
		}
	}
	fmt.Println(tags_map)

	now := time.Now().Unix()

	fmt.Println(tags_map)
	b.URL = data.URL
	b.Title = data.Title
	b.Shortcut = data.Shortcut
	b.Description = data.Description
	b.Tags = data.Tags
	b.UpdatedAt = now

	if b.ReadAt == 0 && data.ReadAt != 0 {
		b.ReadAt = data.ReadAt
	} else if b.ReadAt != 0 && data.ReadAt == 0 {
		b.ReadAt = 0
	}

	q1 := `
	update bookmarks set
		url = ?,
		title = ?,
		shortcut = ?,
		description = ?,
		tags = ?,
		updated_at = ?,
		read_at = ?
	where id = ?
	`
	_, err = tx.sqlTx.Exec(q1, b.URL, b.Title, b.Shortcut, b.Description, b.Tags, b.UpdatedAt, b.ReadAt, b.ID)
	if err != nil {
		return err
	}

	q2 := `delete from tags_bookmarks where bookmark_id = ?`
	_, err = tx.sqlTx.Exec(q2, b.ID)
	if err != nil {
		return err
	}

	q3 := `insert into tags_bookmarks (bookmark_id, tag_id) values (?, ?)`
	for _, tag_id := range tags_map {
		if _, err = tx.sqlTx.Exec(q3, b.ID, tag_id); err != nil {
			return err
		}
	}

	return nil
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
	_, err = tx.sqlTx.Exec(`update bookmarks set read_at = ? where id = ?`,
		time.Now().Unix(), id)
	return
}

func (tx *Tx) DeleteBookmark(id int64) (err error) {
	_, err = tx.sqlTx.Exec(`update bookmarks set deleted_at = ? where id = ?`,
		time.Now().Unix(), id)
	return
}
