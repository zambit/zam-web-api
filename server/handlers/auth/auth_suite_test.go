package auth

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
)

func TestAuthHandlers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Auth Handlers Suite")
}
