package allsrv_test

import (
	"testing"

	"github.com/jsteenb2/mess/allsrv/allsrvtesting"
)

func TestService(t *testing.T) {
	allsrvtesting.TestSVC(t, func(t *testing.T, opts allsrvtesting.SVCTestOpts) allsrvtesting.SVCDeps {
		return allsrvtesting.SVCDeps{SVC: allsrvtesting.NewInmemSVC(t, opts)}
	})
}

func TestServiceSqlite(t *testing.T) {
	allsrvtesting.TestSVC(t, func(t *testing.T, opts allsrvtesting.SVCTestOpts) allsrvtesting.SVCDeps {
		db := newSQLiteDB(t)
		return allsrvtesting.SVCDeps{SVC: allsrvtesting.NewSVC(t, db, opts)}
	})
}
