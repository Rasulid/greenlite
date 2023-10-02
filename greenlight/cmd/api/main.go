package main

import (
	"context"
	"database/sql"
	"flag"
	"os"
	"sync"
	"time"

	_ "github.com/lib/pq"
	"greenlight.rasulabduvaitov.net/internal/data"
	"greenlight.rasulabduvaitov.net/internal/jsonlog"
	"greenlight.rasulabduvaitov.net/internal/mailer"
)

const version = "1.0"

type config struct {
	port        int
	environment string
	db          struct {
		dns             string
		maxOpenConn     int
		maxIdleConn     int
		maxIdleConnTime string
	}
	limiter struct {
		rps float64
		burst int
		enabled bool
	}
	smtp struct {
		host string
		port int
		username string
		password string
		sender string
	}
}

type application struct {
	Config *config
	logger *jsonlog.Logger
	models data.Models
	mailer mailer.Mailer
	wg sync.WaitGroup
}

func main() {
	var cnf config

	flag.IntVar(&cnf.port, "port", 4000, "Port to connect to")
	flag.StringVar(&cnf.environment, "environment", "development", "Environment (Environment|Starting|Production)")


	flag.StringVar(&cnf.db.dns, "db-dns", "postgres://greenlight:pa55word@localhost/greenlight", "PostgreSQL DNS")
	flag.IntVar(&cnf.db.maxOpenConn, "db-max-open-conns", 25, "Postgres max open connection")
	flag.IntVar(&cnf.db.maxIdleConn, "db-max-idle-conns", 25, "Postgres max idle connection")
	flag.StringVar(&cnf.db.maxIdleConnTime, "max-idle-time", "15m", "Posters max connection idle time")

	flag.Float64Var(&cnf.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cnf.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cnf.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	flag.StringVar(&cnf.smtp.host, "smtp-host", "smtp.mailtrap.io", "SMTP host")
	flag.IntVar(&cnf.smtp.port, "smtp-port", 25, "SMTP port")
	flag.StringVar(&cnf.smtp.username, "smtp-username", "12e20bbb3816a9", "SMTP username")
	flag.StringVar(&cnf.smtp.password, "smtp-password", "d5685d90015793", "SMTP password")
	flag.StringVar(&cnf.smtp.sender, "smtp-sender", "Greenlight <no-reply@greenlight.alexedwards.net>", "SMTP sender")
	flag.Parse()

	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	db, err := openDB(cnf)
	if err != nil {
		logger.PrintFatal(err, nil)
	}
	defer db.Close()

	logger.PrintInfo("database connection pool established", nil)

	app := &application{
		Config: &cnf,
		logger: logger,
		models: data.NewMovies(db),
		mailer: mailer.New(cnf.smtp.host, cnf.smtp.port, cnf.smtp.username, cnf.smtp.password, cnf.smtp.sender),
	}

	err = app.server()
	if err != nil {
		logger.PrintFatal(err, nil)
	}

}




func openDB(cnf config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cnf.db.dns)
	if err != nil {
		return nil, err
	}

	ctx, cenel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cenel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(cnf.db.maxOpenConn)
	db.SetMaxIdleConns(cnf.db.maxIdleConn)

	duration, err := time.ParseDuration(cnf.db.maxIdleConnTime)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxIdleTime(duration)

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}
