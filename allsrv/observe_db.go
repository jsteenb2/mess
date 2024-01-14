package allsrv

import (
	"time"

	"github.com/hashicorp/go-metrics"
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

func (d *dbMW) CreateFoo(f Foo) error {
	rec := d.record("create")
	return rec(d.next.CreateFoo(f))
}

func (d *dbMW) ReadFoo(id string) (Foo, error) {
	rec := d.record("read")
	f, err := d.next.ReadFoo(id)
	return f, rec(err)
}

func (d *dbMW) UpdateFoo(f Foo) error {
	rec := d.record("update")
	return rec(d.next.UpdateFoo(f))
}

func (d *dbMW) DelFoo(id string) error {
	rec := d.record("delete")
	return rec(d.next.DelFoo(id))
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
