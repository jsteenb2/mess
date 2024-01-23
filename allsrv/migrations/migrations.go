// package migrations provides for a one-stop shop for accessing migrations.
// this removes the need to worry about relative pathing from the source file
// intending to migrate a DB. All that is required is a go import and a
// reference to the migrations of interest.

package migrations

import (
	"embed"
)

// SQLite represents the sqlite migration files.
//
//go:embed sqlite
var SQLite embed.FS
