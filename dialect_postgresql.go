package reform

import (
	"github.com/lib/pq"
	"strconv"
)

type postgreSQL struct{}

func (postgreSQL) Placeholder(n int) string {
	return "$" + strconv.Itoa(n)
}

func (postgreSQL) Placeholders(n int) []string {
	res := make([]string, n)
	for i := 0; i < n; i++ {
		res[i] = "$" + strconv.Itoa(i+1)
	}
	return res
}

func (postgreSQL) GoType(dbType string) string {
	switch dbType {
	case "smallint":
		return "int16"
	case "integer":
		return "int32"
	case "bigint":
		return "int64"
	case "real":
		return "float32"
	case "double precision", "numeric":
		return "float64"
	case "character varying", "text":
		return "string"
	case "bytea":
		return "[]byte"
	case "timestamp with time zone", "timestamp without time zone", "date":
		return "time.Time"
	case "boolean":
		return "bool"
	case "ARRAY": // TODO fix or remove
		return "[]interface{}"
	default:
		panic("reform: PostgreSQL.GoType: unhandled database type '" + dbType + "'")
	}
}

func (postgreSQL) IsUniqueViolation(err error) bool {
	e, ok := err.(*pq.Error)
	return ok && e.Code == "23505"
}

var PostgreSQL postgreSQL

// check interface
var _ Dialect = PostgreSQL
