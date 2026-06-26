package years_test

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/amberpixels/years"
	"github.com/expectto/be"
	"github.com/expectto/be/be_time"
)

func TestJustParseRaw_Nil(t *testing.T) {
	parsedTime, err := years.JustParseRaw(nil)
	be.Require(t, err).To(be.Succeed())
	be.Expect(t, parsedTime).To(be.Eq(time.Time{}))
}

func TestJustParseRaw_Time(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	parsedTime, err := years.JustParseRaw(now)
	be.Require(t, err).To(be.Succeed())
	be.Expect(t, parsedTime).To(be.Eq(now))
}

func TestJustParseRaw_NumericString(t *testing.T) {
	var timestamp int64 = 1709682885
	value := strconv.Itoa(int(timestamp))
	parsedTime, err := years.JustParseRaw(value)
	be.Require(t, err).To(be.Succeed())
	be.Expect(t, parsedTime).To(be_time.Unix(timestamp))
}

func TestJustParseRaw_Integer(t *testing.T) {
	v := 12345
	parsedTime, err := years.JustParseRaw(v)
	be.Require(t, err).To(be.Succeed())
	be.Expect(t, parsedTime).To(be_time.Unix(int64(v)))
}

func TestJustParseRaw_UnsupportedType(t *testing.T) {
	unsupported := 3.14
	_, err := years.JustParseRaw(unsupported)
	be.Require(t, err).To(be.HaveOccurred())
	be.Expect(t, err.Error()).To(be.Eq(
		fmt.Sprintf("unsupported type %T for JustParseRaw", unsupported),
	))
}
