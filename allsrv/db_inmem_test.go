package allsrv_test

import (
	"testing"

	"github.com/jsteenb2/mess/allsrv"
)

func TestInmemDB(t *testing.T) {
	testDB(t, func(t *testing.T) allsrv.DB {
		return new(allsrv.InmemDB)
	})
}
