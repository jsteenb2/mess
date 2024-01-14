package allsrv_test

import (
	"context"
	"errors"
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
