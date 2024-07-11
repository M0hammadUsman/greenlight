package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/M0hammadUsman/greenlight/internal/data"
	"github.com/jackc/pgx/v5/pgxpool"
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
}

// Dependencies lives here for the application
type application struct {
	config config
	models data.Models
	logi   *log.Logger
	loge   *log.Logger
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
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	// parsing flags
	flag.Parse()

	// Loggers configuration
	logi, loge := configureLoggers()

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

func configureLoggers() (*log.Logger, *log.Logger) {
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

	logi := log.New(os.Stderr, "INFO\t", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
	loge := log.New(os.Stderr, "ERROR\t", log.LstdFlags|log.Lmicroseconds|log.Llongfile)

	return logi, loge
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
