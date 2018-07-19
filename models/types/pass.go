package types

import (
	"golang.org/x/crypto/bcrypt"
	"fmt"
	"github.com/pkg/errors"
	"reflect"
	"database/sql/driver"
)

const defaultCost = 8

// Password simplifies bcrypt password manipulations
type Password []byte

// NewPass
func NewPass(originalPassword string) (Password, error) {
	if len(originalPassword) == 0 {
		return []byte{}, nil
	}

	return bcrypt.GenerateFromPassword([]byte(originalPassword), defaultCost)
}

// Compare compares to raw password (not-encoded)
func (pswd Password) Compare(rawPassword string) (bool, error) {
	if len(pswd) == 0 {
		return false, nil
	}

	err := bcrypt.CompareHashAndPassword(pswd, []byte(rawPassword))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

// Value implements sql.Valuer interface, represent zero-len password as null in database
func (pswd Password) Value() (driver.Value, error) {
	if len(pswd) == 0 {
		return nil, nil
	}

	return []byte(pswd), nil
}

// Scan implements sql.Scanner interface to support assignment from DB
func (pswd *Password) Scan(src interface{}) error {
	if src == nil {
		*pswd = []byte{}
		return nil
	} else if s, ok := src.(string); ok {
		*pswd = []byte(s)
		return nil
	}
	return errors.WithStack(fmt.Errorf(
		"unexpected type received while scanning passowrd field: expect string, receive %s",
		reflect.TypeOf(src).Name(),
	))
}