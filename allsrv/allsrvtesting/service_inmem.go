package allsrvtesting

import (
	"log/slog"
	"testing"

	"github.com/hashicorp/go-metrics"

	"github.com/jsteenb2/mess/allsrv"
)

func NewInmemSVC(t *testing.T, opts SVCTestOpts) allsrv.SVC {
	return NewSVC(t, new(allsrv.InmemDB), opts)
}

func NewSVC(t *testing.T, db allsrv.DB, opts SVCTestOpts) allsrv.SVC {
	opts.PrepDB(t, db)
	var svc allsrv.SVC = allsrv.NewService(db, opts.SVCOpts...)
	svc = allsrv.SVCLogging(newTestLogger(t))(svc)
	svc = allsrv.ObserveSVC(metrics.Default())(svc)

	return svc
}

func newTestLogger(t *testing.T) *slog.Logger {
	return slog.New(slog.NewJSONHandler(&testr{t: t}, nil))
}

type testr struct {
	t *testing.T
}

func (t *testr) Write(p []byte) (n int, err error) {
	t.t.Log(string(p))
	return len(p), nil
}
