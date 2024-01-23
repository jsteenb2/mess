package allsrv_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jsteenb2/mess/allsrv"
)

type dbInitFn func(t *testing.T) allsrv.DB

func testDB(t *testing.T, initFn dbInitFn) {
	t.Helper()

	tests := []struct {
		name string
		fn   func(t *testing.T, initFn dbInitFn)
	}{
		{
			name: "CreateFoo",
			fn:   testDBCreateFoo,
		},
		{
			name: "ReadFoo",
			fn:   testDBReadFoo,
		},
		{
			name: "UpdateFoo",
			fn:   testDBUpdateFoo,
		},
		{
			name: "DelFoo",
			fn:   testDBDeleteFoo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fn(t, initFn)
		})
	}
}

func testDBCreateFoo(t *testing.T, initFn dbInitFn) {
	t.Helper()

	type (
		inputs struct {
			foo allsrv.Foo
		}

		wantFn func(t *testing.T, db allsrv.DB, insertErr error)
	)

	start := time.Time{}.Add(time.Hour).UTC()

	tests := []struct {
		name    string
		prepare func(t *testing.T, db allsrv.DB)
		inputs  inputs
		want    wantFn
	}{
		{
			name: "with valid foo should pass",
			inputs: inputs{
				foo: allsrv.Foo{
					ID:        "1",
					Name:      "name",
					Note:      "note",
					CreatedAt: start,
					UpdatedAt: start,
				},
			},
			want: func(t *testing.T, db allsrv.DB, insertErr error) {
				require.NoError(t, insertErr)

				got, err := db.ReadFoo(context.TODO(), "1")
				require.NoError(t, err)

				want := allsrv.Foo{
					ID:        "1",
					Name:      "name",
					Note:      "note",
					CreatedAt: start,
					UpdatedAt: start,
				}
				assert.Equal(t, want, got)
			},
		},
		{
			name: "with concurrent valid foo creates should pass",
			prepare: func(t *testing.T, db allsrv.DB) {
				newFoo := func(id string) allsrv.Foo {
					return allsrv.Foo{
						ID:        id,
						Name:      "name-" + id,
						Note:      "note-" + id,
						CreatedAt: start.Add(time.Minute),
						UpdatedAt: start.Add(time.Minute),
					}
				}

				// execute while rest of test completes
				for _, f := range []allsrv.Foo{newFoo("1"), newFoo("2"), newFoo("3"), newFoo("4"), newFoo("5")} {
					go func(f allsrv.Foo) {
						require.NoError(t, db.CreateFoo(context.TODO(), f))
					}(f)
				}
			},
			inputs: inputs{
				foo: allsrv.Foo{
					ID:        "9000",
					Name:      "passing",
					Note:      "note",
					CreatedAt: start,
					UpdatedAt: start,
				},
			},
			want: func(t *testing.T, db allsrv.DB, insertErr error) {
				require.NoError(t, insertErr)

				got, err := db.ReadFoo(context.TODO(), "9000")
				require.NoError(t, err)

				want := allsrv.Foo{
					ID:        "9000",
					Name:      "passing",
					Note:      "note",
					CreatedAt: start,
					UpdatedAt: start,
				}
				assert.Equal(t, want, got)
			},
		},
		{
			name:    "with foo containing name that already exists should fail",
			prepare: createFoos(allsrv.Foo{ID: "1", Name: "collision"}),
			inputs: inputs{
				foo: allsrv.Foo{
					ID:        "2",
					Name:      "collision",
					Note:      "some note",
					CreatedAt: start,
					UpdatedAt: start,
				},
			},
			want: func(t *testing.T, db allsrv.DB, insertErr error) {
				require.Error(t, insertErr)
			},
		},
		// {
		// 	name:    "with foo containing ID that already exists should fail",
		// 	prepare: prepDBFoos(allsrv.Foo{ID: "1", Name: "name-1"}),
		// 	inputs: inputs{
		// 		foo: allsrv.Foo{
		// 			ID:        "1",
		// 			Name:      "name-2",
		// 			Note:      "some note",
		// 			CreatedAt: start,
		// 			UpdatedAt: start,
		// 		},
		// 	},
		// 	want: func(t *testing.T, db allsrv.DB, insertErr error) {
		// 		require.Error(t, insertErr)
		// 	},
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// setup
			db := initFn(t)
			if tt.prepare != nil {
				tt.prepare(t, db)
			}

			// action
			insertErr := db.CreateFoo(context.TODO(), tt.inputs.foo)

			// assert
			tt.want(t, db, insertErr)
		})
	}
}

