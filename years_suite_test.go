package years_test

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2" //nolint: revive // ginkgo is fine
	. "github.com/onsi/gomega"    //nolint: revive // gomega is fine
)

const (
	TestDataPath = "internal/testdata"
)

type StaticClock struct {
	now time.Time
}

func (c *StaticClock) Now() time.Time { return c.now }

func TestYears(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Years Suite")
}
