package main

import (
	"flag"
	"fmt"
	"github.com/lmittmann/tint"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path"
	"time"
)

const version = "1.0.0"

type config struct {
	port int
	env  string
}

// Dependencies lives here for the application
type application struct {
	config config
	logi   *log.Logger
	loge   *log.Logger
}

func main() {

	var cfg config

	// Setting up passed flags
	flag.IntVar(&cfg.port, "port", 8080, "Application port")
	flag.StringVar(&cfg.env, "env", "dev", "Current application environment")
	flag.Parse()

	// Loggers configuration
	logi := log.New(os.Stderr, "INFO\t", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
	loge := log.New(os.Stderr, "ERROR\t", log.LstdFlags|log.Lmicroseconds|log.Llongfile)

	tintHandler := tint.NewHandler(os.Stderr, &tint.Options{
		AddSource: true,
		ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
			if attr.Key == slog.SourceKey {
				src := attr.Value.Any().(*slog.Source)
				return slog.Attr{
					Key:   attr.Key,
					Value: slog.AnyValue(fmt.Sprintf("%s:%d", path.Base(src.File), src.Line)),
				}
			}
			return attr
		},
	})
	slog.SetDefault(slog.New(tintHandler))

	app := &application{
		config: cfg,
		logi:   logi,
		loge:   loge,
	}

	// Sensible Configs for server
	srv := http.Server{
		Addr:         fmt.Sprint(":", cfg.port),
		Handler:      app.routes(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  1 * time.Minute,
	}

	slog.Info("starting server", "env", cfg.env, "port", cfg.port)

	log.Fatal(srv.ListenAndServe())

}
