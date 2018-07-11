package types

import (
	"golang.org/x/crypto/bcrypt"
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
