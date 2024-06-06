package years_test

import (
	"github.com/amberpixels/years"
	"github.com/djherbis/times"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"path/filepath"
	"time"
)

var _ = Describe("Voyager", func() {
	Context("TimeNamedFileWaypoints", func() {
		Context("Calendar1", func() {
			const TestCalendarLayout = "2006/Jan/2006-01-02.txt"

			var CalendarPath = filepath.Join(TestDataPath, "calendar1")

			var wf *years.TimeNamedWaypointFile
			var v *years.Voyager
			BeforeEach(func() {
				var err error
				wf, err = years.NewTimeNamedWaypointFile(CalendarPath, TestCalendarLayout)
				Expect(err).To(Succeed())

				v = years.NewVoyager(wf)
			})

			Context("traversing", func() {
				It("should traverse it in Future / Leaves only", func() {
					identifiers := make([]string, 0)
					err := v.Traverse(func(w years.Waypoint) {
						identifiers = append(identifiers, w.Identifier())
					}, years.O_FUTURE(), years.O_LEAVES_ONLY())

					Expect(err).Should(Succeed())
					Expect(identifiers).To(Equal([]string{
						"internal/testdata/calendar1/2024/Feb/2024-02-01.txt",
						"internal/testdata/calendar1/2024/Mar/2024-03-05.txt",
						"internal/testdata/calendar1/2024/Mar/2024-03-06.txt",
					}))
				})

				It("should traverse it in Future / All nodes", func() {
					identifiers := make([]string, 0)
					err := v.Traverse(func(w years.Waypoint) {
						identifiers = append(identifiers, w.Identifier())
					}, years.O_FUTURE(), years.O_ALL())

					Expect(err).Should(Succeed())
					Expect(identifiers).To(Equal([]string{
						"internal/testdata/calendar1/2024",
						"internal/testdata/calendar1/2024/Jan",
						"internal/testdata/calendar1/2024/Feb",
						"internal/testdata/calendar1/2024/Feb/2024-02-01.txt",
						"internal/testdata/calendar1/2024/Mar",
						"internal/testdata/calendar1/2024/Mar/2024-03-05.txt",
						"internal/testdata/calendar1/2024/Mar/2024-03-06.txt",
					}))
				})

				It("should traverse it in Future / Containers only", func() {
					identifiers := make([]string, 0)
					err := v.Traverse(func(w years.Waypoint) {
						identifiers = append(identifiers, w.Identifier())
					}, years.O_FUTURE(), years.O_CONTAINERS_ONLY())

					Expect(err).Should(Succeed())
					Expect(identifiers).To(Equal([]string{
						"internal/testdata/calendar1/2024",
						"internal/testdata/calendar1/2024/Jan",
						"internal/testdata/calendar1/2024/Feb",
						"internal/testdata/calendar1/2024/Mar",
					}))
				})

				It("should traverse it in Pas / Leaves only", func() {
					identifiers := make([]string, 0)
					err := v.Traverse(func(w years.Waypoint) {
						identifiers = append(identifiers, w.Identifier())
					}, years.O_PAST(), years.O_LEAVES_ONLY())

					Expect(err).Should(Succeed())
					Expect(identifiers).To(Equal([]string{
						"internal/testdata/calendar1/2024/Mar/2024-03-06.txt",
						"internal/testdata/calendar1/2024/Mar/2024-03-05.txt",
						"internal/testdata/calendar1/2024/Feb/2024-02-01.txt",
					}))
				})

				It("should traverse it in Pas / Containers only", func() {
					identifiers := make([]string, 0)
					err := v.Traverse(func(w years.Waypoint) {
						identifiers = append(identifiers, w.Identifier())
					}, years.O_PAST(), years.O_CONTAINERS_ONLY())

					Expect(err).Should(Succeed())
					Expect(identifiers).To(Equal([]string{
						"internal/testdata/calendar1/2024/Mar",
						"internal/testdata/calendar1/2024/Feb",
						"internal/testdata/calendar1/2024/Jan",
						"internal/testdata/calendar1/2024",
					}))
				})

				It("should traverse it in Past / All nodes", func() {
					identifiers := make([]string, 0)
					err := v.Traverse(func(w years.Waypoint) {
						identifiers = append(identifiers, w.Identifier())
					}, years.O_PAST(), years.O_ALL())

					Expect(err).Should(Succeed())
					Expect(identifiers).To(Equal([]string{
						"internal/testdata/calendar1/2024/Mar/2024-03-06.txt",
						"internal/testdata/calendar1/2024/Mar/2024-03-05.txt",
						"internal/testdata/calendar1/2024/Mar",
						"internal/testdata/calendar1/2024/Feb/2024-02-01.txt",
						"internal/testdata/calendar1/2024/Feb",
						"internal/testdata/calendar1/2024/Jan",
						"internal/testdata/calendar1/2024",
					}))
				})
			})

			Context("navigating", func() {
				It("should navigate to a specific date", func() {
					navigated, err := v.Navigate("2024-03-06")
					Expect(err).Should(Succeed())
					Expect(navigated).NotTo(BeNil())
					Expect(navigated.Identifier()).To(Equal(filepath.Join(CalendarPath, "2024", "Mar", "2024-03-06.txt")))
				})

				It("should navigate to today", func() {
					mockClock := &StaticClock{
						now: time.Date(2024, time.March, 05, 14, 30, 59, 0, time.UTC),
					}

					years.SetStdClock(mockClock)
					navigated, err := v.Navigate("today")
					Expect(err).Should(Succeed())
					Expect(navigated).NotTo(BeNil())
					Expect(navigated.Identifier()).To(Equal(filepath.Join(CalendarPath, "2024", "Mar", "2024-03-05.txt")))
				})
			})
		})

		Context("Calendar2", func() {
			const TestCalendarLayout = "2006/Jan/02 Mon.txt"

			// Calendar2 is different in the manner of how final files are named.
			// here, on Calendar2 final files are not sufficient for knowing the date (so they require parent information)
			// e.g. 2006/Jan/01-Mon.txt
			var CalendarPath = filepath.Join(TestDataPath, "calendar2")

			var wf *years.TimeNamedWaypointFile
			var v *years.Voyager
			BeforeEach(func() {
				var err error
				wf, err = years.NewTimeNamedWaypointFile(CalendarPath, TestCalendarLayout)
				Expect(err).To(Succeed())

				v = years.NewVoyager(wf)
			})

			Context("traversing", func() {
				It("should traverse it in Future / Leaves only", func() {
					identifiers := make([]string, 0)
					err := v.Traverse(func(w years.Waypoint) {
						identifiers = append(identifiers, w.Identifier())
					}, years.O_FUTURE(), years.O_LEAVES_ONLY())

					Expect(err).Should(Succeed())
					Expect(identifiers).To(Equal([]string{
						"internal/testdata/calendar2/2024/Feb/01 Thu.txt",
						"internal/testdata/calendar2/2024/Mar/05 Tue.txt",
						"internal/testdata/calendar2/2024/Mar/06 Wed.txt",
					}))
				})

				It("should traverse it in Future / Containers only", func() {
					identifiers := make([]string, 0)
					err := v.Traverse(func(w years.Waypoint) {
						identifiers = append(identifiers, w.Identifier())
					}, years.O_FUTURE(), years.O_CONTAINERS_ONLY())

					Expect(err).Should(Succeed())
					Expect(identifiers).To(Equal([]string{
						"internal/testdata/calendar2/2024",
						"internal/testdata/calendar2/2024/Feb",
						"internal/testdata/calendar2/2024/Mar",
					}))
				})

				It("should traverse it in Future / All nodes", func() {
					identifiers := make([]string, 0)
					err := v.Traverse(func(w years.Waypoint) {
						identifiers = append(identifiers, w.Identifier())
					}, years.O_FUTURE(), years.O_ALL())

					Expect(err).Should(Succeed())
					Expect(identifiers).To(Equal([]string{
						"internal/testdata/calendar2/2024",
						"internal/testdata/calendar2/2024/Feb",
						"internal/testdata/calendar2/2024/Feb/01 Thu.txt",
						"internal/testdata/calendar2/2024/Mar",
						"internal/testdata/calendar2/2024/Mar/05 Tue.txt",
						"internal/testdata/calendar2/2024/Mar/06 Wed.txt",
					}))
				})

				It("should traverse it in Pas / Leaves only", func() {
					identifiers := make([]string, 0)
					err := v.Traverse(func(w years.Waypoint) {
						identifiers = append(identifiers, w.Identifier())
					}, years.O_PAST(), years.O_LEAVES_ONLY())

					Expect(err).Should(Succeed())
					Expect(identifiers).To(Equal([]string{
						"internal/testdata/calendar2/2024/Mar/06 Wed.txt",
						"internal/testdata/calendar2/2024/Mar/05 Tue.txt",
						"internal/testdata/calendar2/2024/Feb/01 Thu.txt",
					}))
				})

				It("should traverse it in Pas / Containers only", func() {
					identifiers := make([]string, 0)
					err := v.Traverse(func(w years.Waypoint) {
						identifiers = append(identifiers, w.Identifier())
					}, years.O_PAST(), years.O_CONTAINERS_ONLY())

					Expect(err).Should(Succeed())
					Expect(identifiers).To(Equal([]string{
						"internal/testdata/calendar2/2024/Mar",
						"internal/testdata/calendar2/2024/Feb",
						"internal/testdata/calendar2/2024",
					}))
				})

				It("should traverse it in Past / All nodes", func() {
					identifiers := make([]string, 0)
					err := v.Traverse(func(w years.Waypoint) {
						identifiers = append(identifiers, w.Identifier())
					}, years.O_PAST(), years.O_ALL())

					Expect(err).Should(Succeed())
					Expect(identifiers).To(Equal([]string{
						"internal/testdata/calendar2/2024/Mar/06 Wed.txt",
						"internal/testdata/calendar2/2024/Mar/05 Tue.txt",
						"internal/testdata/calendar2/2024/Mar",
						"internal/testdata/calendar2/2024/Feb/01 Thu.txt",
						"internal/testdata/calendar2/2024/Feb",
						"internal/testdata/calendar2/2024",
					}))
				})
			})

			Context("navigating", func() {
				It("should navigate to a specific date", func() {
					navigated, err := v.Navigate("2024-03-06")
					Expect(err).Should(Succeed())
					Expect(navigated).NotTo(BeNil())
					Expect(navigated.Identifier()).To(Equal(filepath.Join(CalendarPath, "2024", "Mar", "06 Wed.txt")))
				})

				It("should navigate to today", func() {
					mockClock := &StaticClock{
						now: time.Date(2024, time.March, 05, 14, 30, 59, 0, time.UTC),
					}

					years.SetStdClock(mockClock)
					navigated, err := v.Navigate("today")
					Expect(err).Should(Succeed())
					Expect(navigated).NotTo(BeNil())
					Expect(navigated.Identifier()).To(Equal(filepath.Join(CalendarPath, "2024", "Mar", "05 Tue.txt")))
				})
			})
		})

		Context("Logs via timestamp", func() {
			const TestCalendarLayout = "foobar_U@000.log"

			// Calendar2 is different in the manner of how final files are named.
			// here, on Calendar2 final files are not sufficient for knowing the date (so they require parent information)
			// e.g. 2006/Jan/01-Mon.txt
			var CalendarPath = filepath.Join(TestDataPath, "logs_via_timestamp")

			var wf *years.TimeNamedWaypointFile
			var v *years.Voyager
			BeforeEach(func() {
				var err error
				wf, err = years.NewTimeNamedWaypointFile(CalendarPath, TestCalendarLayout)
				Expect(err).To(Succeed())

				v = years.NewVoyager(wf)
			})

			Context("traversing", func() {
				It("should traverse it in Future / Leaves only", func() {
					identifiers := make([]string, 0)
					err := v.Traverse(func(w years.Waypoint) {
						identifiers = append(identifiers, w.Identifier())
					}, years.O_FUTURE(), years.O_LEAVES_ONLY())

					Expect(err).Should(Succeed())
					expectedIdentifiers := []string{
						"internal/testdata/logs_via_timestamp/foobar_1716559191.log",
						"internal/testdata/logs_via_timestamp/foobar_1716559238.log",
						"internal/testdata/logs_via_timestamp/foobar_1716559253.log",
						"internal/testdata/logs_via_timestamp/inner/foobar_1717669999.log",
					}

					Expect(identifiers).To(Equal(expectedIdentifiers))
				})

				It("should traverse it in Past / Leaves only", func() {
					identifiers := make([]string, 0)
					err := v.Traverse(func(w years.Waypoint) {
						identifiers = append(identifiers, w.Identifier())
					}, years.O_PAST(), years.O_LEAVES_ONLY())

					Expect(err).Should(Succeed())
					expectedIdentifiers := []string{
						"internal/testdata/logs_via_timestamp/inner/foobar_1717669999.log",
						"internal/testdata/logs_via_timestamp/foobar_1716559253.log",
						"internal/testdata/logs_via_timestamp/foobar_1716559238.log",
						"internal/testdata/logs_via_timestamp/foobar_1716559191.log",
					}

					Expect(identifiers).To(Equal(expectedIdentifiers))
				})

				It("should traverse it in Future / Containers only", func() {
					identifiers := make([]string, 0)
					err := v.Traverse(func(w years.Waypoint) {
						identifiers = append(identifiers, w.Identifier())
					}, years.O_FUTURE(), years.O_CONTAINERS_ONLY())

					Expect(err).Should(Succeed())
					// no containers actually
					expectedIdentifiers := []string{}

					Expect(identifiers).To(Equal(expectedIdentifiers))
				})

				It("should traverse it in Future / All only", func() {
					identifiers := make([]string, 0)
					err := v.Traverse(func(w years.Waypoint) {
						identifiers = append(identifiers, w.Identifier())
					}, years.O_FUTURE(), years.O_ALL())

					Expect(err).Should(Succeed())
					// no containers - so ALL means same as leaves only
					expectedIdentifiers := []string{
						"internal/testdata/logs_via_timestamp/foobar_1716559191.log",
						"internal/testdata/logs_via_timestamp/foobar_1716559238.log",
						"internal/testdata/logs_via_timestamp/foobar_1716559253.log",
						"internal/testdata/logs_via_timestamp/inner/foobar_1717669999.log"}

					Expect(identifiers).To(Equal(expectedIdentifiers))
				})
			})
		})
	})

	Context("FileWaypoints", func() {

		Context("BirthTime", func() {

			var CalendarPath = filepath.Join(TestDataPath, "by_filetime")
			var wf *years.WaypointFile
			var v *years.Voyager
			BeforeEach(func() {
				var err error
				wf, err = years.NewWaypointFile(CalendarPath, func(timeSpec times.Timespec) time.Time {
					return timeSpec.BirthTime()
				})
				Expect(err).To(Succeed())

				v = years.NewVoyager(wf)
			})

			Context("traversing", func() {
				It("should traverse it in Future / Leaves only", func() {
					identifiers := make([]string, 0)
					err := v.Traverse(func(w years.Waypoint) {
						identifiers = append(identifiers, w.Identifier())
					}, years.O_FUTURE(), years.O_LEAVES_ONLY())

					Expect(err).Should(Succeed())
					Expect(identifiers).To(Equal([]string{
						"internal/testdata/by_filetime/first.txt",
						"internal/testdata/by_filetime/foobar/second.txt",
						"internal/testdata/by_filetime/foobar/third.txt",
					}))
				})

				It("should traverse it in Past / Containers only", func() {
					identifiers := make([]string, 0)
					err := v.Traverse(func(w years.Waypoint) {
						identifiers = append(identifiers, w.Identifier())
					}, years.O_PAST(), years.O_CONTAINERS_ONLY())

					Expect(err).Should(Succeed())
					Expect(identifiers).To(Equal([]string{
						"internal/testdata/by_filetime/foobar",
						"internal/testdata/by_filetime",
					}))
				})

				It("should traverse it in Future / ALL", func() {
					identifiers := make([]string, 0)
					err := v.Traverse(func(w years.Waypoint) {
						identifiers = append(identifiers, w.Identifier())
					}, years.O_PAST(), years.O_ALL())

					Expect(err).Should(Succeed())
					Expect(identifiers).To(Equal([]string{
						"internal/testdata/by_filetime/foobar/third.txt",
						"internal/testdata/by_filetime/foobar/second.txt",
						"internal/testdata/by_filetime/first.txt",
						"internal/testdata/by_filetime/foobar",
						"internal/testdata/by_filetime",
					}))
				})
			})
		})

		Context("ModTime", func() {

			var CalendarPath = filepath.Join(TestDataPath, "by_filetime")
			var wf *years.WaypointFile
			var v *years.Voyager
			BeforeEach(func() {
				var err error
				wf, err = years.NewWaypointFile(CalendarPath, func(timeSpec times.Timespec) time.Time {
					return timeSpec.ModTime()
				})
				Expect(err).To(Succeed())

				v = years.NewVoyager(wf)
			})

			Context("traversing", func() {
				It("should traverse it in Future / Leaves only", func() {
					identifiers := make([]string, 0)
					err := v.Traverse(func(w years.Waypoint) {
						identifiers = append(identifiers, w.Identifier())
					}, years.O_FUTURE(), years.O_LEAVES_ONLY())

					Expect(err).Should(Succeed())
					Expect(identifiers).To(Equal([]string{
						"internal/testdata/by_filetime/first.txt",
						"internal/testdata/by_filetime/foobar/second.txt",
						"internal/testdata/by_filetime/foobar/third.txt",
					}))
				})

				It("should traverse it in Past / Containers only", func() {
					identifiers := make([]string, 0)
					err := v.Traverse(func(w years.Waypoint) {
						identifiers = append(identifiers, w.Identifier())
					}, years.O_PAST(), years.O_CONTAINERS_ONLY())

					Expect(err).Should(Succeed())
					Expect(identifiers).To(Equal([]string{
						"internal/testdata/by_filetime",
						"internal/testdata/by_filetime/foobar",
					}))
				})

				It("should traverse it in Future / ALL", func() {
					identifiers := make([]string, 0)
					err := v.Traverse(func(w years.Waypoint) {
						identifiers = append(identifiers, w.Identifier())
					}, years.O_PAST(), years.O_ALL())

					Expect(err).Should(Succeed())
					Expect(identifiers).To(Equal([]string{
						"internal/testdata/by_filetime",
						"internal/testdata/by_filetime/foobar",
						"internal/testdata/by_filetime/foobar/third.txt",
						"internal/testdata/by_filetime/foobar/second.txt",
						"internal/testdata/by_filetime/first.txt",
					}))
				})
			})
		})
	})

	Context("TimeStringWaypoints", func() {

		var ws = []years.Waypoint{
			years.NewWaypointTimeString("2024-03-05"),
			years.NewWaypointTimeString("2024-03-06"),
			years.NewWaypointTimeString("2024-03-07"),
			years.NewWaypointTimeString("2024-04"),
		}
		var v *years.Voyager
		BeforeEach(func() {
			v = years.NewVoyager(years.NewWaypointGroup("root", ws...))
		})

		Context("traversing", func() {
			It("should traverse it in Future ", func() {
				identifiers := make([]string, 0)
				err := v.Traverse(func(w years.Waypoint) {
					identifiers = append(identifiers, w.Identifier())
				}, years.O_FUTURE())

				Expect(err).Should(Succeed())
				Expect(identifiers).To(Equal([]string{
					"2024-03-05",
					"2024-03-06",
					"2024-03-07",
					"2024-04",
				}))
			})

			It("should traverse it in Past", func() {
				identifiers := make([]string, 0)
				err := v.Traverse(func(w years.Waypoint) {
					identifiers = append(identifiers, w.Identifier())
				}, years.O_PAST())

				Expect(err).Should(Succeed())
				Expect(identifiers).To(Equal([]string{
					"2024-04",
					"2024-03-07",
					"2024-03-06",
					"2024-03-05",
				}))
			})

		})

	})
})
