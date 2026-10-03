//go:build tools

// Package tools pins dependencies the pipeline declares now and imports in
// later phases (SQLite storage, migrations, test assertions). Keeping them
// referenced here means `go mod tidy` leaves them in go.mod before the code
// that uses them exists.
//
// The `tools` build tag keeps this file out of normal builds. Refresh it as
// each dependency gains a real import site.
package tools

import (
	_ "github.com/golang-migrate/migrate/v4"
	_ "github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
)
