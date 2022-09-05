package handlers

import (
	"fmt"
	"net/http"

	"github.com/valueof/bland/lib"
	"github.com/valueof/bland/models"
)

func RegisterHandlers(r *http.ServeMux) {
	r.HandleFunc("/", index)
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
		Title:     "Bland",
		Bookmarks: &bookmarks,
	})
}

func addURL(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		lib.RenderTemplate(w, r, "add.html", TemplateData{
			ActiveNav: "/add",
			Title:     "Add URL",
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
		}

		if _, err := models.AddBookmark(b); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Println(w, err)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
