package reform

import (
	"database/sql"
	"fmt"
	"strings"
)

func (db *DB) selectQuery(view View, tail string) string {
	var columns []string
	for _, col := range view.Columns() {
		columns = append(columns, view.Name()+"."+col)
	}

	return fmt.Sprintf("SELECT %s FROM %s %s",
		strings.Join(columns, ", "), view.Name(), tail)
}

// FIXME some function below do not work with NULLs
// "foo = NULL" is incorrect, should be "foo is NULL"

func (db *DB) FindOne(str Struct, column string, arg interface{}) error {
	tail := fmt.Sprintf("WHERE %s = %s LIMIT 1",
		column, db.Placeholder(1))
	return db.SelectOneTo(str, tail, arg)
}

func (db *DB) FindOneBy(view View, column string, arg interface{}) (Struct, error) {
	tail := fmt.Sprintf("WHERE %s = %s LIMIT 1",
		column, db.Placeholder(1))
	return db.SelectOneFrom(view, tail, arg)
}

func (db *DB) FindByPrimaryKey(table Table, id interface{}) (Struct, error) {
	tail := fmt.Sprintf("WHERE %s LIMIT 1",
		db.primaryKeyCondition(table, 1))
	return db.SelectOneFrom(table, tail, id)
}

func (db *DB) FindAllBy(view View, column string, args ...interface{}) ([]Struct, error) {
	tail := fmt.Sprintf("WHERE %s IN (%s)",
		column, strings.Join(db.Placeholders(len(args)), ", "))
	return db.SelectAllFrom(view, tail, args...)
}

func (db *DB) Reload(record Record) error {
	tail := fmt.Sprintf("WHERE %s LIMIT 1",
		db.primaryKeyCondition(record.Table(), 1))
	return db.SelectOneTo(record, tail, record.PrimaryKeyValues()...)
}

func (db *DB) SelectOneFrom(view View, tail string, args ...interface{}) (Struct, error) {
	str := view.NewStruct()
	q := db.selectQuery(view, tail)
	err := db.QueryRow(q, args...).Scan(str.Pointers()...)
	if err != nil {
		return nil, err
	}
	return str, nil
}

func (db *DB) SelectOneTo(str Struct, tail string, args ...interface{}) error {
	q := db.selectQuery(str.View(), tail)
	return db.QueryRow(q, args...).Scan(str.Pointers()...)
}

func (db *DB) SelectAllFrom(view View, tail string, args ...interface{}) ([]Struct, error) {
	rows, err := db.SelectRows(view, tail, args...)
	if rows != nil {
		defer rows.Close()
	}
	if err != nil {
		return nil, err
	}

	var structs []Struct
	for {
		str := view.NewStruct()
		err = db.Next(str, rows)
		if err == ErrNoRows {
			return structs, nil
		} else if err != nil {
			return structs, err
		}

		structs = append(structs, str)
	}
}

func (db *DB) SelectRows(view View, tail string, args ...interface{}) (*sql.Rows, error) {
	q := db.selectQuery(view, tail)
	return db.Query(q, args...)
}

func (db *DB) Next(str Struct, rows *sql.Rows) error {
	var err error
	next := rows.Next()
	if !next {
		err = rows.Err()
		if err == nil {
			err = ErrNoRows
		}
		return err
	}

	return rows.Scan(str.Pointers()...)
}

func (db *DB) Exec(q string, args ...interface{}) (sql.Result, error) {
	db.Log(q, args)
	return db.SqlBase.Exec(q, args...)
}

func (db *DB) Query(q string, args ...interface{}) (*sql.Rows, error) {
	db.Log(q, args)
	return db.SqlBase.Query(q, args...)
}

func (db *DB) QueryRow(q string, args ...interface{}) *sql.Row {
	db.Log(q, args)
	return db.SqlBase.QueryRow(q, args...)
}
