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

type TemplateData struct {
	ActiveNav string
	Title     string
	Bookmarks *[]models.Bookmark
}

func index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
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

	lib.RenderTemplate(w, r, "index.html", TemplateData{
		ActiveNav: "/",
		Title:     "bland: all",
		Bookmarks: &bookmarks,
	})
}

func unread(w http.ResponseWriter, r *http.Request) {
	bookmarks, err := models.FetchUnreadBookmarks()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, err)
		return
	}

	lib.RenderTemplate(w, r, "index.html", TemplateData{
		ActiveNav: "/unread",
		Title:     "bland: unread",
		Bookmarks: &bookmarks,
	})
}

func shortcuts(w http.ResponseWriter, r *http.Request) {
	bookmarks, err := models.FetchShortcuts()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, err)
		return
	}

	lib.RenderTemplate(w, r, "index.html", TemplateData{
		ActiveNav: "/shortcuts",
		Title:     "bland: shortcuts",
		Bookmarks: &bookmarks,
	})
}

func addURL(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		lib.RenderTemplate(w, r, "add.html", TemplateData{
			ActiveNav: "/add",
			Title:     "bland: add url",
		})
		return
	}

	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Println(w, err)
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

		if _, err := models.AddBookmark(b); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Println(w, err)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
