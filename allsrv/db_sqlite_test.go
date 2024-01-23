package allsrv_test

import (
	"database/sql"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	migsqlite "github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jsteenb2/mess/allsrv"
	"github.com/jsteenb2/mess/allsrv/migrations"
)

func TestSQLite(t *testing.T) {
	testDB(t, func(t *testing.T) allsrv.DB {
		db := newSQLiteInmem(t)
		t.Cleanup(func() {
			assert.NoError(t, db.Close())
		})

		return allsrv.NewSQLiteDB(db)
	})
}

func newSQLiteInmem(t *testing.T) *sqlx.DB {
	const driver = "sqlite3"

	db, err := sql.Open(driver, ":memory:")
	require.NoError(t, err)

	const dbName = "testdb"
	drvr, err := migsqlite.WithInstance(db, &migsqlite.Config{DatabaseName: dbName})
	require.NoError(t, err)

	iodrvr, err := iofs.New(migrations.SQLite, "sqlite")
	require.NoError(t, err)

	m, err := migrate.NewWithInstance("iofs", iodrvr, dbName, drvr)
	require.NoError(t, err)
	require.NoError(t, m.Up())

	dbx := sqlx.NewDb(db, driver)
	dbx.SetMaxIdleConns(1)
	return dbx
}
