package allsrvtesting

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jsteenb2/mess/allsrv"
)

var start = time.Time{}.Add(time.Hour).UTC()

type (
	SVCInitFn func(t *testing.T, opts SVCTestOpts) SVCDeps

	SVCDeps struct {
		SVC allsrv.SVC
	}

	SVCTestOpts struct {
		PrepDB  func(t *testing.T, db allsrv.DB)
		SVCOpts []func(svc *allsrv.Service)
	}
)

func TestSVC(t *testing.T, initFn SVCInitFn) {
	tests := []struct {
		name   string
		testFn func(t *testing.T, initFn SVCInitFn)
	}{
		{name: "Create", testFn: testSVCCreate},
		{name: "Read", testFn: testSVCRead},
		{name: "Update", testFn: testSVCUpdate},
		{name: "Delete", testFn: testSVCDel},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.testFn(t, initFn)
		})
	}
}

func testSVCCreate(t *testing.T, initFn SVCInitFn) {
	type (
		inputs struct {
			foo allsrv.Foo
		}

		wantFn func(t *testing.T, newFoo allsrv.Foo, insertErr error)
	)

	tests := []struct {
		name  string
		opts  SVCTestOpts
		input inputs
		want  wantFn
	}{
		{
			name: "with valid foo should pass",
			input: inputs{
				foo: allsrv.Foo{
					Name: "first_foo",
					Note: "first note",
				},
			},
			want: func(t *testing.T, newFoo allsrv.Foo, insertErr error) {
				wantFoo(allsrv.Foo{
					ID:        "1",
					Name:      "first_foo",
					Note:      "first note",
					CreatedAt: start,
					UpdatedAt: start,
				})
			},
		},
		{
			name: "with valid foo missing note should pass",
			input: inputs{
				foo: allsrv.Foo{
					Name: "first_foo",
					Note: "",
				},
			},
			want: func(t *testing.T, newFoo allsrv.Foo, insertErr error) {
				wantFoo(allsrv.Foo{
					ID:        "1",
					Name:      "first_foo",
					Note:      "",
					CreatedAt: start,
					UpdatedAt: start,
				})
			},
		},
		{
			name: "with foo with conflicting name should fail",
			opts: SVCTestOpts{
				PrepDB: CreateFoos(allsrv.Foo{ID: "9000", Name: "existing-foo"}),
			},
			input: inputs{
				foo: allsrv.Foo{
					Name: "existing-foo",
					Note: "new note",
				},
			},
			want: func(t *testing.T, _ allsrv.Foo, insertErr error) {
				require.Error(t, insertErr)
				assert.True(t, allsrv.IsExistsErr(insertErr))
			},
		},
		{
			name: "with foo missing name should fail",
			input: inputs{
				foo: allsrv.Foo{
					Name: "",
					Note: "note",
				},
			},
			want: func(t *testing.T, _ allsrv.Foo, insertErr error) {
				require.Error(t, insertErr)
				assert.True(t, allsrv.IsInvalidErr(insertErr))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// setup
			deps := initFn(t, withTestOptions(tt.opts))

			// action
			got, err := deps.SVC.CreateFoo(context.TODO(), tt.input.foo)

			// assert
			tt.want(t, got, err)
		})
	}
}

