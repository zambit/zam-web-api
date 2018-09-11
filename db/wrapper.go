package db

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
	"github.com/lib/pq"
	"reflect"
	"unsafe"
)

// Db is wrapper around go-pg/pg DB object aimed to support some additional functionary
type Db struct {
	*sqlx.DB
}

// NamedQueryRow implements hijack on sqlx.Rows turning them into sqlx.Row
func (db *Db) NamedQueryRow(query string, arg interface{}) *sqlx.Row {
	return namedQueryRow(db, query, arg)
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
	db, err := sqlx.Open("postgres", connStr)
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

// IGenericQueryMaker this interface may represent both sqlx.DB and sqlx.Tx
type IGenericQueryMaker interface {
	sqlx.Queryer
	sqlx.Execer
	sqlx.QueryerContext
	sqlx.ExecerContext

	NamedQuery(query string, arg interface{}) (*sqlx.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// ITx represents base tx interface for which both sql.DB and sql.Tx math
type ITx interface {
	IGenericQueryMaker
	NamedQueryRow(query string, arg interface{}) *sqlx.Row
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
	tx, err := db.Beginx()
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
	err = f(addNamedQueryRowMethodOnITx{tx})
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

// addNamedQueryRowMethodOnITx used to demolish luck of query row method on Tx
type addNamedQueryRowMethodOnITx struct {
	IGenericQueryMaker
}

// utils
// NamedQueryRow implements hijack on sqlx.Rows turning them into sqlx.Row
func (a addNamedQueryRowMethodOnITx) NamedQueryRow(query string, arg interface{}) *sqlx.Row {
	return namedQueryRow(a.IGenericQueryMaker, query, arg)
}

func getMapper(a IGenericQueryMaker) *reflectx.Mapper {
	if db, ok := a.(*sqlx.DB); ok {
		return db.Mapper
	}
	if tx, ok := a.(*sqlx.Tx); ok {
		return tx.Mapper
	}
	panic(fmt.Errorf(
		"given generic query maker interface of type %T does have underlying type of either *sqlx.DB not *sqlx.Tx", a,
	))
}

type rowCreatorType func(err error, rows *sql.Rows, mapper *reflectx.Mapper) *sqlx.Row

var createRowFunc rowCreatorType

func namedQueryRow(qm IGenericQueryMaker, query string, arg interface{}) *sqlx.Row {
	rows, err := qm.NamedQuery(query, arg)
	return createRowFunc(err, rows.Rows, getMapper(qm))
}

func init() {
	createRowFunc = createRowCreator()
}

func createRowCreator() rowCreatorType {
	rowType := reflect.TypeOf(sqlx.Row{})

	errValField, ok := rowType.FieldByName("err")
	if !ok {
		panic("failed to find err field on *sqlx.Row")
	}

	rowsValField, ok := rowType.FieldByName("rows")
	if !ok {
		panic("failed to find rows field on *sqlx.Row")
	}

	return func(err error, rows *sql.Rows, mapper *reflectx.Mapper) *sqlx.Row {
		row := &sqlx.Row{}
		rowElem := reflect.ValueOf(row).Elem()

		if err != nil {
			errField := reflect.NewAt(errValField.Type, unsafe.Pointer(rowElem.UnsafeAddr()+errValField.Offset)).Elem()
			errField.Set(reflect.ValueOf(err))
		}

		if rows != nil {
			rowsField := reflect.NewAt(rowsValField.Type, unsafe.Pointer(rowElem.UnsafeAddr()+rowsValField.Offset)).Elem()
			rowsField.Set(reflect.ValueOf(rows))
		}

		row.Mapper = mapper
		return row
	}
}
