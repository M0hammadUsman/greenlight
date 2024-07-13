package main

import (
	"context"
	"errors"
	"flag"
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
	"sync"
	"syscall"
	"time"
)

const version = "1.0.0"

type config struct {
	port int
	env  string
	db   struct {
		dsn         string
		maxCons     int
		maxIdleTime string
	}
	limiter struct {
		enabled bool
		rps     float64
		burst   int
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
}

// Dependencies lives here for the application
type application struct {
	config config
	models data.Models
	mailer mailer.Mailer
	wg     sync.WaitGroup
}

func main() {

	var cfg config

	// Setting up passed flags
	flag.IntVar(&cfg.port, "port", 8080, "API server port")
	flag.StringVar(&cfg.env, "env", "dev", "Environment (dev|stag|prod")
	// DB flags
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("GREENLIGHT_DB_DSN"), "PostgreSQL DSN")
	flag.IntVar(&cfg.db.maxCons, "db-max-conns", 25, "PostgreSQL max connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max idle connection time")
	// IP based rate limiter flags
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", false, "Enable rate limiter")
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	// SMTP Server flags
	flag.StringVar(&cfg.smtp.host, "smtp-host", "sandbox.smtp.mailtrap.io", "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 25, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", "107919fce4d92e", "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", "681725dbce237d", "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "Greenlight <no-reply@greenlight.alexedwards.net>", "SMTP sender")
	// parsing flags
	flag.Parse()

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