func testSVCRead(t *testing.T, initFn SVCInitFn) {
	type (
		inputs struct {
			id string
		}

		wantFn func(t *testing.T, got allsrv.Foo, readErr error)
	)

	var (
		ninekFoo = allsrv.Foo{
			ID:        "9000",
			Name:      "goku",
			Note:      "displeasing to some",
			CreatedAt: start,
			UpdatedAt: start,
		}

		fooTwo = allsrv.Foo{
			ID:        "2",
			Name:      "twoscompany",
			Note:      "some note",
			CreatedAt: start,
			UpdatedAt: start.Add(300 * time.Hour),
		}
	)

	tests := []struct {
		name    string
		options SVCTestOpts
		input   inputs
		want    wantFn
	}{
		{
			name: "with existing id should pass",
			options: SVCTestOpts{
				PrepDB: CreateFoos(ninekFoo, fooTwo),
			},
			input: inputs{
				id: ninekFoo.ID,
			},
			want: func(t *testing.T, got allsrv.Foo, readErr error) {
				wantFoo(ninekFoo)
			},
		},
		{
			name: "with another existing id should pass",
			options: SVCTestOpts{
				PrepDB: CreateFoos(ninekFoo, fooTwo),
			},
			input: inputs{
				id: fooTwo.ID,
			},
			want: func(t *testing.T, got allsrv.Foo, readErr error) {
				wantFoo(fooTwo)
			},
		},
		{
			name: "with an empty string id should fail",
			input: inputs{
				id: "",
			},
			want: func(t *testing.T, got allsrv.Foo, readErr error) {
				require.Error(t, readErr)
				assert.True(t, allsrv.IsInvalidErr(readErr), "got_err="+readErr.Error())
			},
		},
		{
			name: "with id for non-existent foo should fail",
			input: inputs{
				id: "NOTFOUND",
			},
			want: func(t *testing.T, got allsrv.Foo, readErr error) {
				require.Error(t, readErr)
				assert.True(t, allsrv.IsNotFoundErr(readErr))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// setup
			deps := initFn(t, withTestOptions(tt.options))

			// action
			got, err := deps.SVC.ReadFoo(context.TODO(), tt.input.id)

			// assert
			tt.want(t, got, err)
		})
	}
}

func testSVCUpdate(t *testing.T, initFn SVCInitFn) {
	type (
		inputs struct {
			upd allsrv.FooUpd
		}

		wantFn func(t *testing.T, updatedFoo allsrv.Foo, updErr error)
	)

	tests := []struct {
		name  string
		opts  SVCTestOpts
		input inputs
		want  wantFn
	}{
		{
			name: "with valid full update of existing foo should pass",
			opts: SVCTestOpts{
				PrepDB: CreateFoos(allsrv.Foo{
					ID:        "1",
					Name:      "first_foo",
					Note:      "first note",
					CreatedAt: start.Add(-time.Minute),
					UpdatedAt: start.Add(-time.Minute),
				}),
			},
			input: inputs{
				upd: allsrv.FooUpd{
					ID:   "1",
					Name: Ptr("updated_foo"),
					Note: Ptr("updated note"),
				},
			},
			want: func(t *testing.T, updatedFoo allsrv.Foo, updErr error) {
				wantFoo(allsrv.Foo{
					ID:        "1",
					Name:      "updated_foo",
					Note:      "updated note",
					CreatedAt: start.Add(-time.Minute),
					UpdatedAt: start,
				})
			},
		},
		{
			name: "with valid name only update of existing foo should pass",
			opts: SVCTestOpts{
				PrepDB: CreateFoos(allsrv.Foo{
					ID:        "1",
					Name:      "first_foo",
					Note:      "first note",
					CreatedAt: start.Add(-time.Minute),
					UpdatedAt: start.Add(-time.Minute),
				}),
			},
			input: inputs{
				upd: allsrv.FooUpd{
					ID:   "1",
					Name: Ptr("updated_foo"),
				},
			},
			want: func(t *testing.T, updatedFoo allsrv.Foo, updErr error) {
				wantFoo(allsrv.Foo{
					ID:        "1",
					Name:      "updated_foo",
					Note:      "first note",
					CreatedAt: start.Add(-time.Minute),
					UpdatedAt: start,
				})
			},
		},
		{
			name: "with valid note only update of existing foo should pass",
			opts: SVCTestOpts{
				PrepDB: CreateFoos(allsrv.Foo{
					ID:        "1",
					Name:      "first_foo",
					Note:      "first note",
					CreatedAt: start.Add(-time.Minute),
					UpdatedAt: start.Add(-time.Minute),
				}),
			},
			input: inputs{
				upd: allsrv.FooUpd{
					ID:   "1",
					Note: Ptr("updated note"),
				},
			},
			want: func(t *testing.T, updatedFoo allsrv.Foo, updErr error) {
				wantFoo(allsrv.Foo{
					ID:        "1",
					Name:      "first_foo",
					Note:      "updated note",
					CreatedAt: start.Add(-time.Minute),
					UpdatedAt: start,
				})
			},
		},
		{
			name: "with update of non-existent foo should fail",
			input: inputs{
				upd: allsrv.FooUpd{
					ID:   "1",
					Note: Ptr("updated note"),
				},
			},
			want: func(t *testing.T, updatedFoo allsrv.Foo, updErr error) {
				require.Error(t, updErr)
				assert.True(t, allsrv.IsNotFoundErr(updErr))
			},
		},
		{
			name: "when updating foo too a name that collides with existing should fail",
			opts: SVCTestOpts{
				PrepDB: CreateFoos(allsrv.Foo{ID: "1", Name: "start-foo"}, allsrv.Foo{ID: "9000", Name: "existing-foo"}),
			},
			input: inputs{
				upd: allsrv.FooUpd{
					ID:   "1",
					Name: Ptr("existing-foo"),
					Note: Ptr("some note"),
				},
			},
			want: func(t *testing.T, updatedFoo allsrv.Foo, updErr error) {
				require.Error(t, updErr)
				assert.True(t, allsrv.IsExistsErr(updErr))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// setup
			deps := initFn(t, withTestOptions(tt.opts))

			// action
			got, err := deps.SVC.UpdateFoo(context.TODO(), tt.input.upd)

			// assert
			tt.want(t, got, err)
		})
	}
}

