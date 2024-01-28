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
	type (
		inputs struct {
			req *http.Request
		}

		wantFn func(t *testing.T, rec *httptest.ResponseRecorder, db allsrv.DB)

		testCase struct {
			name    string
			prepare func(t *testing.T, db allsrv.DB)
			svcOpts []func(*allsrv.Service)
			svrOpts []allsrv.SvrOptFn
			inputs  inputs
			want    wantFn
		}
	)

	start := time.Time{}.Add(time.Hour).UTC()

	testSvr := func(t *testing.T, tt testCase) {
		db := new(allsrv.InmemDB)

		if tt.prepare != nil {
			tt.prepare(t, db)
		}

		svcOpts := append(defaultSVCOpts(start), tt.svcOpts...)
		svc := allsrv.NewService(db, svcOpts...)

		defaultSvrOpts := []allsrv.SvrOptFn{allsrv.WithMetrics(newTestMetrics(t))}
		svrOpts := append(defaultSvrOpts, tt.svrOpts...)

		rec := httptest.NewRecorder()

		svr := allsrv.NewServerV2(svc, svrOpts...)
		svr.ServeHTTP(rec, tt.inputs.req)

		tt.want(t, rec, db)
	}

	t.Run("foo create", func(t *testing.T) {
		tests := []testCase{
			{
				name:    "when provided a valid foo and authorized user should pass",
				svrOpts: []allsrv.SvrOptFn{allsrv.WithBasicAuthV2("dodgers@stink.com", "PaSsWoRd")},
				inputs: inputs{
					req: newJSONReq("POST", "/v1/foos",
						newJSONBody(t, allsrv.ReqCreateFooV1{
							Data: allsrv.Data[allsrv.FooCreateAttrs]{
								Type: "foo",
								Attrs: allsrv.FooCreateAttrs{
									Name: "first-foo",
									Note: "some note",
								},
							},
						}),
						withBasicAuth("dodgers@stink.com", "PaSsWoRd"),
					),
				},
				want: func(t *testing.T, rec *httptest.ResponseRecorder, db allsrv.DB) {
					assert.Equal(t, http.StatusCreated, rec.Code)
					expectData[allsrv.ResourceFooAttrs](t, rec.Body, allsrv.Data[allsrv.ResourceFooAttrs]{
						Type: "foo",
						ID:   "1",
						Attrs: allsrv.ResourceFooAttrs{
							Name:      "first-foo",
							Note:      "some note",
							CreatedAt: start.Format(time.RFC3339),
							UpdatedAt: start.Format(time.RFC3339),
						},
					})

					dbHasFoo(t, db, allsrv.Foo{
						ID:        "1",
						Name:      "first-foo",
						Note:      "some note",
						CreatedAt: start,
						UpdatedAt: start,
					})
				},
			},
			{
				name:    "when missing required auth should fail",
				svrOpts: []allsrv.SvrOptFn{allsrv.WithBasicAuthV2("dodgers@stink.com", "PaSsWoRd")},
				inputs: inputs{
					req: newJSONReq("POST", "/v1/foos",
						newJSONBody(t, allsrv.ReqCreateFooV1{
							Data: allsrv.Data[allsrv.FooCreateAttrs]{
								Type: "foo",
								Attrs: allsrv.FooCreateAttrs{
									Name: "first-foo",
									Note: "some note",
								},
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
						Data: allsrv.Data[allsrv.FooCreateAttrs]{
							Type: "foo",
							Attrs: allsrv.FooCreateAttrs{
								Name: "existing-foo",
								Note: "some note",
							},
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
						Data: allsrv.Data[allsrv.FooCreateAttrs]{
							Type: "WRONGO",
							Attrs: allsrv.FooCreateAttrs{
								Name: "first-foo",
								Note: "some note",
							},
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
				testSvr(t, tt)
			})
		}
	})

	t.Run("foo read", func(t *testing.T) {
		tests := []testCase{
			{
				name: "with authorized user for existing foo should pass",
				prepare: createFoos(allsrv.Foo{
					ID:        "1",
					Name:      "first-foo",
					Note:      "some note",
					CreatedAt: start,
					UpdatedAt: start,
				}),
				svrOpts: []allsrv.SvrOptFn{allsrv.WithBasicAuthV2("dodogers@fire.dumpster", "truth")},
				inputs: inputs{
					req: get("/v1/foos/1", withBasicAuth("dodogers@fire.dumpster", "truth")),
				},
				want: func(t *testing.T, rec *httptest.ResponseRecorder, _ allsrv.DB) {
					assert.Equal(t, http.StatusOK, rec.Code)
					expectData[allsrv.ResourceFooAttrs](t, rec.Body, allsrv.Data[allsrv.ResourceFooAttrs]{
						Type: "foo",
						ID:   "1",
						Attrs: allsrv.ResourceFooAttrs{
							Name:      "first-foo",
							Note:      "some note",
							CreatedAt: start.Format(time.RFC3339),
							UpdatedAt: start.Format(time.RFC3339),
						},
					})
				},
			},
			{
				name: "with unauthorized user for existing foo should fail",
				prepare: createFoos(allsrv.Foo{
					ID:        "1",
					Name:      "first-foo",
					Note:      "some note",
					CreatedAt: start,
				}),
				svrOpts: []allsrv.SvrOptFn{allsrv.WithBasicAuthV2("dodogers@fire.dumpster", "truth")},
				inputs: inputs{
					req: get("/v1/foos/1", withBasicAuth("dodogers@are.exellence", "false")),
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
				},
			},
			{
				name: "with request for non-existent foo should fail",
				inputs: inputs{
					req: get("/v1/foos/1"),
				},
				want: func(t *testing.T, rec *httptest.ResponseRecorder, _ allsrv.DB) {
					assert.Equal(t, http.StatusNotFound, rec.Code)
					expectErrs(t, rec.Body, allsrv.RespErr{
						Status: http.StatusNotFound,
						Code:   3,
						Msg:    "foo not found for id: 1",
					})
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				testSvr(t, tt)
			})
		}
	})

	t.Run("foo update", func(t *testing.T) {
		tests := []testCase{
			{
				name: "when provided a full valid update and authorized user should pass",
				prepare: createFoos(allsrv.Foo{
					ID:        "1",
					Name:      "first-foo",
					Note:      "some note",
					CreatedAt: start,
				}),
				svcOpts: []func(*allsrv.Service){allsrv.WithSVCNowFn(nowFn(start.Add(time.Hour), time.Hour))},
				svrOpts: []allsrv.SvrOptFn{allsrv.WithBasicAuthV2("dodgers@stink.com", "PaSsWoRd")},
				inputs: inputs{
					req: newJSONReq("PATCH", "/v1/foos/1",
						newJSONBody(t, allsrv.ReqUpdateFooV1{
							Data: allsrv.Data[allsrv.FooUpdAttrs]{
								Type: "foo",
								ID:   "1",
								Attrs: allsrv.FooUpdAttrs{
									Name: ptr("new-name"),
									Note: ptr("new note"),
								},
							},
						}),
						withBasicAuth("dodgers@stink.com", "PaSsWoRd"),
					),
				},
				want: func(t *testing.T, rec *httptest.ResponseRecorder, db allsrv.DB) {
					assert.Equal(t, http.StatusOK, rec.Code)
					expectData[allsrv.ResourceFooAttrs](t, rec.Body, allsrv.Data[allsrv.ResourceFooAttrs]{
						Type: "foo",
						ID:   "1",
						Attrs: allsrv.ResourceFooAttrs{
							Name:      "new-name",
							Note:      "new note",
							CreatedAt: start.Format(time.RFC3339),
							UpdatedAt: start.Add(time.Hour).Format(time.RFC3339),
						},
					})

					dbHasFoo(t, db, allsrv.Foo{
						ID:        "1",
						Name:      "new-name",
						Note:      "new note",
						CreatedAt: start,
						UpdatedAt: start.Add(time.Hour),
					})
				},
			},
			{
				name: "when provided a name only update should pass",
				prepare: createFoos(allsrv.Foo{
					ID:        "1",
					Name:      "first-name",
					CreatedAt: start,
				}),
				svcOpts: []func(*allsrv.Service){allsrv.WithSVCNowFn(nowFn(start.Add(time.Hour), time.Hour))},
				inputs: inputs{
					req: newJSONReq("PATCH", "/v1/foos/1",
						newJSONBody(t, allsrv.ReqUpdateFooV1{
							Data: allsrv.Data[allsrv.FooUpdAttrs]{
								Type: "foo",
								ID:   "1",
								Attrs: allsrv.FooUpdAttrs{
									Note: ptr("new note"),
								},
							},
						}),
						withBasicAuth("dodgers@stink.com", "PaSsWoRd"),
					),
				},
				want: func(t *testing.T, rec *httptest.ResponseRecorder, db allsrv.DB) {
					assert.Equal(t, http.StatusOK, rec.Code)
					expectData[allsrv.ResourceFooAttrs](t, rec.Body, allsrv.Data[allsrv.ResourceFooAttrs]{
						Type: "foo",
						ID:   "1",
						Attrs: allsrv.ResourceFooAttrs{
							Name:      "first-name",
							Note:      "new note",
							CreatedAt: start.Format(time.RFC3339),
							UpdatedAt: start.Add(time.Hour).Format(time.RFC3339),
						},
					})

					dbHasFoo(t, db, allsrv.Foo{
						ID:        "1",
						Name:      "first-name",
						Note:      "new note",
						CreatedAt: start,
						UpdatedAt: start.Add(time.Hour),
					})
				},
			},
			{
				name: "when missing required auth should fail",
				prepare: createFoos(allsrv.Foo{
					ID:        "1",
					Name:      "first-foo",
					Note:      "some note",
					CreatedAt: start,
				}),
				svrOpts: []allsrv.SvrOptFn{allsrv.WithBasicAuthV2("dodgers@stink.com", "PaSsWoRd")},
				inputs: inputs{
					req: newJSONReq("PATCH", "/v1/foos/1",
						newJSONBody(t, allsrv.ReqUpdateFooV1{
							Data: allsrv.Data[allsrv.FooUpdAttrs]{
								Type: "foo",
								ID:   "1",
								Attrs: allsrv.FooUpdAttrs{
									Note: ptr("new note"),
								},
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
				},
			},
			{
				name:    "when updating foo too a name that collides with existing should fail",
				prepare: createFoos(allsrv.Foo{ID: "1", Name: "start-foo"}, allsrv.Foo{ID: "9000", Name: "existing-foo"}),
				inputs: inputs{
					req: newJSONReq("PATCH", "/v1/foos/1", newJSONBody(t, allsrv.ReqUpdateFooV1{
						Data: allsrv.Data[allsrv.FooUpdAttrs]{
							Type: "foo",
							ID:   "1",
							Attrs: allsrv.FooUpdAttrs{
								Name: ptr("existing-foo"),
								Note: ptr("some note"),
							},
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

					dbHasFoo(t, db, allsrv.Foo{
						ID:   "1",
						Name: "start-foo",
					})
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				testSvr(t, tt)
			})
		}
	})

	t.Run("foo delete", func(t *testing.T) {
		tests := []testCase{
			{
				name: "with authorized user for existing foo should pass",
				prepare: createFoos(allsrv.Foo{
					ID:        "1",
					Name:      "first-foo",
					Note:      "some note",
					CreatedAt: start,
				}),
				svrOpts: []allsrv.SvrOptFn{allsrv.WithBasicAuthV2("dodogers@fire.dumpster", "truth")},
				inputs: inputs{
					req: del("/v1/foos/1", withBasicAuth("dodogers@fire.dumpster", "truth")),
				},
				want: func(t *testing.T, rec *httptest.ResponseRecorder, db allsrv.DB) {
					assert.Equal(t, http.StatusOK, rec.Code)
					expectJSONBody(t, rec.Body, func(t *testing.T, got allsrv.RespBody[any]) {
						require.Nil(t, got.Data)
						require.Nil(t, got.Errs)
						require.NotZero(t, got.Meta.TraceID)
					})

					_, err := db.ReadFoo(context.TODO(), "1")
					require.Error(t, err)
				},
			},
			{
				name: "with unauthorized user for existing foo should fail",
				prepare: createFoos(allsrv.Foo{
					ID:        "1",
					Name:      "first-foo",
					Note:      "some note",
					CreatedAt: start,
				}),
				svrOpts: []allsrv.SvrOptFn{allsrv.WithBasicAuthV2("dodogers@fire.dumpster", "truth")},
				inputs: inputs{
					req: del("/v1/foos/1", withBasicAuth("dodogers@are.exellence", "false")),
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
				},
			},
			{
				name: "with request for non-existent foo should fail",
				inputs: inputs{
					req: del("/v1/foos/1"),
				},
				want: func(t *testing.T, rec *httptest.ResponseRecorder, _ allsrv.DB) {
					assert.Equal(t, http.StatusNotFound, rec.Code)
					expectErrs(t, rec.Body, allsrv.RespErr{
						Status: http.StatusNotFound,
						Code:   3,
						Msg:    "foo not found for id: 1",
					})
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				testSvr(t, tt)
			})
		}
	})
}

func expectErrs(t *testing.T, r io.Reader, want ...allsrv.RespErr) {
	t.Helper()

	expectJSONBody(t, r, func(t *testing.T, got allsrv.RespBody[any]) {
		t.Helper()

		require.Nil(t, got.Data)
		require.NotEmpty(t, got.Errs)

		assert.Equal(t, want, got.Errs)
	})
}

func expectData[Attrs any | []any](t *testing.T, r io.Reader, want allsrv.Data[Attrs]) {
	t.Helper()

	expectJSONBody(t, r, func(t *testing.T, got allsrv.RespBody[Attrs]) {
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
	opts = append([]func(*http.Request){withContentType("application/json")}, opts...)
	return newReq(method, target, body, opts...)
}

func del(target string, opts ...func(*http.Request)) *http.Request {
	return newReq("DELETE", target, nil, opts...)
}

func get(target string, opts ...func(*http.Request)) *http.Request {
	return newReq("GET", target, nil, opts...)
}

func newReq(method, target string, body io.Reader, opts ...func(r *http.Request)) *http.Request {
	r := httptest.NewRequest(method, target, body)
	for _, o := range opts {
		o(r)
	}
	return r
}

func withContentType(ct string) func(*http.Request) {
	return func(r *http.Request) {
		r.Header.Set("Content-Type", ct)
	}
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

func nowFn(start time.Time, incr time.Duration) func() time.Time {
	return func() time.Time {
		t := start
		start = start.Add(incr)
		return t
	}
}

func ptr[T any](v T) *T {
	return &v
}
