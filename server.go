package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/google/uuid"
	"github.com/valueof/bland/data"
	"github.com/valueof/bland/handlers"
	"github.com/valueof/bland/lib"
)

var addr *string
var dev *bool
var db *string
var setup *bool

func init() {
	addr = flag.String("addr", "", "server address")
	db = flag.String("db", "", "db file")
	dev = flag.Bool("dev", false, "dev environment (simplifies logging)")
	setup = flag.Bool("setup", false, "create the db and exit")
}

func tracing(uuid func() string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := r.Header.Get("X-Request-Id")
			if id == "" {
				id = uuid()
			}

			env := "prod"
			if *dev {
				env = "dev"
			}

			ctx := context.WithValue(r.Context(), lib.REQUEST_ID_KEY, id)
			ctx = context.WithValue(ctx, lib.REQUEST_ENV_KEY, env)
			w.Header().Set("X-Request-Id", id)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func logging(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				rid := lib.GetRequestID(r.Context())
				if *dev {
					logger.Println(r.Method, r.URL.Path)
				} else {
					logger.Println(rid, r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags)

	flag.Parse()

	if *setup {
		createDB(*db)
		return
	}

	if *addr == "" {
		flag.Usage()
		return
	}

	if *dev {
		logger.Println("starting in DEV mode")
	} else {
		logger.Println("starting in PROD mode")
	}

	logger.Println("connecting to db")
	err := data.ConnectToDB(*db)
	if err != nil {
		logger.Fatalf("could not connect to db: %v", err)
	}
	logger.Println("connected")

	router := http.NewServeMux()
	handlers.RegisterHandlers(router)

	static := http.FileServer(http.Dir("static"))
	router.Handle("/static/", http.StripPrefix("/static/", static))

	s := &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  15 * time.Second,
		Addr:         *addr,
		Handler:      tracing(uuid.NewString)(logging(logger)(router)),
	}

	done := make(chan bool)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	go func() {
		<-quit
		logger.Println("shutting down")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		s.SetKeepAlivesEnabled(false)

		if err := s.Shutdown(ctx); err != nil {
			logger.Fatalf("could not gracefully shutdown: %v", err)
		}
		close(done)
	}()

	logger.Println("ready at", *addr)
	if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("could not listen on %s: %v", *addr, err)
	}

	<-done
	logger.Println("goodbye")
}
