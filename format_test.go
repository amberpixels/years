package years_test

import (
	"testing"
	"time"

	"github.com/amberpixels/years"
	"github.com/expectto/be"
)

func TestFormat(t *testing.T) {
	t0 := time.Date(2025, time.April, 30, 13, 45, 59, 0, time.UTC)

	be.Expect(t, years.Format(t0, years.LayoutDate)).To(be.Eq("2025-04-30"))
	be.Expect(t, years.Format(t0, years.LayoutDateTime)).To(be.Eq("2025-04-30 13:45:59"))
	be.Expect(t, years.Format(t0, years.LayoutDateTimeShort)).To(be.Eq("2025-04-30 13:45"))
	be.Expect(t, years.Format(t0, years.LayoutHuman)).To(be.Eq("Apr 30, 2025 13:45"))
	be.Expect(t, years.Format(t0, years.LayoutHumanDate)).To(be.Eq("Apr 30, 2025"))
}

func TestFormat_ZeroTimeIsEmpty(t *testing.T) {
	be.Expect(t, years.Format(time.Time{}, years.LayoutDate)).To(be.Eq(""))
}

func TestFormatPtr(t *testing.T) {
	t0 := time.Date(2025, time.April, 30, 0, 0, 0, 0, time.UTC)

	be.Expect(t, years.FormatPtr(&t0, years.LayoutDate)).To(be.Eq("2025-04-30"))
	be.Expect(t, years.FormatPtr(nil, years.LayoutDate)).To(be.Eq(""))

	var zero time.Time
	be.Expect(t, years.FormatPtr(&zero, years.LayoutDate)).To(be.Eq(""))
}
