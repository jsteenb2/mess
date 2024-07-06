package allsrv

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jsteenb2/errors"
)

// Foo represents the foo domain entity.
type Foo struct {
	ID        string
	Name      string
	Note      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// OK validates the fields are provided.
func (f Foo) OK() error {
	if f.Name == "" {
		return InvalidErr("name is required")
	}
	return nil
}

// FooUpd is a record for updating an existing foo.
type FooUpd struct {
	ID   string
	Name *string
	Note *string
}

// SVC defines the service behavior.
type SVC interface {
	CreateFoo(ctx context.Context, f Foo) (Foo, error)
	ReadFoo(ctx context.Context, id string) (Foo, error)
	UpdateFoo(ctx context.Context, f FooUpd) (Foo, error)
	DelFoo(ctx context.Context, id string) error
}

// Service dependencies
type (
	// DB represents the foo persistence layer.
	DB interface {
		CreateFoo(ctx context.Context, f Foo) error
		ReadFoo(ctx context.Context, id string) (Foo, error)
		UpdateFoo(ctx context.Context, f Foo) error
		DelFoo(ctx context.Context, id string) error
	}
)

// Service is the home for business logic of the foo domain.
type Service struct {
	db DB

	idFn  func() string
	nowFn func() time.Time
}

func WithSVCIDFn(fn func() string) func(*Service) {
	return func(s *Service) {
		s.idFn = fn
	}
}

func WithSVCNowFn(fn func() time.Time) func(*Service) {
	return func(s *Service) {
		s.nowFn = fn
	}
}

func NewService(db DB, opts ...func(*Service)) *Service {
	s := Service{
		db:    db,
		idFn:  func() string { return uuid.Must(uuid.NewV4()).String() },
		nowFn: func() time.Time { return time.Now().UTC() },
	}

	for _, o := range opts {
		o(&s)
	}

	return &s
}

func (s *Service) CreateFoo(ctx context.Context, f Foo) (Foo, error) {
	if err := f.OK(); err != nil {
		return Foo{}, errors.Wrap(err)
	}

	now := s.nowFn()
	f.ID, f.CreatedAt, f.UpdatedAt = s.idFn(), now, now

	if err := s.db.CreateFoo(ctx, f); err != nil {
		return Foo{}, errors.Wrap(err)
	}

	return f, nil
}

func (s *Service) ReadFoo(ctx context.Context, id string) (Foo, error) {
	if id == "" {
		return Foo{}, errIDRequired
	}
	f, err := s.db.ReadFoo(ctx, id)
	return f, errors.Wrap(err)
}

func (s *Service) UpdateFoo(ctx context.Context, f FooUpd) (Foo, error) {
	existing, err := s.db.ReadFoo(ctx, f.ID)
	if err != nil {
		return Foo{}, errors.Wrap(err)
	}
	if newName := f.Name; newName != nil {
		existing.Name = *newName
	}
	if newNote := f.Note; newNote != nil {
		existing.Note = *newNote
	}
	existing.UpdatedAt = s.nowFn()

	err = s.db.UpdateFoo(ctx, existing)
	if err != nil {
		return Foo{}, errors.Wrap(err)
	}

	return existing, nil
}

func (s *Service) DelFoo(ctx context.Context, id string) error {
	if id == "" {
		return errors.Wrap(errIDRequired)
	}
	return errors.Wrap(s.db.DelFoo(ctx, id))
}
