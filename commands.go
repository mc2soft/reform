package reform

import (
	"fmt"
	"strings"
)

func (db *DB) Insert(r Record) error {
	table := r.Table()

	r.BeforeInsert()

	var columns, placeholders []string
	var values []interface{}
	for i, v := range r.Values() {
		if table.OmitEmpty()[i] && v == table.ZeroValues()[i] {
			continue
		}
		columns = append(columns, table.Columns()[i])
		placeholders = append(placeholders, db.Placeholder(len(columns)))
		values = append(values, v)
	}

	q := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		table.Name(),
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	switch db.Dialect {
	case PostgreSQL:
		q += fmt.Sprintf(" RETURNING %s", strings.Join(table.PrimaryKeyColumns(), ", "))
		err := db.QueryRow(q, values...).Scan(r.PrimaryKeyPointers()...)
		if db.IsUniqueViolation(err) {
			return ErrUniqueViolation
		}
		return err
	default:
		res, err := db.Exec(q, values...)
		if err != nil {
			return err
		}
		id, err := res.LastInsertId()
		if err != nil {
			return err
		}
		r.SetPrimaryKey(id)
		return nil
	}
}

func (db *DB) Update(r Record) error {
	return db.UpdateOnly(r, r.Table().Columns()[len(r.PrimaryKeyValues()):])
}

func (db *DB) primaryKeyCondition(table Table, firstPlaceholder int) string {
	columns := table.PrimaryKeyColumns()
	conditions := make([]string, len(columns))
	for i, column := range columns {
		conditions[i] = fmt.Sprintf("%s = %s", column, db.Placeholder(firstPlaceholder+i))
	}

	return strings.Join(conditions, " AND ")
}

func (db *DB) UpdateOnly(r Record, onlyColumns []string) error {
	r.BeforeUpdate()

	table := r.Table()

	if r.PrimaryKeyEmpty() {
		return ErrNoPrimaryKey
	}

	// copy slice
	onlyColumns = append([]string(nil), onlyColumns...)

	var pairs []string
	var values []interface{}
	for i, v := range r.Values() {
		// skip primary key
		if i < len(r.PrimaryKeyValues()) {
			continue
		}

		column := table.Columns()[i]

		found := false
		for i, col := range onlyColumns {
			if column == col {
				found = true
				onlyColumns = append(onlyColumns[:i], onlyColumns[i+1:]...)
				break
			}
		}
		if !found {
			continue
		}

		values = append(values, v)
		pairs = append(pairs, column+" = "+db.Placeholder(len(values)))
	}

	if len(onlyColumns) != 0 {
		panic(fmt.Sprintf("some columns left in update only: %v", onlyColumns))
	}

	q := fmt.Sprintf("UPDATE %s SET %s WHERE %s",
		table.Name(),
		strings.Join(pairs, ", "),
		db.primaryKeyCondition(table, len(values)+1),
	)
	values = append(values, r.PrimaryKeyValues()...)

	res, err := db.Exec(q, values...)
	if err != nil {
		if db.IsUniqueViolation(err) {
			return ErrUniqueViolation
		}
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return ErrNoRows
	}
	return nil
}

// UPDATEs if id is present, INSERTs otherwise
func (db *DB) Upsert(r Record) error {
	if r.PrimaryKeyEmpty() {
		return db.Insert(r)
	} else {
		return db.Update(r)
	}
}

// Saves the record no matter what
func (db *DB) Save(r Record) error {
	err := db.Upsert(r)
	if err == ErrNoRows {
		err = db.Insert(r)
	}
	return err
}

// Delete destroys the record (by primary key)
func (db *DB) Delete(r Record) error {
	table := r.Table()

	q := fmt.Sprintf("DELETE FROM %s WHERE %s",
		table.Name(),
		db.primaryKeyCondition(table, 1),
	)

	res, err := db.Exec(q, r.PrimaryKeyValues()...)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return ErrNoRows
	}
	return nil
}

// Apply does insert or update based on primary key, tracking changes
func (db *DB) Apply(table Table, changes func(Record), id interface{}) (Record, error) {
	var (
		s   Struct
		r   Record
		old Struct
		err error
	)
	s, err = db.FindByPrimaryKey(table, id)
	if err != nil {
		if err != ErrNoRows {
			return nil, err
		}

		r = table.NewRecord()
		r.SetPrimaryKey(id)
	} else {
		r = s.(Record)
		old = Copy(s)
	}

	changes(r)

	if old != nil {
		changed := ChangedFields(old, r)
		if len(changed) > 0 {
			if table.UpdatedColumn() != "" {
				changed = append(changed, table.UpdatedColumn())
			}
			err = db.UpdateOnly(r, changed)
		}
	} else {
		err = db.Insert(r)
	}

	return r, err
}
