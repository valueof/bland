package lib

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type key int

const REQUEST_ID_KEY key = 0
const REQUEST_ENV_KEY key = 1

func GetRequestID(ctx context.Context) string {
	id, ok := ctx.Value(REQUEST_ID_KEY).(string)
	if !ok {
		return "unknown"
	}
	return id
}

func IsDev(ctx context.Context) bool {
	env, ok := ctx.Value(REQUEST_ENV_KEY).(string)
	if !ok {
		return false
	}
	return env == "dev"
}

func GetLogger(ctx context.Context) *log.Logger {
	id := GetRequestID(ctx)
	if IsDev(ctx) {
		return log.New(os.Stdout, "", log.Lshortfile)
	}
	return log.New(os.Stdout, fmt.Sprintf("[%s]", id), log.LstdFlags)
}

func RenderTemplate(w http.ResponseWriter, r *http.Request, name string, data any) {
	ctx := r.Context()
	logger := GetLogger(ctx)

	templates := []string{filepath.Join("templates", "base.html")}
	templates = append(templates, filepath.Join("templates", name))
	t, err := template.New("base.html").ParseFiles(templates...)
	if err != nil {
		logger.Printf("ParseFiles(templates/%s): %v", name, err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Internal Server Error")
		return
	}

	err = t.Execute(w, data)
	if err != nil {
		logger.Printf("Execute(): %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Internal Server Error")
	}
}
