package main

import (
	"cmp"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	migsqlite "github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/hashicorp/go-metrics"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	"github.com/jsteenb2/mess/allsrv"
	"github.com/jsteenb2/mess/allsrv/migrations"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{AddSource: true}))

	var db allsrv.DB = new(allsrv.InmemDB)
	if dsn := os.Getenv("ALLSRV_SQLITE_DSN"); dsn != "" {
		var err error
		db, err = newSQLiteDB(dsn)
		if err != nil {
			logger.Error("failed to open sqlite db", "err", err.Error())
			os.Exit(1)
		}
		logger.Info("sqlite database opened", "dsn", dsn)
	}

	mux := http.NewServeMux()

	selectedSVR := strings.TrimSpace(strings.ToLower(os.Getenv("ALLSRV_SERVER")))
	if selectedSVR != "v2" {
		logger.Info("registering v1 server")
		allsrv.NewServer(db, allsrv.WithBasicAuth("admin", "pass"), allsrv.WithMux(mux))
	}
	if selectedSVR != "v1" {
		logger.Info("registering v2 server")

		var svc allsrv.SVC = allsrv.NewService(db)
		svc = allsrv.SVCLogging(logger)(svc)

		met, err := metrics.New(metrics.DefaultConfig("allsrv"), metrics.NewInmemSink(5*time.Second, time.Minute))
		if err != nil {
			logger.Error("failed to create metrics", "err", err.Error())
			os.Exit(1)
		}
		svc = allsrv.ObserveSVC(met)(svc)

		allsrv.NewServerV2(svc, allsrv.WithBasicAuthV2("admin", "pass"), allsrv.WithMux(mux))
	}

	addr := "localhost:" + strings.TrimPrefix(cmp.Or(os.Getenv("ALLSRV_PORT"), "8091"), ":")
	logger.Info("listening at " + addr)
	if err := http.ListenAndServe(addr, mux); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("shut down error encountered", "err", err.Error(), "addr", addr)
		os.Exit(1)
	}
}

func newSQLiteDB(dsn string) (allsrv.DB, error) {
	const driver = "sqlite3"

	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}

	const dbName = "testdb"
	drvr, err := migsqlite.WithInstance(db, &migsqlite.Config{DatabaseName: dbName})
	if err != nil {
		return nil, err
	}

	iodrvr, err := iofs.New(migrations.SQLite, "sqlite")
	if err != nil {
		return nil, err
	}

	m, err := migrate.NewWithInstance("iofs", iodrvr, dbName, drvr)
	if err != nil {
		return nil, err
	}
	err = m.Up()
	if err != nil {
		return nil, err
	}

	dbx := sqlx.NewDb(db, driver)
	dbx.SetMaxIdleConns(1)

	return allsrv.NewSQLiteDB(dbx), nil
}
