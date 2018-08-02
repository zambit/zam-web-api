package utils

import (
	"go.uber.org/dig"
)

// MustProvide wraps container provide method and panics on error instead of returning
func MustProvide(c *dig.Container, provider interface{}, opts ...dig.ProvideOption) {
	err := c.Provide(provider, opts...)
	if err != nil {
		panic(err)
	}
}

// MustInvoke wraps container invoke method and panics on error instead of returning
func MustInvoke(c *dig.Container, invoker interface{}, opts ...dig.InvokeOption) {
	err := c.Invoke(invoker, opts...)
	if err != nil {
		panic(err)
	}
}
