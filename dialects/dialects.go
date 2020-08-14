// Package dialects implements reform.Dialect selector.
package dialects

import (
	"strings"

	"github.com/mc2soft/reform"
	"github.com/mc2soft/reform/dialects/mssql" //nolint:staticcheck
	"github.com/mc2soft/reform/dialects/mysql"
	"github.com/mc2soft/reform/dialects/postgresql"
	"github.com/mc2soft/reform/dialects/sqlite3"
	"github.com/mc2soft/reform/dialects/sqlserver"
)

// ForDriver returns reform Dialect for given driver string, or nil.
func ForDriver(driver string) reform.Dialect {
	// for sqlite3_with_sleep
	if strings.HasPrefix(driver, "sqlite3") {
		return sqlite3.Dialect
	}

	switch driver {
	case "postgres", "pgx":
		return postgresql.Dialect
	case "mysql":
		return mysql.Dialect
	case "mssql":
		return mssql.Dialect //nolint:staticcheck
	case "sqlserver":
		return sqlserver.Dialect
	default:
		return nil
	}
}
