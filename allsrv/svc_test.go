package allsrv_test

import (
	"testing"

	"github.com/hashicorp/go-metrics"

	"github.com/jsteenb2/mess/allsrv"
)

func TestService(t *testing.T) {
	testSVC(t, func(t *testing.T, fields svcTestOpts) svcDeps {
		db := new(allsrv.InmemDB)
		fields.prepDB(t, db)

		var svc allsrv.SVC = allsrv.NewService(db, fields.svcOpts...)
		svc = allsrv.SVCLogging(newTestLogger(t))(svc)
		svc = allsrv.ObserveSVC(metrics.Default())(svc)

		return svcDeps{svc: svc}
	})
}