func testDBReadFoo(t *testing.T, initFn dbInitFn) {
	t.Helper()

	type (
		inputs struct {
			id string
		}

		wantFn func(t *testing.T, got allsrv.Foo, readErr error)
	)

	start := time.Time{}.Add(time.Hour).UTC()

	tests := []struct {
		name    string
		prepare func(t *testing.T, db allsrv.DB)
		inputs  inputs
		want    wantFn
	}{
		{
			name: "with id for existing foo should pass",
			prepare: createFoos(allsrv.Foo{
				ID:        "1",
				Name:      "name-1",
				Note:      "note-1",
				CreatedAt: start,
				UpdatedAt: start.Add(time.Hour),
			}),
			inputs: inputs{
				id: "1",
			},
			want: func(t *testing.T, got allsrv.Foo, readErr error) {
				require.NoError(t, readErr)

				want := allsrv.Foo{
					ID:        "1",
					Name:      "name-1",
					Note:      "note-1",
					CreatedAt: start,
					UpdatedAt: start.Add(time.Hour),
				}
				assert.Equal(t, want, got)
			},
		},
		{
			name: "with concurrent valid foo update the reading should pass",
			prepare: func(t *testing.T, db allsrv.DB) {
				err := db.CreateFoo(context.TODO(), allsrv.Foo{
					ID:        "1",
					Name:      "one",
					Note:      "note",
					CreatedAt: start,
					UpdatedAt: start,
				})
				require.NoError(t, err)

				newFoo := func(note string) allsrv.Foo {
					return allsrv.Foo{
						ID:        "1",
						Name:      "one",
						Note:      note,
						CreatedAt: start,
						UpdatedAt: start.Add(time.Hour),
					}
				}

				// execute these while test is executing read
				for _, f := range []allsrv.Foo{newFoo("a"), newFoo("b"), newFoo("c"), newFoo("d"), newFoo("e")} {
					go func(f allsrv.Foo) {
						require.NoError(t, db.UpdateFoo(context.TODO(), f))
					}(f)
				}
			},
			inputs: inputs{
				id: "1",
			},
			want: func(t *testing.T, got allsrv.Foo, readErr error) {
				require.NoError(t, readErr)
				assert.Contains(t, []string{"note", "a", "b", "c", "d", "e"}, got.Note)
			},
		},
		{
			name: "with id for non-existent foo should fail",
			inputs: inputs{
				id: "1",
			},
			want: func(t *testing.T, _ allsrv.Foo, readErr error) {
				// this is pretty gross, we're matching against a raw error/text value
				// any change in the error message means we have to update tests too
				want := errors.New("foo not found for id: 1")
				assert.Equal(t, want.Error(), readErr.Error())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// setup
			db := initFn(t)
			if tt.prepare != nil {
				tt.prepare(t, db)
			}

			// action
			got, err := db.ReadFoo(context.TODO(), tt.inputs.id)

			// assert
			tt.want(t, got, err)
		})
	}
}

