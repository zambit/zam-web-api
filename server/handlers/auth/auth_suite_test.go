package auth

import (
	"testing"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestAuthHandlers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Auth Handlers Suite")
}

