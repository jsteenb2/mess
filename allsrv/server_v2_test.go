package allsrv_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jsteenb2/mess/allsrv"
)

func TestServerV2(t *testing.T) {
	t.Run("foo create", func(t *testing.T) {
		t.Run("when provided a valid foo should pass", func(t *testing.T) {
			db := new(allsrv.InmemDB)

			var svr http.Handler = allsrv.NewServerV2(
				db,
				allsrv.WithBasicAuthV2("dodgers@stink.com", "PaSsWoRd"),
				allsrv.WithMetrics(newTestMetrics(t)),
				allsrv.WithIDFn(func() string {
					return "id1"
				}),
				allsrv.WithNowFn(func() time.Time {
					return time.Time{}.UTC().Add(time.Hour)
				}),
			)

			req := httptest.NewRequest("POST", "/v1/foos", newJSONBody(t, allsrv.ReqCreateFooV1{
				Name: "first-foo",
				Note: "some note",
			}))
			req.Header.Set("Content-Type", "application/json")
			req.SetBasicAuth("dodgers@stink.com", "PaSsWoRd")
			rec := httptest.NewRecorder()

			svr.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusCreated, rec.Code)
			expectData[allsrv.FooAttrs](t, rec.Body, allsrv.RespData[allsrv.FooAttrs]{
				Type: "foo",
				ID:   "id1",
				Attributes: allsrv.FooAttrs{
					Name:      "first-foo",
					Note:      "some note",
					CreatedAt: time.Time{}.UTC().Add(time.Hour).Format(time.RFC3339),
				},
			})

			got, err := db.ReadFoo(context.TODO(), "id1")
			require.NoError(t, err)

			want := allsrv.Foo{
				ID:        "id1",
				Name:      "first-foo",
				Note:      "some note",
				CreatedAt: time.Time{}.UTC().Add(time.Hour),
			}
			assert.Equal(t, want, got)
		})

		t.Run("when missing required auth should fail", func(t *testing.T) {
			var svr http.Handler = allsrv.NewServerV2(
				new(allsrv.InmemDB),
				allsrv.WithBasicAuthV2("dodgers@stink.com", "PaSsWoRd"),
				allsrv.WithMetrics(newTestMetrics(t)),
				allsrv.WithIDFn(func() string {
					return "id1"
				}),
				allsrv.WithNowFn(func() time.Time {
					return time.Time{}.UTC().Add(time.Hour)
				}),
			)

			req := httptest.NewRequest("POST", "/v1/foos", newJSONBody(t, allsrv.ReqCreateFooV1{
				Name: "first-foo",
				Note: "some note",
			}))
			req.Header.Set("Content-Type", "application/json")
			req.SetBasicAuth("dodgers@stink.com", "WRONGO")
			rec := httptest.NewRecorder()

			svr.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusUnauthorized, rec.Code)
			expectErrs(t, rec.Body, func(t *testing.T, got []allsrv.RespErr) {
				require.Len(t, got, 1)

				want := allsrv.RespErr{
					Status: 401,
					Code:   4,
					Msg:    "unauthorized access",
					Source: &allsrv.RespErrSource{
						Header: "Authorization",
					},
				}
				assert.Equal(t, want, got[0])
			})
		})
	})
}

func expectErrs(t *testing.T, r io.Reader, fn func(t *testing.T, got []allsrv.RespErr)) {
	t.Helper()

	expectJSONBody(t, r, func(t *testing.T, got allsrv.RespResourceBody[any]) {
		require.Nil(t, got.Data)
		require.NotEmpty(t, got.Errs)

		fn(t, got.Errs)
	})
}

func expectData[Attrs any | []any](t *testing.T, r io.Reader, want allsrv.RespData[Attrs]) {
	t.Helper()

	expectJSONBody(t, r, func(t *testing.T, got allsrv.RespResourceBody[Attrs]) {
		require.Empty(t, got.Errs)
		require.NotNil(t, got.Data)

		assert.Equal(t, want, *got.Data)
	})
}
