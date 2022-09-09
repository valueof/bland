package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/valueof/bland/lib"
	"github.com/valueof/bland/models"
)

func RegisterHandlers(r *http.ServeMux) {
	r.HandleFunc("/", index)
	r.HandleFunc("/unread", unread)
	r.HandleFunc("/shortcuts", shortcuts)
	r.HandleFunc("/add", addURL)
}

type withBookmarks struct {
	Bookmarks *[]models.Bookmark
}

func index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		if url, ok := models.GetShortcutURL(strings.Trim(r.URL.Path, "/")); ok {
			http.Redirect(w, r, url, http.StatusSeeOther)
			return
		}

		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintln(w, "404 Not Found")
		return
	}

	bookmarks, err := models.FetchAllBookmarks()
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
	bookmarks, err := models.FetchUnreadBookmarks()

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
	bookmarks, err := models.FetchShortcuts()

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

func addURL(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		lib.RenderTemplate(w, r, "add.html", lib.TemplateData{
			Title: "bland: add url",
		})
		return
	}

	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Println(err)
			return
		}

		b := models.Bookmark{
			URL:         r.FormValue("url"),
			Title:       r.FormValue("title"),
			Description: r.FormValue("description"),
			Shortcut:    r.FormValue("shortcut"),
			Tags:        []string{},
			Authors:     []string{},
			ToRead:      r.FormValue("toread") != "",
		}

		tags := strings.TrimSpace(r.FormValue("tags"))
		if len(tags) > 0 {
			for _, t := range strings.Split(tags, " ") {
				b.Tags = append(b.Tags, strings.TrimSpace(t))
			}
		}

		tx, err := models.BeginTx(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if _, err := tx.AddBookmark(b); err != nil {
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
	}
}

// func fetchMetadata(w http.ResponseWriter, r *http.Request) {
// meta name=description content
// meta name=twitter:description content
// meta name=og:description content
// meta name=twitter:title content
// meta property=og:title content
// title
//		aux: look for header/h1 tags in the body
// og:site_name content can help with cleaning up title?
// }