func testSVCDel(t *testing.T, initFn SVCInitFn) {
	type (
		inputs struct {
			id string
		}

		wantFn func(t *testing.T, svc allsrv.SVC, delErr error)
	)

	tests := []struct {
		name    string
		options SVCTestOpts
		input   inputs
		want    wantFn
	}{
		{
			name: "with id for existing foo should pass",
			options: SVCTestOpts{
				PrepDB: CreateFoos(allsrv.Foo{ID: "9000", Name: "goku"}),
			},
			input: inputs{
				id: "9000",
			},
			want: func(t *testing.T, svc allsrv.SVC, delErr error) {
				require.NoError(t, delErr)

				_, err := svc.ReadFoo(context.TODO(), "9000")
				require.Error(t, err)
				assert.True(t, allsrv.IsNotFoundErr(err))
			},
		},
		{
			name: "with id for non-existent foo should fail",
			input: inputs{
				id: "9000",
			},
			want: func(t *testing.T, svc allsrv.SVC, delErr error) {
				require.Error(t, delErr)
				assert.True(t, allsrv.IsNotFoundErr(delErr))
			},
		},
		{
			name: "without id should fail",
			input: inputs{
				id: "",
			},
			want: func(t *testing.T, svc allsrv.SVC, delErr error) {
				require.Error(t, delErr)
				assert.True(t, allsrv.IsInvalidErr(delErr), "got_err="+delErr.Error())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// setup
			deps := initFn(t, withTestOptions(tt.options))

			// action
			err := deps.SVC.DelFoo(context.TODO(), tt.input.id)

			// assert
			tt.want(t, deps.SVC, err)
		})
	}
}

// withTestOptions provides some sane default values for tests.
func withTestOptions(opts SVCTestOpts) SVCTestOpts {
	if opts.PrepDB == nil {
		opts.PrepDB = func(t *testing.T, db allsrv.DB) {}
	}
	// purposefully checking nil here, empty slice indicates no options
	if opts.SVCOpts == nil {
		opts.SVCOpts = DefaultSVCOpts(start)
	}
	return opts
}

func DefaultSVCOpts(start time.Time) []func(*allsrv.Service) {
	return []func(*allsrv.Service){
		allsrv.WithSVCIDFn(IDGen(1, 1)),
		allsrv.WithSVCNowFn(NowFn(start, time.Hour)),
	}
}

func wantFoo(want allsrv.Foo) func(t *testing.T, got allsrv.Foo, opErr error) {
	return func(t *testing.T, got allsrv.Foo, opErr error) {
		t.Helper()

		require.NoError(t, opErr)
		assert.Equal(t, want, got)
	}
}

func CreateFoos(foos ...allsrv.Foo) func(t *testing.T, db allsrv.DB) {
	return func(t *testing.T, db allsrv.DB) {
		t.Helper()

		for _, f := range foos {
			err := db.CreateFoo(context.TODO(), f)
			require.NoError(t, err)
		}
	}
}
