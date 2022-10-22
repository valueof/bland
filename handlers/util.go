package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func parseIDFromRequest(r *http.Request) (id int64, err error) {
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

func parseIDFromPath(r *http.Request, base string) (id int64, err error) {
	id, err = strconv.ParseInt(strings.Trim(strings.TrimPrefix(r.URL.Path, base), "/"), 10, 64)
	return
}
