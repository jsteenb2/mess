package allsrv_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	
	"github.com/jsteenb2/mess/allsrv"
)

func TestServer(t *testing.T) {
	t.Run("foo create", func(t *testing.T) {
		t.Run("when provided a valid foo should pass", func(t *testing.T) {
			db := new(allsrv.InmemDB)
			svr := allsrv.NewServer(db, "dodgers@stink.com", "PaSsWoRd")
			
			req := httptest.NewRequest("POST", "/foo", newJSONBody(t, allsrv.Foo{
				Name: "first-foo",
				Note: "some note",
			}))
			req.SetBasicAuth("dodgers@stink.com", "PaSsWoRd")
			rec := httptest.NewRecorder()
			
			svr.ServeHTTP(rec, req)
			
			assert.Equal(t, http.StatusCreated, rec.Code)
			expectJSONBody(t, rec.Body, func(t *testing.T, got allsrv.Foo) {
				assert.NotZero(t, got.ID)
				got.ID = "" // this hurts :-(
				
				want := allsrv.Foo{
					ID:   "", // ruh ohhh
					Name: "first-foo",
					Note: "some note",
				}
				assert.Equal(t, want, got)
			})
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