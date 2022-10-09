package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/valueof/bland/lib"
	"github.com/valueof/bland/models"
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
	Bookmarks *[]models.Bookmark
}

type withTags struct {
	Tags *[]models.Tag
}

type withFormValues struct {
	ID          int64
	URL         string
	Title       string
	Description string
	Shortcut    string
	Tags        string
	ToRead      bool
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

func tags(w http.ResponseWriter, r *http.Request) {
	tagName := strings.TrimPrefix(r.URL.Path, "/tags/")
	if tagName == "" {
		tags, err := models.FetchAllTags()
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
	bookmarks, err := models.FetchBookmarksByTag(tagName)
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
		tags, err := models.FetchAllAuthors()
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
	bookmarks, err := models.FetchBookmarksByTag("by:" + tagName)
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
			Data:  withFormValues{},
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

		b := models.Bookmark{
			URL:         r.FormValue("url"),
			Title:       r.FormValue("title"),
			Description: r.FormValue("description"),
			Shortcut:    r.FormValue("shortcut"),
			Tags:        []string{},
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

func editURL(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		id, err := strconv.Atoi(strings.Trim(strings.TrimPrefix(r.URL.Path, "/edit/"), "/"))
		if err != nil {
			fmt.Printf("editURL: %v\n", err)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		b, err := models.FetchBookmarkByID(int64(id))
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		tags := []string{}
		tags = append(tags, b.Tags...)
		for _, a := range b.Authors {
			tags = append(tags, "by:"+a)
		}

		lib.RenderTemplate(w, r, "form.html", lib.TemplateData{
			Title: "bland: add url",
			Data: withFormValues{
				ID:          b.ID,
				URL:         b.URL,
				Title:       b.Title,
				Description: b.Description,
				Shortcut:    b.Shortcut,
				Tags:        strings.Join(tags, " "),
			},
		})
		return
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
