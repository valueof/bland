package lib

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type key int

const REQUEST_ID_KEY key = 0
const REQUEST_ENV_KEY key = 1

var NEWLINE_RE *regexp.Regexp = regexp.MustCompile(`\r?\n`)

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

func addBreaks(unsafe string) template.HTML {
	safe := template.HTMLEscapeString(unsafe)
	safe = NEWLINE_RE.ReplaceAllString(safe, "<br>")
	return template.HTML(safe)
}

type TemplateData struct {
	Data  any
	Path  string
	Title string
	Host  string
}

func RenderTemplate(w http.ResponseWriter, r *http.Request, name string, data TemplateData) {
	ctx := r.Context()
	logger := GetLogger(ctx)

	funcs := template.FuncMap{
		"addBreaks": addBreaks,
		"toLower":   strings.ToLower,
		"hasPrefix": strings.HasPrefix,
	}

	templates := []string{filepath.Join("templates", "base.html")}
	templates = append(templates, filepath.Join("templates", name))
	t, err := template.New("base.html").Funcs(funcs).ParseFiles(templates...)
	if err != nil {
		logger.Printf("ParseFiles(templates/%s): %v", name, err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Internal Server Error")
		return
	}

	data.Host = r.Host
	data.Path = r.URL.Path

	err = t.Execute(w, data)
	if err != nil {
		logger.Printf("Execute(): %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Internal Server Error")
	}
}
