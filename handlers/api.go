package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/valueof/bland/models"
)

func registerApiHandlers(r *http.ServeMux) {
	r.HandleFunc("/api/mark-read", markAsRead)
	r.HandleFunc("/api/delete-bookmark", deleteBookmark)
}

func parseIDFromRequest(w http.ResponseWriter, r *http.Request) (id int64, err error) {
	if r.Method != "POST" {
		return 0, fmt.Errorf("wrong request method: expected POST, got %s", r.Method)
	}

	body := new(strings.Builder)
	_, err = io.Copy(body, r.Body)
	if err != nil {
		return 0, err
	}

	id, err = strconv.ParseInt(body.String(), 10, 64)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func markAsRead(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDFromRequest(w, r)
	if err != nil {
		fmt.Printf("markAsRead: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	tx, err := models.BeginTx(r.Context())
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
	id, err := parseIDFromRequest(w, r)
	if err != nil {
		fmt.Printf("deleteBookmark: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	tx, err := models.BeginTx(r.Context())
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

	w.WriteHeader((http.StatusOK))
}
