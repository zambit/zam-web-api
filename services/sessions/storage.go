package sessions

import (
	"github.com/pkg/errors"
	"time"
)

// Storage errors
var (
	// ErrUnexpectedToken returned when given token violates the backend-storage token format
	ErrUnexpectedToken = errors.New("given token violates the storage token format")

	// ErrNotFound returned when no such token has been found
	ErrNotFound = errors.New("no such token registered or already deleted")

	// ErrExpired returned when requested token already expires
	ErrExpired = errors.New("token expired")
)

// Token represents user session token
type Token []byte

// IStorage collects, persist and manages user auth sessions via tokens and associated-optional data.
type IStorage interface {
	// New creates new session
	New(data map[string]interface{}, expireAfter time.Duration) (Token, error)

	// Get returns data associated with this token
	Get(token Token) (data map[string]interface{}, err error)

	// RefreshToken
	RefreshToken(oldToken Token, expireAfter time.Duration) (Token, error)

	// Delete makes token invalid so sequential Get call will returns ErrNotFound
	Delete(toke Token) error
}
