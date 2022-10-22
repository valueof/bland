package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/valueof/bland/data"
	"github.com/valueof/bland/lib"
)

func RegisterHandlers(r *http.ServeMux) {
	r.HandleFunc("/", index)
	r.HandleFunc("/unread/", unread)
	r.HandleFunc("/shortcuts/", shortcuts)
	r.HandleFunc("/tags/", tags)
	r.HandleFunc("/authors/", authors)
	r.HandleFunc("/add/", addURL)
	r.HandleFunc("/edit/", editURL)

	registerApiHandlers(r)
}

type withBookmarks struct {
	Bookmarks *[]data.Bookmark
}

type withTags struct {
	Tags *[]data.Tag
}

func index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		if url, ok := data.GetShortcutURL(strings.Trim(r.URL.Path, "/")); ok {
			http.Redirect(w, r, url, http.StatusSeeOther)
			return
		}

		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintln(w, "404 Not Found")
		return
	}

	bookmarks, err := data.FetchAllBookmarks()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, err)
		return
	}

	lib.RenderTemplate(w, r, "index.html", lib.TemplateData{
		Title: "bland: all",
		Data: withBookmarks{
			Bookmarks: &bookmarks,
		},
	})
}

func unread(w http.ResponseWriter, r *http.Request) {
	bookmarks, err := data.FetchUnreadBookmarks()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, err)
		return
	}

	lib.RenderTemplate(w, r, "index.html", lib.TemplateData{
		Title: "bland: unread",
		Data: withBookmarks{
			Bookmarks: &bookmarks,
		},
	})
}

func shortcuts(w http.ResponseWriter, r *http.Request) {
	bookmarks, err := data.FetchShortcuts()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, err)
		return
	}

	lib.RenderTemplate(w, r, "index.html", lib.TemplateData{
		Title: "bland: shortcuts",
		Data: withBookmarks{
			Bookmarks: &bookmarks,
		},
	})
}

func tags(w http.ResponseWriter, r *http.Request) {
	tagName := strings.TrimPrefix(r.URL.Path, "/tags/")
	if tagName == "" {
		tags, err := data.FetchAllTags()
		if err != nil {
			fmt.Printf("fmt.FetchAllTags: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		lib.RenderTemplate(w, r, "tags.html", lib.TemplateData{
			Title: "bland: all tags",
			Data: withTags{
				Tags: &tags,
			},
		})

		return
	}

	tagName = strings.Trim(tagName, "/")
	bookmarks, err := data.FetchBookmarksByTag(tagName)
	if err != nil {
		fmt.Printf("fmt.FetchBookmarksByTag: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	lib.RenderTemplate(w, r, "index.html", lib.TemplateData{
		Title: "bland: " + tagName,
		Data: withBookmarks{
			Bookmarks: &bookmarks,
		},
	})
}

func authors(w http.ResponseWriter, r *http.Request) {
	tagName := strings.TrimPrefix(r.URL.Path, "/authors/")
	if tagName == "" {
		tags, err := data.FetchAllAuthors()
		if err != nil {
			fmt.Printf("fmt.FetchAllAuthors: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		for i, t := range tags {
			t.Name = strings.TrimPrefix(t.Name, "by:")
			tags[i] = t
		}

		lib.RenderTemplate(w, r, "tags.html", lib.TemplateData{
			Title: "bland: all authors",
			Data: withTags{
				Tags: &tags,
			},
		})

		return
	}

	tagName = strings.Trim(tagName, "/")
	bookmarks, err := data.FetchBookmarksByTag("by:" + tagName)
	if err != nil {
		fmt.Printf("fmt.FetchBookmarksByTag: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	lib.RenderTemplate(w, r, "index.html", lib.TemplateData{
		Title: "bland: by " + tagName,
		Data: withBookmarks{
			Bookmarks: &bookmarks,
		},
	})
}

func addURL(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		lib.RenderTemplate(w, r, "form.html", lib.TemplateData{
			Title: "bland: add url",
			Data: &data.Bookmark{
				ReadAt: 1, // For the 'add url' form, the “to read” checkbox should be unchecked by default
			},
		})
		return
	}

	// TODO(anton): Add server-side checking for required fields
	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Println(err)
			return
		}

		tx, err := data.BeginTx(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		b := data.BookmarkFromRequest(r)
		if _, err := tx.AddBookmark(*b); err != nil {
			fmt.Println(err)
			tx.Rollback()
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := tx.Commit(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	fmt.Printf("wrong request method: expected GET/POST, got %s\n", r.Method)
	w.WriteHeader(http.StatusBadRequest)
}

func editURL(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		id, err := parseIDFromPath(r, "/edit/")
		if err != nil {
			fmt.Printf("editURL: %v\n", err)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		b, err := data.FetchBookmarkByID(id)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		lib.RenderTemplate(w, r, "form.html", lib.TemplateData{
			Title: "bland: edit url",
			Data:  b,
		})
		return
	}

	if r.Method == "POST" {
		id, err := parseIDFromPath(r, "/edit/")
		if err != nil {
			fmt.Printf("editURL: %v\n", err)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		b := data.BookmarkFromRequest(r)
		b.ID = id

		tx, err := data.BeginTx(r.Context())
		if err != nil {
			fmt.Printf("data.BeginTx: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = tx.UpdateBookmark(*b)
		if err != nil {
			fmt.Printf("tx.UpdateBookmark: %v\n", err)
			tx.Rollback()
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := tx.Commit(); err != nil {
			fmt.Printf("tx.Commit: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/#bookmark-%d", b.ID), http.StatusSeeOther)
		return
	}

	fmt.Printf("wrong request method: expected GET/POST, got %s\n", r.Method)
	w.WriteHeader(http.StatusBadRequest)
}
