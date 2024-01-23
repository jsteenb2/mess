package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/golang-migrate/migrate/v4"
	migsqlite "github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	"github.com/jsteenb2/mess/allsrv"
	"github.com/jsteenb2/mess/allsrv/migrations"
)

func main() {
	var db allsrv.DB = new(allsrv.InmemDB)
	if dsn := os.Getenv("ALLSRV_SQLITE_DSN"); dsn != "" {
		var err error
		db, err = newSQLiteDB(dsn)
		if err != nil {
			log.Println("failed to open sqlite db: " + err.Error())
			os.Exit(1)
		}
	}

	var svr http.Handler
	switch os.Getenv("ALLSRV_SERVER") {
	case "v1":
		log.Println("starting v1 server")
		svr = allsrv.NewServer(db, allsrv.WithBasicAuth("admin", "pass"))
	case "v2":
		log.Println("starting v2 server")
		svr = allsrv.NewServerV2(db, allsrv.WithBasicAuthV2("admin", "pass"))
	default: // run both
		log.Println("starting combination v1/v2 server")
		mux := http.NewServeMux()
		allsrv.NewServer(db, allsrv.WithMux(mux), allsrv.WithBasicAuth("admin", "pass"))
		allsrv.NewServerV2(db, allsrv.WithMux(mux), allsrv.WithBasicAuthV2("admin", "pass"))
		svr = mux
	}

	port := ":8091"
	log.Println("listening at http://localhost" + port)
	if err := http.ListenAndServe(port, svr); err != nil && err != http.ErrServerClosed {
		log.Println(err.Error())
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
