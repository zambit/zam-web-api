package notifications

import (
	"github.com/google/uuid"
	"math/rand"
	"strings"
)

// IGenerator holds different auxiliary functions for generating purposes
type IGenerator interface {
	// RandomCode generates random confirmation-like code
	RandomCode() string

	// RandomToken generates random token-like char sequence
	RandomToken() string
}

// NewWithCodeAlphabet returns new generator which uses UUID4 for token and some alphabet for token generation
func NewWithCodeAlphabet(codeLen int, codeAlphabet string) IGenerator {
	return generator{
		codeLen:      codeLen,
		codeAlphabet: codeAlphabet,
	}
}

// generator default IGenerator implementation
type generator struct {
	codeLen      int
	codeAlphabet string
}

// RandomCode implements IGenerator interface
func (g generator) RandomCode() string {
	resBytes := make([]byte, g.codeLen)
	// entropy container
	randomness := make([]byte, g.codeLen)
	_, err := rand.Read(randomness)
	if err != nil {
		panic(err)
	}

	// fill output
	l := len(g.codeAlphabet)
	for pos := range resBytes {
		random := uint8(randomness[pos])
		randomPos := random % uint8(l)
		resBytes[pos] = g.codeAlphabet[randomPos]
	}
	return string(resBytes)
}

// RandomToken implements IGenerator interface
func (g generator) RandomToken() string {
	return strings.Replace(uuid.New().String(), "-", "", 4)
}
