package types

import (
	"github.com/pkg/errors"
	"github.com/ttacon/libphonenumber"
)

// defaultRegion static so far
const defaultRegion = "RU"

// Phone is normalized phone representation
type Phone string

// ErrInvalidPhoneNumber indicates that
var ErrInvalidPhoneNumber = errors.New("invalid phone number")

// New creates normalized phone form and validates it
func NewPhone(rawPhone string) (Phone, error) {
	num, err := libphonenumber.Parse(rawPhone, defaultRegion)
	if err != nil {
		return "", ErrInvalidPhoneNumber
	}

	if !libphonenumber.IsValidNumber(num) {
		return "", ErrInvalidPhoneNumber
	}

	// return formatted phone repr
	return Phone(libphonenumber.Format(num, libphonenumber.E164)), nil
}
