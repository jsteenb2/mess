package allsrv_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jsteenb2/mess/allsrv"
)

func TestServerV2(t *testing.T) {
	start := time.Time{}.Add(time.Hour).UTC()

	t.Run("foo create", func(t *testing.T) {
		type (
			inputs struct {
				req *http.Request
			}

			wantFn func(t *testing.T, rec *httptest.ResponseRecorder, db allsrv.DB)
		)

		tests := []struct {
			name    string
			prepare func(t *testing.T, db allsrv.DB)
			svrOpts []allsrv.SvrOptFn
			inputs  inputs
			want    wantFn
		}{
			{
				name:    "when provided a valid foo and authorized user should pass",
				svrOpts: []allsrv.SvrOptFn{allsrv.WithBasicAuthV2("dodgers@stink.com", "PaSsWoRd")},
				inputs: inputs{
					req: newJSONReq("POST", "/v1/foos",
						newJSONBody(t, allsrv.ReqCreateFooV1{
							Type: "foo",
							Attrs: allsrv.FooAttrs{
								Name: "first-foo",
								Note: "some note",
							},
						}),
						withBasicAuth("dodgers@stink.com", "PaSsWoRd"),
					),
				},
				want: func(t *testing.T, rec *httptest.ResponseRecorder, db allsrv.DB) {
					assert.Equal(t, http.StatusCreated, rec.Code)
					expectData[allsrv.FooAttrs](t, rec.Body, allsrv.Data[allsrv.FooAttrs]{
						Type: "foo",
						ID:   "1",
						Attrs: allsrv.FooAttrs{
							Name:      "first-foo",
							Note:      "some note",
							CreatedAt: start.Format(time.RFC3339),
						},
					})

					dbHasFoo(t, db, allsrv.Foo{
						ID:        "1",
						Name:      "first-foo",
						Note:      "some note",
						CreatedAt: start,
					})
				},
			},
			{
				name:    "when missing required auth should fail",
				svrOpts: []allsrv.SvrOptFn{allsrv.WithBasicAuthV2("dodgers@stink.com", "PaSsWoRd")},
				inputs: inputs{
					req: newJSONReq("POST", "/v1/foos",
						newJSONBody(t, allsrv.ReqCreateFooV1{
							Type: "foo",
							Attrs: allsrv.FooAttrs{
								Name: "first-foo",
								Note: "some note",
							},
						}),
						withBasicAuth("dodgers@stink.com", "WRONGO"),
					),
				},
				want: func(t *testing.T, rec *httptest.ResponseRecorder, db allsrv.DB) {
					assert.Equal(t, http.StatusUnauthorized, rec.Code)
					expectErrs(t, rec.Body, allsrv.RespErr{
						Status: http.StatusUnauthorized,
						Code:   4,
						Msg:    "unauthorized access",
						Source: &allsrv.RespErrSource{
							Header: "Authorization",
						},
					})

					_, err := db.ReadFoo(context.TODO(), "1")
					require.Error(t, err)
				},
			},
			{
				name:    "when creating foo with name that collides with existing should fail",
				prepare: createFoos(allsrv.Foo{ID: "9000", Name: "existing-foo"}),
				inputs: inputs{
					req: newJSONReq("POST", "/v1/foos", newJSONBody(t, allsrv.ReqCreateFooV1{
						Type: "foo",
						Attrs: allsrv.FooAttrs{
							Name: "existing-foo",
							Note: "some note",
						},
					})),
				},
				want: func(t *testing.T, rec *httptest.ResponseRecorder, db allsrv.DB) {
					assert.Equal(t, http.StatusConflict, rec.Code)
					expectErrs(t, rec.Body, allsrv.RespErr{
						Status: http.StatusConflict,
						Code:   1,
						Msg:    "foo existing-foo exists",
						Source: &allsrv.RespErrSource{
							Pointer: "/data/attributes/name",
						},
					})

					_, err := db.ReadFoo(context.TODO(), "1")
					require.Error(t, err)
				},
			},
			{
				name: "when creating foo with invalid resource type should fail",
				inputs: inputs{
					req: newJSONReq("POST", "/v1/foos", newJSONBody(t, allsrv.ReqCreateFooV1{
						Type: "WRONGO",
						Attrs: allsrv.FooAttrs{
							Name: "first-foo",
							Note: "some note",
						},
					})),
				},
				want: func(t *testing.T, rec *httptest.ResponseRecorder, db allsrv.DB) {
					assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
					expectErrs(t, rec.Body, allsrv.RespErr{
						Status: http.StatusUnprocessableEntity,
						Code:   2,
						Msg:    "type must be foo",
						Source: &allsrv.RespErrSource{
							Pointer: "/data/type",
						},
					})

					_, err := db.ReadFoo(context.TODO(), "1")
					require.Error(t, err)
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				db := new(allsrv.InmemDB)

				if tt.prepare != nil {
					tt.prepare(t, db)
				}

				defaultOpts := []allsrv.SvrOptFn{
					allsrv.WithIDFn(newIDGen(1, 1)),
					allsrv.WithNowFn(newNowFn(start, time.Hour)),
					allsrv.WithMetrics(newTestMetrics(t)),
				}
				opts := append(defaultOpts, tt.svrOpts...)

				rec := httptest.NewRecorder()

				svr := allsrv.NewServerV2(db, opts...)
				svr.ServeHTTP(rec, tt.inputs.req)

				tt.want(t, rec, db)
			})
		}
	})
}

func expectErrs(t *testing.T, r io.Reader, want ...allsrv.RespErr) {
	t.Helper()

	expectJSONBody(t, r, func(t *testing.T, got allsrv.RespResourceBody[any]) {
		t.Helper()

		require.Nil(t, got.Data)
		require.NotEmpty(t, got.Errs)

		assert.Equal(t, want, got.Errs)
	})
}

func expectData[Attrs any | []any](t *testing.T, r io.Reader, want allsrv.Data[Attrs]) {
	t.Helper()

	expectJSONBody(t, r, func(t *testing.T, got allsrv.RespResourceBody[Attrs]) {
		t.Helper()

		require.Empty(t, got.Errs)
		require.NotNil(t, got.Data)

		assert.Equal(t, want, *got.Data)
	})
}

func dbHasFoo(t *testing.T, db allsrv.DB, want allsrv.Foo) {
	t.Helper()

	got, err := db.ReadFoo(context.TODO(), want.ID)
	require.NoError(t, err)

	assert.Equal(t, want, got)
}

func createFoos(foos ...allsrv.Foo) func(t *testing.T, db allsrv.DB) {
	return func(t *testing.T, db allsrv.DB) {
		t.Helper()

		for _, f := range foos {
			err := db.CreateFoo(context.TODO(), f)
			require.NoError(t, err)
		}
	}
}

func newJSONReq(method, target string, body io.Reader, opts ...func(*http.Request)) *http.Request {
	req := httptest.NewRequest(method, target, body)
	req.Header.Set("Content-Type", "application/json")
	for _, o := range opts {
		o(req)
	}
	return req
}

func withBasicAuth(user, pass string) func(*http.Request) {
	return func(req *http.Request) {
		req.SetBasicAuth(user, pass)
	}
}

func newIDGen(start, incr int) func() string {
	return func() string {
		id := strconv.Itoa(start)
		start += incr
		return id
	}
}

func newNowFn(start time.Time, incr time.Duration) func() time.Time {
	return func() time.Time {
		t := start
		start = start.Add(incr)
		return t
	}
}
