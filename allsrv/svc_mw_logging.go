package allsrv

import (
	"context"
	"log/slog"
	"time"
	
	"github.com/jsteenb2/errors"
)

// SVCLogging wraps the service with logging concerns.
func SVCLogging(logger *slog.Logger) func(SVC) SVC {
	return func(next SVC) SVC {
		return &svcMWLogger{
			logger: logger,
			next:   next,
		}
	}
}

type svcMWLogger struct {
	logger *slog.Logger
	next   SVC
}

func (s *svcMWLogger) CreateFoo(ctx context.Context, f Foo) (Foo, error) {
	logFn := s.logFn(ctx, "input_name", f.Name, "input_note", f.Note)
	
	f, err := s.next.CreateFoo(ctx, f)
	logger := logFn(err)
	if err != nil {
		logger.Error("failed to create foo")
	} else {
		logger.Info("foo created successfully", "new_foo_id", f.ID)
	}
	
	return f, err
}

func (s *svcMWLogger) ReadFoo(ctx context.Context, id string) (Foo, error) {
	logFn := s.logFn(ctx, "input_id", id)
	
	f, err := s.next.ReadFoo(ctx, id)
	logger := logFn(err)
	if err != nil {
		logger.Error("failed to read foo")
	}
	
	return f, err
}

func (s *svcMWLogger) UpdateFoo(ctx context.Context, f FooUpd) (Foo, error) {
	fields := []any{"input_id", f.ID}
	if f.Name != nil {
		fields = append(fields, "input_name", *f.Name)
	}
	if f.Note != nil {
		fields = append(fields, "input_note", *f.Note)
	}
	
	logFn := s.logFn(ctx, fields...)
	
	updatedFoo, err := s.next.UpdateFoo(ctx, f)
	logger := logFn(err)
	if err != nil {
		logger.Error("failed to update foo")
	} else {
		logger.Info("foo updated successfully")
	}
	
	return updatedFoo, err
}

func (s *svcMWLogger) DelFoo(ctx context.Context, id string) error {
	logFn := s.logFn(ctx, "input_id", id)
	
	err := s.next.DelFoo(ctx, id)
	logger := logFn(err)
	if err != nil {
		logger.Error("failed to delete foo")
	} else {
		logger.Info("foo deleted successfully")
	}
	
	return err
}

func (s *svcMWLogger) logFn(ctx context.Context, fields ...any) func(error) *slog.Logger {
	start := time.Now()
	return func(err error) *slog.Logger {
		logger := s.logger.
			With(fields...).
			With(
				"took_ms", time.Since(start).Round(time.Millisecond).String(),
				"origin", getOrigin(ctx),
				"user_agent", getUserAgent(ctx),
				"trace_id", getTraceID(ctx),
			)
		if err != nil {
			logger = logger.With("err", err.Error())
			logger = logger.WithGroup("err_fields").With(errors.Fields(err)...)
		}
		return logger
	}
}
