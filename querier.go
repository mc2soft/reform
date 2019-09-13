package reform

import (
	"database/sql"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// Querier performs queries and commands.
type Querier struct {
	dbtx DBTX
	tag  string
	Dialect
	Logger        Logger
	inTransaction bool
	slaves        []DBTX
	onCommitCalls []func() error
}

func newQuerier(dbtx DBTX, dialect Dialect, logger Logger) *Querier {
	return &Querier{
		dbtx:    dbtx,
		Dialect: dialect,
		Logger:  logger,
	}
}

func (q *Querier) logBefore(query string, args []interface{}) {
	if q.Logger != nil {
		q.Logger.Before(query, args)
	}
}

func (q *Querier) logAfter(query string, args []interface{}, d time.Duration, err error) {
	if q.Logger != nil {
		q.Logger.After(query, args, d, err)
	}
}

func (q *Querier) startQuery(command string) string {
	if q.tag == "" {
		return command
	}
	return command + " /* " + q.tag + " */"
}

// WithTag returns a copy of Querier with set tag. Returned Querier is tied to the same DB or TX.
// See Tagging section in documentation for details.
func (q *Querier) WithTag(format string, args ...interface{}) *Querier {
	newQ := newQuerier(q.dbtx, q.Dialect, q.Logger)
	if len(args) == 0 {
		newQ.tag = format
	} else {
		newQ.tag = fmt.Sprintf(format, args...)
	}
	return newQ
}

// QualifiedView returns quoted qualified view name.
func (q *Querier) QualifiedView(view View) string {
	v := q.QuoteIdentifier(view.Name())
	if view.Schema() != "" {
		v = q.QuoteIdentifier(view.Schema()) + "." + v
	}
	return v
}

// QualifiedColumns returns a slice of quoted qualified column names for given view.
func (q *Querier) QualifiedColumns(view View) []string {
	v := q.QualifiedView(view)
	res := view.Columns()
	for i := 0; i < len(res); i++ {
		res[i] = v + "." + q.QuoteIdentifier(res[i])
	}
	return res
}

func (q *Querier) selectDBTX(query string) DBTX {
	if q.inTransaction || len(q.slaves) == 0 || !strings.HasPrefix(strings.TrimSpace(query), "SELECT") {
		return q.dbtx
	}

	ind := rand.Intn(len(q.slaves))
	return q.slaves[ind]
}

// Exec executes a query without returning any rows.
// The args are for any placeholder parameters in the query.
func (q *Querier) Exec(query string, args ...interface{}) (sql.Result, error) {
	q.logBefore(query, args)
	start := time.Now()
	dbtx := q.selectDBTX(query)
	res, err := dbtx.Exec(query, args...)
	q.logAfter(query, args, time.Since(start), err)
	return res, err
}

// Query executes a query that returns rows, typically a SELECT.
// The args are for any placeholder parameters in the query.
func (q *Querier) Query(query string, args ...interface{}) (*sql.Rows, error) {
	q.logBefore(query, args)
	start := time.Now()
	dbtx := q.selectDBTX(query)
	rows, err := dbtx.Query(query, args...)
	q.logAfter(query, args, time.Since(start), err)
	return rows, err
}

// QueryRow executes a query that is expected to return at most one row.
// QueryRow always returns a non-nil value. Errors are deferred until Row's Scan method is called.
func (q *Querier) QueryRow(query string, args ...interface{}) *sql.Row {
	q.logBefore(query, args)
	start := time.Now()
	dbtx := q.selectDBTX(query)
	row := dbtx.QueryRow(query, args...)
	q.logAfter(query, args, time.Since(start), nil)
	return row
}

func (q *Querier) IsInTransaction() bool {
	return q.inTransaction
}

func (q *Querier) AddOnCommitCall(f func() error) {
	if !q.inTransaction {
		panic("OnCommit callback added outside transaction")
	}

	q.onCommitCalls = append(q.onCommitCalls, f)
}

// check interface
var _ DBTX = (*Querier)(nil)
