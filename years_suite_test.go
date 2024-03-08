package years_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	TestDataPath = "internal/testdata"
)

func TestYears(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Years Suite")
}