func testDBUpdateFoo(t *testing.T, initFn dbInitFn) {
	type (
		inputs struct {
			foo allsrv.Foo
		}

		wantFn func(t *testing.T, db allsrv.DB, updateErr error)
	)

	start := time.Time{}.Add(time.Hour).UTC()

	tests := []struct {
		name    string
		prepare func(t *testing.T, db allsrv.DB)
		inputs  inputs
		want    wantFn
	}{
		{
			name: "with valid update for existing foo should pass",
			prepare: createFoos(allsrv.Foo{
				ID:        "1",
				Name:      "name",
				Note:      "note",
				CreatedAt: start,
				UpdatedAt: start,
			}),
			inputs: inputs{
				foo: allsrv.Foo{
					ID:        "1",
					Name:      "name",
					Note:      "some other note",
					CreatedAt: start,
					UpdatedAt: start.Add(time.Hour),
				},
			},
			want: func(t *testing.T, db allsrv.DB, updateErr error) {
				require.NoError(t, updateErr)

				got, err := db.ReadFoo(context.TODO(), "1")
				require.NoError(t, err)

				want := allsrv.Foo{
					ID:        "1",
					Name:      "name",
					Note:      "some other note",
					CreatedAt: start,
					UpdatedAt: start.Add(time.Hour),
				}
				assert.Equal(t, want, got)
			},
		},
		{
			name: "with concurrent valid foo updates should pass",
			prepare: func(t *testing.T, db allsrv.DB) {
				require.NoError(t, db.CreateFoo(context.TODO(), allsrv.Foo{
					ID:        "1",
					Name:      "one",
					Note:      "note",
					CreatedAt: start,
					UpdatedAt: start,
				}))

				newFoo := func(note string) allsrv.Foo {
					return allsrv.Foo{
						ID:        "1",
						Name:      "one",
						Note:      note,
						UpdatedAt: start.Add(time.Hour),
					}
				}

				for _, f := range []allsrv.Foo{newFoo("a"), newFoo("b"), newFoo("c"), newFoo("d"), newFoo("e")} {
					go func(f allsrv.Foo) {
						require.NoError(t, db.UpdateFoo(context.TODO(), f))
					}(f)
				}
			},
			inputs: inputs{
				foo: allsrv.Foo{
					ID:        "1",
					Name:      "one",
					Note:      "final",
					UpdatedAt: start.Add(24 * time.Hour),
				},
			},
			want: func(t *testing.T, db allsrv.DB, updateErr error) {
				require.NoError(t, updateErr)

				got, err := db.ReadFoo(context.TODO(), "1")
				require.NoError(t, err)

				assert.Contains(t, []string{"final", "note", "a", "b", "c", "d", "e"}, got.Note)
			},
		},
		{
			name: "with update for non-existent foo should fail",
			inputs: inputs{
				foo: allsrv.Foo{
					ID:        "1",
					Name:      "name",
					Note:      "note",
					CreatedAt: start,
					UpdatedAt: start.Add(time.Hour),
				},
			},
			want: func(t *testing.T, db allsrv.DB, updateErr error) {
				require.Error(t, updateErr)

				// this is pretty gross, we're matching against a raw error/text value
				// any change in the error message means we have to update tests too
				want := errors.New("foo not found for id: 1")
				assert.Equal(t, want.Error(), updateErr.Error())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// setup
			db := initFn(t)
			if tt.prepare != nil {
				tt.prepare(t, db)
			}

			// action
			err := db.UpdateFoo(context.TODO(), tt.inputs.foo)

			// assert
			tt.want(t, db, err)
		})
	}
}

func testDBDeleteFoo(t *testing.T, initFn dbInitFn) {
	t.Helper()

	type (
		inputs struct {
			id string
		}

		wantFn func(t *testing.T, db allsrv.DB, delErr error)
	)

	start := time.Time{}.Add(time.Hour).UTC()

	tests := []struct {
		name    string
		prepare func(t *testing.T, db allsrv.DB)
		inputs  inputs
		want    wantFn
	}{
		{
			name:    "with id for existing foo should pass",
			prepare: createFoos(allsrv.Foo{ID: "1", Name: "blue"}),
			inputs: inputs{
				id: "1",
			},
			want: func(t *testing.T, db allsrv.DB, delErr error) {
				require.NoError(t, delErr)

				_, err := db.ReadFoo(context.TODO(), "1")

				// this is pretty gross, we're matching against a raw error/text value
				// any change in the error message means we have to update tests too
				want := errors.New("foo not found for id: 1")
				assert.Equal(t, want.Error(), err.Error())
			},
		},
		{
			name: "with concurrent valid foo creates should pass",
			prepare: func(t *testing.T, db allsrv.DB) {
				newFoo := func(id string) allsrv.Foo {
					return allsrv.Foo{
						ID:        id,
						Name:      "name-" + id,
						Note:      "note-" + id,
						CreatedAt: start,
						UpdatedAt: start,
					}
				}

				for _, f := range []allsrv.Foo{newFoo("1"), newFoo("2"), newFoo("3"), newFoo("4"), newFoo("5")} {
					require.NoError(t, db.CreateFoo(context.TODO(), f))
				}

				// leave the foo "1" del for the input
				for _, id := range []string{"2", "3", "4", "5"} {
					go func(id string) {
						require.NoError(t, db.DelFoo(context.TODO(), id))
					}(id)
				}
			},
			inputs: inputs{id: "1"},
			want: func(t *testing.T, db allsrv.DB, delErr error) {
				require.NoError(t, delErr)
			},
		},
		{
			name:   "with id for non-existent foo should fail",
			inputs: inputs{id: "1"},
			want: func(t *testing.T, db allsrv.DB, delErr error) {
				require.Error(t, delErr)

				// this is pretty gross, we're matching against a raw error/text value
				// any change in the error message means we have to update tests too
				want := errors.New("foo not found for id: 1")
				assert.Equal(t, want.Error(), delErr.Error())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// setup
			db := initFn(t)
			if tt.prepare != nil {
				tt.prepare(t, db)
			}

			// action
			delErr := db.DelFoo(context.TODO(), tt.inputs.id)

			// assert
			tt.want(t, db, delErr)
		})
	}
}
