package allsrv_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jsteenb2/mess/allsrv"
)

func TestInmemDB(t *testing.T) {
	t.Run("create foo", func(t *testing.T) {
		t.Run("with valid foo should pass", func(t *testing.T) {
			db := new(allsrv.InmemDB)

			want := allsrv.Foo{
				ID:   "1",
				Name: "name",
				Note: "note",
			}
			err := db.CreateFoo(context.TODO(), want)
			require.NoError(t, err)

			got, err := db.ReadFoo(context.TODO(), "1")
			require.NoError(t, err)

			assert.Equal(t, want, got)
		})

		t.Run("with concurrent valid foo creates should pass", func(t *testing.T) {
			db := new(allsrv.InmemDB)

			newFoo := func(id string) allsrv.Foo {
				return allsrv.Foo{
					ID:   id,
					Name: "name-" + id,
					Note: "note-" + id,
				}
			}

			var wg sync.WaitGroup
			for _, f := range []allsrv.Foo{newFoo("1"), newFoo("2"), newFoo("3"), newFoo("4"), newFoo("5")} {
				wg.Add(1)
				go func(f allsrv.Foo) {
					defer wg.Done()
					require.NoError(t, db.CreateFoo(context.TODO(), f))
				}(f)
			}
			wg.Wait()
		})

		t.Run("with foo containing name that already exists should fail", func(t *testing.T) {
			db := new(allsrv.InmemDB)

			want := allsrv.Foo{
				ID:   "1",
				Name: "collision",
				Note: "note",
			}
			err := db.CreateFoo(context.TODO(), want)
			require.NoError(t, err)

			err = db.CreateFoo(context.TODO(), want)

			// this is pretty gross, we're matching against a raw error/text value
			// any change in the error message means we have to update tests too
			wantErr := errors.New("foo collision exists")
			assert.Equal(t, wantErr, err)
		})
	})

	t.Run("read foo", func(t *testing.T) {
		t.Run("with id for existing foo should pass", func(t *testing.T) {
			db := new(allsrv.InmemDB)

			want := allsrv.Foo{
				ID:   "1",
				Name: "name",
				Note: "note",
			}
			err := db.CreateFoo(context.TODO(), want)
			require.NoError(t, err)

			got, err := db.ReadFoo(context.TODO(), "1")
			require.NoError(t, err)

			assert.Equal(t, want, got)
		})

		t.Run("with concurrent valid foo update the reading should pass", func(t *testing.T) {
			db := new(allsrv.InmemDB)
			require.NoError(t, db.CreateFoo(context.TODO(), allsrv.Foo{
				ID:   "1",
				Name: "one",
				Note: "note",
			}))

			newFoo := func(note string) allsrv.Foo {
				return allsrv.Foo{
					ID:   "1",
					Name: "one",
					Note: note,
				}
			}

			var wg sync.WaitGroup
			for _, f := range []allsrv.Foo{newFoo("a"), newFoo("b"), newFoo("c"), newFoo("d"), newFoo("e")} {
				wg.Add(1)
				go func(f allsrv.Foo) {
					defer wg.Done()
					require.NoError(t, db.UpdateFoo(context.TODO(), f))
				}(f)
			}

			got, err := db.ReadFoo(context.TODO(), "1")
			require.NoError(t, err)

			assert.Contains(t, []string{"note", "a", "b", "c", "d", "e"}, got.Note)
			wg.Wait()
		})

		t.Run("with id for non-existent foo should fail", func(t *testing.T) {
			db := new(allsrv.InmemDB)

			_, err := db.ReadFoo(context.TODO(), "1")

			// this is pretty gross, we're matching against a raw error/text value
			// any change in the error message means we have to update tests too
			want := errors.New("foo not found for id: 1")
			assert.Equal(t, want, err)
		})
	})

	t.Run("update foo", func(t *testing.T) {
		t.Run("with valid update for existing foo should pass", func(t *testing.T) {
			db := new(allsrv.InmemDB)

			want := allsrv.Foo{
				ID:   "1",
				Name: "name",
				Note: "note",
			}
			err := db.CreateFoo(context.TODO(), want)
			require.NoError(t, err)

			want.Note = "some other note"
			err = db.UpdateFoo(context.TODO(), want)
			require.NoError(t, err)

			got, err := db.ReadFoo(context.TODO(), "1")
			require.NoError(t, err)

			assert.Equal(t, want, got)
		})

		t.Run("with concurrent valid foo updates should pass", func(t *testing.T) {
			db := new(allsrv.InmemDB)
			require.NoError(t, db.CreateFoo(context.TODO(), allsrv.Foo{
				ID:   "1",
				Name: "one",
				Note: "note",
			}))

			newFoo := func(note string) allsrv.Foo {
				return allsrv.Foo{
					ID:   "1",
					Name: "one",
					Note: note,
				}
			}

			var wg sync.WaitGroup
			for _, f := range []allsrv.Foo{newFoo("a"), newFoo("b"), newFoo("c"), newFoo("d"), newFoo("e")} {
				wg.Add(1)
				go func(f allsrv.Foo) {
					defer wg.Done()
					require.NoError(t, db.UpdateFoo(context.TODO(), f))
				}(f)
			}

			got, err := db.ReadFoo(context.TODO(), "1")
			require.NoError(t, err)
			wg.Wait()

			assert.Contains(t, []string{"note", "a", "b", "c", "d", "e"}, got.Note)
		})

		t.Run("with update for non-existent foo should fail", func(t *testing.T) {
			db := new(allsrv.InmemDB)

			err := db.UpdateFoo(context.TODO(), allsrv.Foo{
				ID:   "1",
				Name: "name",
				Note: "note",
			})

			// this is pretty gross, we're matching against a raw error/text value
			// any change in the error message means we have to update tests too
			want := errors.New("foo not found for id: 1")
			assert.Equal(t, want, err)
		})
	})

	t.Run("delete foo", func(t *testing.T) {
		t.Run("with id for existing foo should pass", func(t *testing.T) {
			db := new(allsrv.InmemDB)

			err := db.CreateFoo(context.TODO(), allsrv.Foo{
				ID:   "1",
				Name: "name",
				Note: "note",
			})
			require.NoError(t, err)

			err = db.DelFoo(context.TODO(), "1")
			require.NoError(t, err)

			_, err = db.ReadFoo(context.TODO(), "1")

			// this is pretty gross, we're matching against a raw error/text value
			// any change in the error message means we have to update tests too
			want := errors.New("foo not found for id: 1")
			assert.Equal(t, want, err)
		})

		t.Run("with concurrent valid foo creates should pass", func(t *testing.T) {
			db := new(allsrv.InmemDB)

			newFoo := func(id string) allsrv.Foo {
				return allsrv.Foo{
					ID:   id,
					Name: "name-" + id,
					Note: "note-" + id,
				}
			}

			for _, f := range []allsrv.Foo{newFoo("1"), newFoo("2"), newFoo("3"), newFoo("4"), newFoo("5")} {
				require.NoError(t, db.CreateFoo(context.TODO(), f))
			}

			var wg sync.WaitGroup
			for _, id := range []string{"1", "2", "3", "4", "5"} {
				wg.Add(1)
				go func(id string) {
					defer wg.Done()
					require.NoError(t, db.DelFoo(context.TODO(), id))
				}(id)
			}
			wg.Wait()

			for _, id := range []string{"1", "2", "3", "4", "5"} {
				err := db.DelFoo(context.TODO(), id)
				wantErr := errors.New("foo not found for id: " + id)
				require.Error(t, wantErr, err)
			}
		})

		t.Run("with id for non-existent foo should fail", func(t *testing.T) {
			db := new(allsrv.InmemDB)

			err := db.DelFoo(context.TODO(), "1")

			// this is pretty gross, we're matching against a raw error/text value
			// any change in the error message means we have to update tests too
			want := errors.New("foo not found for id: 1")
			assert.Equal(t, want, err)
		})
	})
}
