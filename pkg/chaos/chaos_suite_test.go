package chaos_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestChaos(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Chaos Suite")
}
