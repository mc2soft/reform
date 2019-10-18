package reform

type sqlite3 struct{}

func (sqlite3) Placeholder(n int) string {
	return "?"
}

func (sqlite3) Placeholders(n int) []string {
	res := make([]string, n)
	for i := 0; i < n; i++ {
		res[i] = "?"
	}
	return res
}

func (sqlite3) GoType(dbType string) string {
	switch dbType {
	default:
		panic("reform: SQLite3.GoType: unhandled database type '" + dbType + "'")
	}
}

func (sqlite3) IsUniqueViolation(err error) bool {
	// FIXME: Implement
	return false
}

var SQLite3 sqlite3

// check interface
var _ Dialect = SQLite3
