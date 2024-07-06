package allsrv

import (
	"context"
	"database/sql"
	"sync"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/jsteenb2/errors"
	"github.com/mattn/go-sqlite3"
)

// NewSQLiteDB creates a new sqlite db.
func NewSQLiteDB(db *sqlx.DB) DB {
	return &sqlDB{
		db: db,
		sq: sq.StatementBuilder.PlaceholderFormat(sq.Question),
	}
}

type sqlDB struct {
	db *sqlx.DB
	sq sq.StatementBuilderType

	mu sync.RWMutex
}

func (s *sqlDB) CreateFoo(ctx context.Context, f Foo) error {
	sb := s.sq.
		Insert("foos").
		Columns("id", "name", "note", "created_at", "updated_at").
		Values(f.ID, f.Name, f.Note, f.CreatedAt, f.UpdatedAt)

	_, err := s.exec(ctx, sb)
	return errors.Wrap(err)
}

func (s *sqlDB) ReadFoo(ctx context.Context, id string) (Foo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	const query = `SELECT * FROM foos WHERE id=?`

	var ent entFoo
	err := s.db.GetContext(ctx, &ent, s.db.Rebind(query), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Foo{}, NotFoundErr("foo not found for id: " + id)
		}
		return Foo{}, errors.Wrap(err, errSQLiteFields(err))
	}

	out := Foo{
		ID:        ent.ID,
		Name:      ent.Name,
		Note:      ent.Note,
		CreatedAt: ent.CreatedAt,
		UpdatedAt: ent.UpdatedAt,
	}

	return out, nil
}

func (s *sqlDB) UpdateFoo(ctx context.Context, f Foo) error {
	sb := s.sq.
		Update("foos").
		Set("name", f.Name).
		Set("note", f.Note).
		Set("updated_at", f.UpdatedAt).
		Where(sq.Eq{"id": f.ID})
	
	return errors.Wrap(s.update(ctx, sb))
}

func (s *sqlDB) DelFoo(ctx context.Context, id string) error {
	err := s.update(ctx, s.sq.Delete("foos").Where(sq.Eq{"id": id}))
	return errors.Wrap(err)
}

func (s *sqlDB) exec(ctx context.Context, sqlizer sq.Sqlizer) (sql.Result, error) {
	query, args, err := sqlizer.ToSql()
	if err != nil {
		return nil, errors.Wrap(err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	res, err := s.db.ExecContext(ctx, query, args...)
	if sqErr := new(sqlite3.Error); errors.As(err, sqErr) {
		switch sqErr.Code {
		case sqlite3.ErrConstraint:
			return nil, ExistsErr("foo exists", errSQLiteFields(err))
		}
	}
	return res, errors.Wrap(err, errSQLiteFields(err))
}

func (s *sqlDB) update(ctx context.Context, sqlizer sq.Sqlizer) error {
	res, err := s.exec(ctx, sqlizer)
	if err != nil {
		return errors.Wrap(err, errSQLiteFields(err))
	}

	n, err := res.RowsAffected()
	if err == nil && n == 0 {
		return NotFoundErr("foo not found")
	}

	return nil
}

type entFoo struct {
	ID        string    `db:"id"`
	Name      string    `db:"name"`
	Note      string    `db:"note"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func errSQLiteFields(err error) []errors.KV {
	if sqErr := new(sqlite3.Error); errors.As(err, sqErr) {
		return errors.KVs(
			"sqlite_err_code", sqErr.Code,
			"sqlite_err_extended_code", sqErr.ExtendedCode,
			"sqlite_system_errno", sqErr.SystemErrno,
		)
	}
	return nil
}
