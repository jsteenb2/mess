package allsrv

import (
	"context"
	"time"

	"github.com/hashicorp/go-metrics"
	"github.com/opentracing/opentracing-go"
)

// ObserveSVC provides a metrics and spanning middleware.
func ObserveSVC(met *metrics.Metrics) func(next SVC) SVC {
	return func(next SVC) SVC {
		return &svcObserver{
			met:  met,
			next: next,
		}
	}
}

type svcObserver struct {
	met  *metrics.Metrics
	next SVC
}

func (s *svcObserver) CreateFoo(ctx context.Context, f Foo) (Foo, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "svc_foo_create")
	defer span.Finish()

	rec := s.record("create")
	f, err := s.next.CreateFoo(ctx, f)
	return f, rec(err)
}

func (s *svcObserver) ReadFoo(ctx context.Context, id string) (Foo, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "svc_foo_read")
	defer span.Finish()

	rec := s.record("read")
	f, err := s.next.ReadFoo(ctx, id)
	return f, rec(err)
}

func (s *svcObserver) UpdateFoo(ctx context.Context, f FooUpd) (Foo, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "svc_foo_update")
	defer span.Finish()

	rec := s.record("update")
	updatedFoo, err := s.next.UpdateFoo(ctx, f)
	return updatedFoo, rec(err)
}

func (s *svcObserver) DelFoo(ctx context.Context, id string) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "svc_foo_delete")
	defer span.Finish()

	rec := s.record("delete")
	return rec(s.next.DelFoo(ctx, id))
}

func (s *svcObserver) record(op string) func(error) error {
	start := time.Now()
	name := []string{metricsPrefix, op}
	s.met.IncrCounter(append(name, "reqs"), 1)
	return func(err error) error {
		if err != nil {
			s.met.IncrCounter(append(name, "errs"), 1)
		}
		s.met.MeasureSince(append(name, "dur"), start)
		return err
	}
}
