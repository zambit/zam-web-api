package db

import (
	"database/sql"
	"fmt"
	"github.com/lib/pq"
	"context"
)

// Db is wrapper around go-pg/pg DB object aimed to support some additional functionary
type Db struct {
	*sql.DB
}

// Factory just binds uri to New func call
func Factory(uri string) func() (*Db, error) {
	return func() (*Db, error) {
		return New(uri)
	}
}

// New creates new DB object connecting to remote using uri
func New(uri string) (*Db, error) {
	// Parse uri and connect
	connStr, err := pq.ParseURL(uri)
	if err != nil {
		return nil, err
	}
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Check connectivity
	_, err = db.Exec("SELECT 1 + 1;")
	if err != nil {
		return nil, err
	}

	// pg.DB may be safely copied by value
	return &Db{db}, nil
}

// ITx represents base tx interface for which both sql.DB and sql.Tx math
type ITx interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)

	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// RollbackError allow to recognize special errors which occurs while Recovering in case of closure error or panic recovery
type RollbackError interface {
	error

	// Panic returns captured panic or nil
	Panic() interface{}

	// ClosureErr returns error gained from closure or nil
	ClosureErr() error

	// Cause always returns error gained from Rollback call
	Cause() error
}

// Tx wraps function which will be executed safely relative to pg transaction aspect even in case of panic.
// It's like context manager.
func (db *Db) Tx(f func(tx ITx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// perform rollback in case of panic
	defer func() {
		p := recover()
		if p != nil {
			err := tx.Rollback()
			if err != nil {
				// panic special rollbackErr
				panic(rollbackErr{err: err, capturedPanic: p})
			}
			// in case when rollback error not occurs, just rethrow panic
			panic(p)
		}
	}()

	// perform closure
	err = f(tx)
	if err != nil {
		rErr := tx.Rollback()
		if rErr != nil {
			// need to track rollbacks errors with special rollbackErr
			return rollbackErr{err: rErr, closureErr: err}
		}
		return err
	}

	// commit if above successful
	return tx.Commit()
}

// rollbackErr implements RollbackError
type rollbackErr struct {
	err           error
	closureErr    error
	capturedPanic interface{}
}

// Error implements RollbackError
func (e rollbackErr) Error() string {
	pErr, ok := e.capturedPanic.(error)
	switch {
	case e.capturedPanic != nil && !ok:
		return fmt.Sprintf(
			"error occurred %v on rollback while recovering from non-err panic: %v", e.err, e.capturedPanic,
		)
	case e.capturedPanic == nil && ok:
		return fmt.Sprintf("error occurred %v on rollback while recovering: %v", e.err, pErr)
	case e.closureErr != nil:
		return fmt.Sprintf(
			"error occurred %v on rollback while recovering from closure error: %v", e.err, e.closureErr,
		)
	default:
		panic("should not occurs")
	}
}

// Panic implements RollbackError interface
func (e rollbackErr) Panic() interface{} {
	return e.capturedPanic
}

// ClosureErr implements RollbackError interface
func (e rollbackErr) ClosureErr() error {
	return e.closureErr
}

// Cause implements RollbackError interface
func (e rollbackErr) Cause() error {
	return e.err
}
