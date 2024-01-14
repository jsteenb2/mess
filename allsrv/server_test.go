package allsrv_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/go-metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jsteenb2/mess/allsrv"
)

func TestServer(t *testing.T) {
	t.Run("foo create", func(t *testing.T) {
		t.Run("when provided a valid foo should pass", func(t *testing.T) {
			db := allsrv.ObserveDB("inmem", newTestMetrics(t))(new(allsrv.InmemDB))
			svr := allsrv.NewServer(db,
				allsrv.WithBasicAuth("dodgers@stink.com", "PaSsWoRd"),
				allsrv.WithIDFn(func() string {
					return "id1"
				}),
			)

			req := httptest.NewRequest("POST", "/foo", newJSONBody(t, allsrv.Foo{
				Name: "first-foo",
				Note: "some note",
			}))
			req.SetBasicAuth("dodgers@stink.com", "PaSsWoRd")
			rec := httptest.NewRecorder()

			svr.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusCreated, rec.Code)
			expectJSONBody(t, rec.Body, func(t *testing.T, got allsrv.Foo) {
				want := allsrv.Foo{
					ID:   "id1",
					Name: "first-foo",
					Note: "some note",
				}
				assert.Equal(t, want, got)
			})
		})

		t.Run("when provided invalid basic auth should fail", func(t *testing.T) {
			svr := allsrv.NewServer(new(allsrv.InmemDB), allsrv.WithBasicAuth("dodgers@stink.com", "PaSsWoRd"))

			req := httptest.NewRequest("POST", "/foo", newJSONBody(t, allsrv.Foo{
				Name: "first-foo",
				Note: "some note",
			}))
			req.SetBasicAuth("dodgers@rule.com", "wrongO")
			rec := httptest.NewRecorder()

			svr.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		})
	})

	t.Run("foo read", func(t *testing.T) {
		t.Run("when querying for existing foo id should pass", func(t *testing.T) {
			db := allsrv.ObserveDB("inmem", newTestMetrics(t))(new(allsrv.InmemDB))
			err := db.CreateFoo(context.TODO(), allsrv.Foo{
				ID:   "reader1",
				Name: "read",
				Note: "another note",
			})
			require.NoError(t, err)

			svr := allsrv.NewServer(db, allsrv.WithBasicAuth("dodgers@stink.com", "PaSsWoRd"))

			req := httptest.NewRequest("GET", "/foo?id=reader1", nil)
			req.SetBasicAuth("dodgers@stink.com", "PaSsWoRd")
			rec := httptest.NewRecorder()

			svr.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
			expectJSONBody(t, rec.Body, func(t *testing.T, got allsrv.Foo) {
				want := allsrv.Foo{
					ID:   "reader1",
					Name: "read",
					Note: "another note",
				}
				assert.Equal(t, want, got)
			})
		})

		t.Run("when provided invalid basic auth should fail", func(t *testing.T) {
			svr := allsrv.NewServer(new(allsrv.InmemDB), allsrv.WithBasicAuth("dodgers@stink.com", "PaSsWoRd"))

			req := httptest.NewRequest("GET", "/foo?id=reader1", nil)
			req.SetBasicAuth("dodgers@rule.com", "wrongO")
			rec := httptest.NewRecorder()

			svr.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		})
	})

	t.Run("foo update", func(t *testing.T) {
		t.Run("when updating an existing foo with valid changes should pass", func(t *testing.T) {
			db := allsrv.ObserveDB("inmem", newTestMetrics(t))(new(allsrv.InmemDB))
			err := db.CreateFoo(context.TODO(), allsrv.Foo{
				ID:   "id1",
				Name: "first_name",
				Note: "first note",
			})
			require.NoError(t, err)

			svr := allsrv.NewServer(db, allsrv.WithBasicAuth("dodgers@stink.com", "PaSsWoRd"))

			req := httptest.NewRequest("PUT", "/foo", newJSONBody(t, allsrv.Foo{
				ID:   "id1",
				Name: "second_name",
				Note: "second note",
			}))
			req.SetBasicAuth("dodgers@stink.com", "PaSsWoRd")
			rec := httptest.NewRecorder()

			svr.ServeHTTP(rec, req)

			// note: lame we don't get the updated foo back
			assert.Equal(t, http.StatusOK, rec.Code)
		})

		t.Run("when provided invalid basic auth should fail", func(t *testing.T) {
			db := new(allsrv.InmemDB)
			err := db.CreateFoo(context.TODO(), allsrv.Foo{
				ID:   "id1",
				Name: "first_name",
				Note: "first note",
			})
			require.NoError(t, err)

			svr := allsrv.NewServer(db, allsrv.WithBasicAuth("dodgers@stink.com", "PaSsWoRd"))

			req := httptest.NewRequest("PUT", "/foo", newJSONBody(t, allsrv.Foo{
				ID:   "id1",
				Name: "second_name",
				Note: "second note",
			}))
			req.SetBasicAuth("dodgers@rule.com", "wrongO")
			rec := httptest.NewRecorder()

			svr.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		})
	})

	t.Run("foo delete", func(t *testing.T) {
		t.Run("when deleting an existing foo should pass", func(t *testing.T) {
			db := allsrv.ObserveDB("inmem", newTestMetrics(t))(new(allsrv.InmemDB))
			err := db.CreateFoo(context.TODO(), allsrv.Foo{
				ID:   "id1",
				Name: "first_name",
				Note: "first note",
			})
			require.NoError(t, err)

			svr := allsrv.NewServer(db, allsrv.WithBasicAuth("dodgers@stink.com", "PaSsWoRd"))

			req := httptest.NewRequest("DELETE", "/foo?id=id1", nil)
			req.SetBasicAuth("dodgers@stink.com", "PaSsWoRd")
			rec := httptest.NewRecorder()

			svr.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
		})

		t.Run("when provided invalid basic auth should fail", func(t *testing.T) {
			svr := allsrv.NewServer(new(allsrv.InmemDB), allsrv.WithBasicAuth("dodgers@stink.com", "PaSsWoRd"))

			req := httptest.NewRequest("DELETE", "/foo?id=id1", nil)
			req.SetBasicAuth("dodgers@rule.com", "wrongO")
			rec := httptest.NewRecorder()

			svr.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		})
	})
}

func newJSONBody(t *testing.T, v any) *bytes.Buffer {
	t.Helper()

	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(v)
	require.NoError(t, err)

	return &buf
}

func expectJSONBody[T any](t *testing.T, r io.Reader, assertFn func(t *testing.T, got T)) {
	t.Helper()

	var out T
	err := json.NewDecoder(r).Decode(&out)
	require.NoError(t, err)

	assertFn(t, out)
}

func newTestMetrics(t *testing.T) *metrics.Metrics {
	t.Helper()

	met, err := metrics.New(&metrics.Config{}, &metrics.BlackholeSink{})
	require.NoError(t, err)

	return met
}
