package reform

func (db *DB) Begin() (*DB, error) {
	db.Log("BEGIN", nil)
	tx, err := db.SqlBase.(SqlDB).Begin()
	if err != nil {
		return nil, err
	}

	return &DB{
		SqlBase: tx,
		Dialect: db.Dialect,
		Logger:  db.Logger,
	}, nil
}

func (db *DB) Commit() error {
	if db.transactionFinished {
		return nil
	}

	db.Log("COMMIT", nil)
	db.transactionFinished = true
	return db.SqlBase.(SqlTx).Commit()
}

func (db *DB) Rollback() error {
	if db.transactionFinished {
		return nil
	}

	db.Log("ROLLBACK", nil)
	db.transactionFinished = true
	return db.SqlBase.(SqlTx).Rollback()
}
