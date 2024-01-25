package allsrv

import (
	"context"
	"time"

	"github.com/hashicorp/go-metrics"
	"github.com/opentracing/opentracing-go"
)

const (
	metricsPrefix = "mess"
)

// ObserveDB provides observability to the database.
func ObserveDB(name string, met *metrics.Metrics) func(DB) DB {
	return func(next DB) DB {
		return &dbMW{
			name: name,
			next: next,
			met:  met,
		}
	}
}

type dbMW struct {
	name string
	next DB
	met  *metrics.Metrics
}

func (d *dbMW) CreateFoo(ctx context.Context, f Foo) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "db_"+d.name+"_foo_create")
	defer span.Finish()

	rec := d.record("create")
	return rec(d.next.CreateFoo(ctx, f))
}

func (d *dbMW) ReadFoo(ctx context.Context, id string) (Foo, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "db_"+d.name+"_foo_read")
	defer span.Finish()

	rec := d.record("read")
	f, err := d.next.ReadFoo(ctx, id)
	return f, rec(err)
}

func (d *dbMW) UpdateFoo(ctx context.Context, f Foo) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "db_"+d.name+"_foo_update")
	defer span.Finish()

	rec := d.record("update")
	return rec(d.next.UpdateFoo(ctx, f))
}

func (d *dbMW) DelFoo(ctx context.Context, id string) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "db_"+d.name+"_foo_delete")
	defer span.Finish()

	rec := d.record("delete")
	return rec(d.next.DelFoo(ctx, id))
}

func (d *dbMW) record(op string) func(error) error {
	start := time.Now()
	name := []string{metricsPrefix, d.name, op}
	d.met.IncrCounter(append(name, "reqs"), 1)
	return func(err error) error {
		if err != nil {
			d.met.IncrCounter(append(name, "errs"), 1)
		}
		d.met.MeasureSince(append(name, "dur"), start)
		return err
	}
}
