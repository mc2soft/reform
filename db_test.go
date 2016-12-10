package reform_test

import (
	"errors"

	"github.com/AlekSi/pointer"
	"github.com/enodata/faker"

	"github.com/lib/pq"
	"gopkg.in/reform.v1"
	"gopkg.in/reform.v1/dialects/mysql"
	"gopkg.in/reform.v1/dialects/postgresql"
	"gopkg.in/reform.v1/dialects/sqlite3"
	. "gopkg.in/reform.v1/internal/test/models"
)

func (s *ReformSuite) TestBeginCommit() {
	setIdentityInsert(s.T(), s.q, "people", true)

	s.Require().NoError(s.q.Rollback())
	s.q = nil

	person := &Person{ID: 42, Email: pointer.ToString(faker.Internet().Email())}

	tx, err := DB.Begin()
	s.Require().NoError(err)
	s.NoError(tx.Insert(person))
	s.NoError(tx.Commit())
	s.Equal(tx.Commit(), reform.ErrTxDone)
	s.Equal(tx.Rollback(), reform.ErrTxDone)
	s.NoError(DB.Reload(person))
	s.NoError(DB.Delete(person))
}

func (s *ReformSuite) TestBeginRollback() {
	setIdentityInsert(s.T(), s.q, "people", true)

	s.Require().NoError(s.q.Rollback())
	s.q = nil

	person := &Person{ID: 42, Email: pointer.ToString(faker.Internet().Email())}

	tx, err := DB.Begin()
	s.Require().NoError(err)
	s.NoError(tx.Insert(person))
	s.NoError(tx.Rollback())
	s.Equal(tx.Commit(), reform.ErrTxDone)
	s.Equal(tx.Rollback(), reform.ErrTxDone)
	s.Equal(DB.Reload(person), reform.ErrNoRows)
}

// This behavior is checked for documentation purposes only. reform does not rely on it.
func (s *ReformSuite) TestErrorInTransaction() {
	if s.q.Dialect == postgresql.Dialect {
		s.T().Skip(s.q.Dialect.String() + " works differently, see TestAbortedTransaction")
	}

	setIdentityInsert(s.T(), s.q, "people", true)

	s.Require().NoError(s.q.Rollback())
	s.q = nil

	person1 := &Person{ID: 42, Email: pointer.ToString(faker.Internet().Email())}
	person2 := &Person{ID: 43, Email: pointer.ToString(faker.Internet().Email())}

	// commit works
	tx, err := DB.Begin()
	s.Require().NoError(err)
	s.NoError(tx.Insert(person1))
	s.Error(tx.Insert(person1)) // duplicate PK
	s.NoError(tx.Insert(person2))
	s.NoError(tx.Commit())
	s.Equal(tx.Commit(), reform.ErrTxDone)
	s.Equal(tx.Rollback(), reform.ErrTxDone)
	s.NoError(DB.Reload(person1))
	s.NoError(DB.Reload(person2))
	s.NoError(DB.Delete(person1))
	s.NoError(DB.Delete(person2))

	// rollback works
	tx, err = DB.Begin()
	s.Require().NoError(err)
	s.NoError(tx.Insert(person1))
	s.Error(tx.Insert(person1)) // duplicate PK
	s.NoError(tx.Insert(person2))
	s.NoError(tx.Rollback())
	s.Equal(tx.Commit(), reform.ErrTxDone)
	s.Equal(tx.Rollback(), reform.ErrTxDone)
	s.Equal(DB.Reload(person1), reform.ErrNoRows)
	s.Equal(DB.Reload(person2), reform.ErrNoRows)
}

// This behavior is checked for documentation purposes only. reform does not rely on it.
func (s *ReformSuite) TestAbortedTransaction() {
	if s.q.Dialect == mysql.Dialect || s.q.Dialect == sqlite3.Dialect {
		s.T().Skip(s.q.Dialect.String() + " works differently, see TestErrorInTransaction")
	}

	setIdentityInsert(s.T(), s.q, "people", true)

	s.Require().NoError(s.q.Rollback())
	s.q = nil

	person := &Person{ID: 42, Email: pointer.ToString(faker.Internet().Email())}

	// commit fails
	tx, err := DB.Begin()
	s.Require().NoError(err)
	s.NoError(tx.Insert(person))
	s.EqualError(tx.Insert(person), `pq: duplicate key value violates unique constraint "people_pkey"`)
	s.Equal(tx.Commit(), pq.ErrInFailedTransaction)
	s.Equal(tx.Commit(), reform.ErrTxDone)
	s.Equal(tx.Rollback(), reform.ErrTxDone)
	s.Equal(DB.Reload(person), reform.ErrNoRows)

	// rollback works
	tx, err = DB.Begin()
	s.Require().NoError(err)
	s.NoError(tx.Insert(person))
	s.EqualError(tx.Insert(person), `pq: duplicate key value violates unique constraint "people_pkey"`)
	s.NoError(tx.Rollback())
	s.Equal(tx.Commit(), reform.ErrTxDone)
	s.Equal(tx.Rollback(), reform.ErrTxDone)
	s.Equal(DB.Reload(person), reform.ErrNoRows)
}

func (s *ReformSuite) TestInTransaction() {
	setIdentityInsert(s.T(), s.q, "people", true)

	s.Require().NoError(s.q.Rollback())
	s.q = nil

	person := &Person{ID: 42, Email: pointer.ToString(faker.Internet().Email())}

	err := DB.InTransaction(func(tx *reform.TX) error {
		err := tx.Insert(person)
		s.NoError(err)
		return errors.New("epic error")
	})
	s.EqualError(err, "epic error")
	s.Equal(DB.Reload(person), reform.ErrNoRows)

	s.Panics(func() {
		err = DB.InTransaction(func(tx *reform.TX) error {
			err := tx.Insert(person)
			s.NoError(err)
			panic("epic panic!")
		})
	})
	s.Equal(DB.Reload(person), reform.ErrNoRows)

	err = DB.InTransaction(func(tx *reform.TX) error {
		err := tx.Insert(person)
		s.NoError(err)
		return nil
	})
	s.NoError(err)
	s.NoError(DB.Reload(person))

	s.NoError(DB.Delete(person))
}