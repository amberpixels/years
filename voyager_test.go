package years_test

import (
	"github.com/amberpixels/years"
	"github.com/expectto/be/be_time"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"path/filepath"
	"time"
)

var _ = Describe("Voyager", func() {
	Context("Calendar 2006/Jan/2006-01-02.txt", func() {
		const TestCalendarLayout = "2006/Jan/2006-01-02.txt"
		var CalendarPath = filepath.Join(TestDataPath, "calendar")

		var v *years.Voyager

		BeforeEach(func() {
			v = years.NewVoyager(TestCalendarLayout)
			Expect(v.Prepare(CalendarPath)).To(Succeed())
		})

		It("should generate a proper waypoints tree", func() {
			wt := v.WaypointsTree()
			Expect(wt).NotTo(BeNil())

			// Root (calendar) folder has no time attached yet
			Expect(wt.Path).To(Equal(CalendarPath))
			Expect(wt.Name).To(Equal("calendar"))
			Expect(wt.Unit).To(Equal(years.UnitUndefined))
			Expect(wt.Time).To(BeZero())
			Expect(wt.Waypoints).To(HaveLen(1))

			// Year 2024 waypoint:
			wYear2024 := wt.Waypoints[0]
			Expect(wYear2024.Path).To(Equal(CalendarPath + "/2024"))
			Expect(wYear2024.Name).To(Equal("2024"))
			Expect(wYear2024.Unit).To(Equal(years.Year))
			Expect(wYear2024.Time).To(be_time.Year(2024))
			Expect(wYear2024.Waypoints).To(HaveLen(3))

			// Months waypoints (Date ordered):

			wJan := wYear2024.Waypoints[0]
			Expect(wJan.Path).To(Equal(CalendarPath + "/2024/Jan"))
			Expect(wJan.Name).To(Equal("Jan"))
			Expect(wJan.Unit).To(Equal(years.Month))
			Expect(wJan.Time).To(And(
				be_time.Month(time.January),
				be_time.Year(2024),
			))

			Expect(wJan.Waypoints).To(HaveLen(0))

			wFeb := wYear2024.Waypoints[1]
			Expect(wFeb.Path).To(Equal(CalendarPath + "/2024/Feb"))
			Expect(wFeb.Name).To(Equal("Feb"))
			Expect(wFeb.Unit).To(Equal(years.Month))
			Expect(wFeb.Time).To(And(
				be_time.Month(time.February),
				be_time.Year(2024),
			))
			Expect(wFeb.Waypoints).To(HaveLen(1))

			wMarch := wYear2024.Waypoints[2]
			Expect(wMarch.Path).To(Equal(CalendarPath + "/2024/Mar"))
			Expect(wMarch.Name).To(Equal("Mar"))
			Expect(wMarch.Unit).To(Equal(years.Month))
			Expect(wMarch.Time).To(And(
				be_time.Month(time.March),
				be_time.Year(2024),
			))

			Expect(wMarch.Waypoints).To(HaveLen(2))

			// Days:
			wFeb01 := wFeb.Waypoints[0]
			Expect(wFeb01.Path).To(Equal(CalendarPath + "/2024/Feb/2024-02-01.txt"))
			Expect(wFeb01.Name).To(Equal("2024-02-01.txt"))
			Expect(wFeb01.Unit).To(Equal(years.Day))
			Expect(wFeb01.Time).To(And(
				be_time.Day(1),
				be_time.Month(time.February),
				be_time.Year(2024),
			))
			Expect(wFeb01.Waypoints).To(BeEmpty())

			wMarch05 := wMarch.Waypoints[0]
			Expect(wMarch05.Path).To(Equal(CalendarPath + "/2024/Mar/2024-03-05.txt"))
			Expect(wMarch05.Name).To(Equal("2024-03-05.txt"))
			Expect(wMarch05.Unit).To(Equal(years.Day))
			Expect(wMarch05.Time).To(And(
				be_time.Day(5),
				be_time.Month(time.March),
				be_time.Year(2024),
			))
			Expect(wMarch05.Waypoints).To(BeEmpty())

			wMarch06 := wMarch.Waypoints[1]
			Expect(wMarch06.Path).To(Equal(CalendarPath + "/2024/Mar/2024-03-06.txt"))
			Expect(wMarch06.Name).To(Equal("2024-03-06.txt"))
			Expect(wMarch06.Unit).To(Equal(years.Day))
			Expect(wMarch05.Time).To(And(
				be_time.Day(5),
				be_time.Month(time.March),
				be_time.Year(2024),
			))
			Expect(wMarch06.Time).NotTo(BeZero())
			Expect(wMarch06.Waypoints).To(BeEmpty())
		})

		It("should traverse files Past->Future", func() {
			paths := make([]string, 0)
			v.Traverse(func(w *years.Waypoint) {
				paths = append(paths, w.Path)
			}, years.O_FUTURE(), years.O_FILES_ONLY())

			Expect(paths).To(Equal([]string{
				"internal/testdata/calendar/2024/Feb/2024-02-01.txt",
				"internal/testdata/calendar/2024/Mar/2024-03-05.txt",
				"internal/testdata/calendar/2024/Mar/2024-03-06.txt",
			}))
		})

		It("should traverse files Future->Past", func() {
			paths := make([]string, 0)
			v.Traverse(func(w *years.Waypoint) {
				paths = append(paths, w.Path)
			}, years.O_PAST(), years.O_FILES_ONLY())

			Expect(paths).To(Equal([]string{
				"internal/testdata/calendar/2024/Mar/2024-03-06.txt",
				"internal/testdata/calendar/2024/Mar/2024-03-05.txt",
				"internal/testdata/calendar/2024/Feb/2024-02-01.txt",
			}))
		})

		It("should traverse directories Past->Future", func() {
			paths := make([]string, 0)
			v.Traverse(func(w *years.Waypoint) {
				paths = append(paths, w.Path)
			}, years.O_FUTURE(), years.O_DIRS_ONLY())

			Expect(paths).To(Equal([]string{
				"internal/testdata/calendar/2024",
				"internal/testdata/calendar/2024/Jan",
				"internal/testdata/calendar/2024/Feb",
				"internal/testdata/calendar/2024/Mar",
			}))
		})

		It("should traverse directories Future->Past", func() {
			paths := make([]string, 0)
			v.Traverse(func(w *years.Waypoint) {
				paths = append(paths, w.Path)
			}, years.O_PAST(), years.O_DIRS_ONLY())

			Expect(paths).To(Equal([]string{
				"internal/testdata/calendar/2024/Mar",
				"internal/testdata/calendar/2024/Feb",
				"internal/testdata/calendar/2024/Jan",
				"internal/testdata/calendar/2024",
			}))
		})

		It("should traverse all nodes Future->Past", func() {
			paths := make([]string, 0)
			v.Traverse(func(w *years.Waypoint) {
				paths = append(paths, w.Path)
			}, years.O_PAST(), years.O_ALL())

			Expect(paths).To(Equal([]string{
				"internal/testdata/calendar/2024/Mar/2024-03-06.txt",
				"internal/testdata/calendar/2024/Mar/2024-03-05.txt",
				"internal/testdata/calendar/2024/Mar",
				"internal/testdata/calendar/2024/Feb/2024-02-01.txt",
				"internal/testdata/calendar/2024/Feb",
				"internal/testdata/calendar/2024/Jan",
				"internal/testdata/calendar/2024",
			}))
		})

		It("should traverse all nodes Past->Future", func() {
			paths := make([]string, 0)
			v.Traverse(func(w *years.Waypoint) {
				paths = append(paths, w.Path)
			}, years.O_FUTURE(), years.O_ALL())

			Expect(paths).To(Equal([]string{
				"internal/testdata/calendar/2024",
				"internal/testdata/calendar/2024/Jan",
				"internal/testdata/calendar/2024/Feb",
				"internal/testdata/calendar/2024/Feb/2024-02-01.txt",
				"internal/testdata/calendar/2024/Mar",
				"internal/testdata/calendar/2024/Mar/2024-03-05.txt",
				"internal/testdata/calendar/2024/Mar/2024-03-06.txt",
			}))
		})
	})
})
