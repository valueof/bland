package handlers

import (
	"fmt"
	"net/http"

	"github.com/valueof/bland/lib"
	"github.com/valueof/bland/models"
)

func RegisterHandlers(r *http.ServeMux) {
	r.HandleFunc("/", index)
}

type RenderData struct {
	Bookmarks []models.Bookmark
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

	lib.RenderTemplate(w, r, "index.html", RenderData{Bookmarks: bookmarks})
}
