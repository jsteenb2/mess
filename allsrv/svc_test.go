package allsrv_test

import (
	"testing"

	"github.com/hashicorp/go-metrics"

	"github.com/jsteenb2/mess/allsrv"
)

func TestService(t *testing.T) {
	testSVC(t, func(t *testing.T, opts svcTestOpts) svcDeps {
		return svcDeps{svc: newInmemSVC(t, opts)}
	})
}

func newInmemSVC(t *testing.T, opts svcTestOpts) allsrv.SVC {
	db := new(allsrv.InmemDB)
	opts.prepDB(t, db)

	var svc allsrv.SVC = allsrv.NewService(db, opts.svcOpts...)
	svc = allsrv.SVCLogging(newTestLogger(t))(svc)
	svc = allsrv.ObserveSVC(metrics.Default())(svc)

	return svc
}
