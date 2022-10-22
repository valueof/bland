package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/valueof/bland/data"
	"golang.org/x/net/html"
)

func registerApiHandlers(r *http.ServeMux) {
	r.HandleFunc("/api/mark-read", markAsRead)
	r.HandleFunc("/api/delete-bookmark", deleteBookmark)
	r.HandleFunc("/api/fetch-metadata", fetchMetadata)
}

func markAsRead(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDFromRequest(r)
	if err != nil {
		fmt.Printf("markAsRead: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	tx, err := data.BeginTx(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := tx.MarkAsRead(id); err != nil {
		fmt.Println(err)
		tx.Rollback()
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func deleteBookmark(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDFromRequest(r)
	if err != nil {
		fmt.Printf("deleteBookmark: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	tx, err := data.BeginTx(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := tx.DeleteBookmark(id); err != nil {
		fmt.Println(err)
		tx.Rollback()
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type FetchMetadataResult struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

var SPACE_RE *regexp.Regexp = regexp.MustCompile(`\s+`)

func attr(n *html.Node, name string) string {
	for _, a := range n.Attr {
		if a.Key == name {
			return a.Val
		}
	}
	return ""
}

func fetchMetadata(w http.ResponseWriter, r *http.Request) {
	d := FetchMetadataResult{}
	defer func() {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(d)
	}()

	if r.Method != "GET" {
		fmt.Printf("wrong request method: expected GET, got %s\n", r.Method)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	u := r.URL.Query().Get("u")
	u = strings.TrimSpace(u)
	if u == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	d.URL = u
	resp, err := http.Get(u)
	if err != nil {
		fmt.Printf("http.Get returned an error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		w.WriteHeader(resp.StatusCode)
		return
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		fmt.Printf("html.Parse returned an error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	title := ""
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			// Parse <meta> tags
			if n.Data == "meta" {
				name := ""
				if attr(n, "name") != "" {
					name = attr(n, "name")
				} else if attr(n, "property") != "" {
					name = attr(n, "property")
				}

				switch name {
				case "description", "twitter:description", "og:description":
					d.Description = attr(n, "content")
				case "twitter:title", "og:title":
					d.Title = attr(n, "content")
				}
			}

			// Parse <title>
			if n.Data == "title" {
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					if c.Type == html.TextNode {
						title += c.Data
					}
				}

				title = SPACE_RE.ReplaceAllString(title, " ")
				title = strings.TrimSpace(title)
				// TODO(anton): strip og:site_name
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}

	f(doc)

	if d.Title == "" && title != "" {
		d.Title = title
	}
}
