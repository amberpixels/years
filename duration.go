package years

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// iso8601DurationRe matches the fixed-length subset of ISO-8601 durations used by
// media APIs (notably YouTube's contentDetails.duration): an optional day
// component plus a time component of hours/minutes/seconds, e.g. "PT15M30S",
// "P1DT2H3M4S", "PT45S". Weeks/months/years are intentionally unsupported — they
// are calendar quantities, not fixed durations, so they can't map to a Duration.
var iso8601DurationRe = regexp.MustCompile(
	`^P(?:(\d+)D)?(?:T(?:(\d+)H)?(?:(\d+)M)?(?:(\d+)S)?)?$`,
)

// ErrInvalidISODuration is returned by ParseISODuration for input that is not a
// supported ISO-8601 duration.
var ErrInvalidISODuration = errors.New("invalid ISO-8601 duration")

// ParseISODuration parses an ISO-8601 duration string into a time.Duration. It
// accepts the day+time subset (PnDTnHnMnS); calendar components (weeks, months,
// years) are rejected. An empty string is a zero duration with no error; any
// other input that carries no components (e.g. "P", "PT") is an error.
func ParseISODuration(s string) (time.Duration, error) {
	if s == "" {
		return 0, nil
	}
	m := iso8601DurationRe.FindStringSubmatch(s)
	if m == nil || (m[1] == "" && m[2] == "" && m[3] == "" && m[4] == "") {
		return 0, fmt.Errorf("%w: %q", ErrInvalidISODuration, s)
	}
	atoi := func(v string) int {
		n, _ := strconv.Atoi(v) // regex guarantees digits (or empty -> 0)
		return n
	}
	days, hours, mins, secs := atoi(m[1]), atoi(m[2]), atoi(m[3]), atoi(m[4])
	totalSecs := ((days*24+hours)*60+mins)*60 + secs
	return time.Duration(totalSecs) * time.Second, nil
}

// FormatDurationClock renders d in colon-separated media style: "M:SS" when under
// an hour (e.g. "15:30") and "H:MM:SS" otherwise (e.g. "1:02:03"). Negative
// durations are treated as zero. Sub-second remainders are truncated.
func FormatDurationClock(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	total := int(d / time.Second)
	h, m, s := total/3600, (total%3600)/60, total%60
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%d:%02d", m, s)
}

// HumanizeDuration renders d as a compact, human-friendly approximation: the
// most-significant non-zero unit plus the next unit when it too is non-zero,
// e.g. 2h5m3s -> "2h 5m", 90m -> "1h 30m", 45s -> "45s", 3d4h -> "3d 4h".
// Sub-second durations render as "0s". The sign is dropped (magnitude only); use
// Humanize/HumanizeFrom for signed, relative "ago"/"in" phrasing.
func HumanizeDuration(d time.Duration) string {
	if d < 0 {
		d = -d
	}
	if d < time.Second {
		return "0s"
	}
	parts := []struct {
		v int
		u string
	}{
		{int(d / (24 * time.Hour)), "d"},
		{int(d % (24 * time.Hour) / time.Hour), "h"},
		{int(d % time.Hour / time.Minute), "m"},
		{int(d % time.Minute / time.Second), "s"},
	}
	i := 0
	for i < len(parts) && parts[i].v == 0 {
		i++
	}
	out := []string{strconv.Itoa(parts[i].v) + parts[i].u}
	if i+1 < len(parts) && parts[i+1].v != 0 {
		out = append(out, strconv.Itoa(parts[i+1].v)+parts[i+1].u)
	}
	return strings.Join(out, " ")
}
