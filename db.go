package reform

type DB struct {
	SqlBase
	Dialect
	Logger              Logger
	transactionFinished bool
}

func (db *DB) Log(q string, args []interface{}) {
	if db.Logger != nil {
		db.Logger.Log(q, args)
	}
}
