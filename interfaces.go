package reform

import (
	"database/sql"
	"errors"
)

var (
	ErrNoPrimaryKey    = errors.New("sql: no primary key")
	ErrNoRows          = sql.ErrNoRows
	ErrUniqueViolation = errors.New("sql: unique violation")
)

type View interface {
	Name() string
	Columns() []string
	OmitEmpty() []bool
	ZeroValues() []interface{}
	NewStruct() Struct
}

type Table interface {
	View
	NewRecord() Record
	PrimaryKeyColumns() []string
	UpdatedColumn() string
}

type Struct interface {
	Values() []interface{}
	Pointers() []interface{}
	View() View
}

type Record interface {
	Struct
	Table() Table
	PrimaryKeyPointers() []interface{}
	PrimaryKeyValues() []interface{}
	PrimaryKeyEmpty() bool
	SetPrimaryKey(id ...interface{})
	BeforeInsert()
	BeforeUpdate()
}

type Dialect interface {
	Placeholder(n int) string
	Placeholders(n int) []string
	GoType(dbType string) string
	IsUniqueViolation(err error) bool
}

type Logger interface {
	Log(query string, args []interface{})
}

type SqlBase interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

type SqlDB interface {
	SqlBase
	Begin() (*sql.Tx, error)
}

type SqlTx interface {
	SqlBase
	Commit() error
	Rollback() error
}
