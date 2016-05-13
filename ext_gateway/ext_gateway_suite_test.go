package ext_gateway_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestExtGateway(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ExtGateway Suite")
}
