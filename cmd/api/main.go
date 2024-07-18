package main

import (
	"context"
	"errors"
	"expvar"
	"fmt"
	"github.com/M0hammadUsman/greenlight/internal/data"
	"github.com/M0hammadUsman/greenlight/internal/mailer"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lmittmann/tint"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path"
	"runtime"
	"sync"
	"syscall"
	"time"
)

var version string

// Create a buildTime variable to hold the executable binary build time. Note that this
// must be a string type, as the -X linker flag will only work with string variables.
var buildTime string

// Dependencies lives here for the application
type application struct {
	config config
	models data.Models
	mailer mailer.Mailer
	wg     sync.WaitGroup
}

func main() {

	cfg := parseConfigFlags()
	// Loggers configuration
	configureLoggers()

	// DB configuration
	db, err := openDB(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	slog.Info("database connection pool established")

	app := &application{
		config: cfg,
		models: data.NewModels(db),
		mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
	}
	//Exposing custom metrics
	exposeCustomMetrics(db)
	// Starting server
	if err = app.serve(); err != nil {
		slog.Error(err.Error(), "shutdown", "hard...")
	}
}

func configureLoggers() {
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
}

func openDB(cfg config) (*pgxpool.Pool, error) {
	// Use sql.Open() to create an empty connection pool, using the DSN from the config struct
	maxIdleDuration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}
	dbConfig, err := pgxpool.ParseConfig(cfg.db.dsn)
	if err != nil {
		return nil, err
	}
	dbConfig.MaxConns = int32(cfg.db.maxCons)
	dbConfig.MaxConnIdleTime = maxIdleDuration
	dbConfig.MaxConnLifetime = 48 * time.Hour // Just want this to be LARGE, defaults to 1h
	db, err := pgxpool.NewWithConfig(context.Background(), dbConfig)
	if err != nil {
		return nil, err
	}
	// Create a context with a 5-second timeout deadline.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Use Ping to establish a new connection to the database, passing in the context we created above
	// as a parameter. If the connection couldn't be established successfully within the 5-second deadline,
	//then this will return an error.
	if err = db.Ping(ctx); err != nil {
		return nil, err
	}
	return db, nil
}

func (app *application) serve() error {
	srv := http.Server{ // Sensible Configs for server
		Addr:         fmt.Sprint(":", app.config.port),
		Handler:      app.routes(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  1 * time.Minute,
	}
	shutdownError := make(chan error)
	// Background goroutine to listen to termination signals -> SIGINT, SIGTERM
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
		s := <-quit
		slog.Info("shutting down server", "signal", s.String())
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			shutdownError <- err
		}
		slog.Info("completing background tasks", "addr", srv.Addr)
		app.wg.Wait()
		shutdownError <- nil
	}() // Start the server normally
	slog.Info("starting server", "env", app.config.env, "port", app.config.port)
	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	} else {
		slog.Info("waiting for ongoing http requests", "max wait", "5 sec...")
	}
	if err = <-shutdownError; err != nil {
		return err
	}
	slog.Info("stopped server", "addr", srv.Addr)
	return nil
}

func exposeCustomMetrics(DB *pgxpool.Pool) {
	expvar.NewString("version").Set(version)
	expvar.Publish("goroutines", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))
	expvar.Publish("timestamp", expvar.Func(func() any {
		return time.Now().Unix()
	}))
	expvar.Publish("database", expvar.Func(func() any {
		DBStats := DB.Stat()
		stats := struct {
			AcquireCount            int64
			AcquireDuration         time.Duration
			AcquiredConns           int32
			CanceledAcquireCount    int64
			ConstructingConns       int32
			EmptyAcquireCount       int64
			IdleConns               int32
			MaxConns                int32
			TotalConns              int32
			NewConnsCount           int64
			MaxLifetimeDestroyCount int64
			MaxIdleDestroyCount     int64
		}{
			AcquireCount:            DBStats.AcquireCount(),
			AcquireDuration:         DBStats.AcquireDuration(),
			AcquiredConns:           DBStats.AcquiredConns(),
			CanceledAcquireCount:    DBStats.CanceledAcquireCount(),
			ConstructingConns:       DBStats.ConstructingConns(),
			EmptyAcquireCount:       DBStats.EmptyAcquireCount(),
			IdleConns:               DBStats.IdleConns(),
			MaxConns:                DBStats.MaxConns(),
			TotalConns:              DBStats.TotalConns(),
			NewConnsCount:           DBStats.NewConnsCount(),
			MaxLifetimeDestroyCount: DBStats.MaxLifetimeDestroyCount(),
			MaxIdleDestroyCount:     DBStats.MaxIdleDestroyCount(),
		}
		return stats
	}))
}
