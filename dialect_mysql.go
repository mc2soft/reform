package reform

type mySQL struct{}

func (mySQL) Placeholder(n int) string {
	return "?"
}

func (mySQL) Placeholders(n int) []string {
	res := make([]string, n)
	for i := 0; i < n; i++ {
		res[i] = "?"
	}
	return res
}

func (mySQL) GoType(dbType string) string {
	switch dbType {
	default:
		panic("reform: MySQL.GoType: unhandled database type '" + dbType + "'")
	}
}

func (mySQL) IsUniqueViolation(err error) bool {
	// FIXME: Implement
	return false
}

var MySQL mySQL

// check interface
var _ Dialect = MySQL
