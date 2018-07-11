package types

import (
	"golang.org/x/crypto/bcrypt"
	"fmt"
	"github.com/pkg/errors"
	"reflect"
)

const defaultCost = 8

// Password simplifies bcrypt password manipulations
type Password []byte

// NewPass
func NewPass(originalPassword string) (Password, error) {
	return bcrypt.GenerateFromPassword([]byte(originalPassword), defaultCost)
}

// Compare compares to raw password (not-encoded)
func (pswd Password) Compare(rawPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(pswd, []byte(rawPassword))
	if err != nil {
		return false, err
	}
	return true, nil
}

// // Value implements sql.Valuer interface
// func (pswd Password) Value() (driver.Value, error) {
// 	panic("implement me")
// }

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