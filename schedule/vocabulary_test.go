package schedule_test

import (
	"testing"
	"time"

	"github.com/expectto/be"
	"github.com/expectto/be/be_string"

	"github.com/amberpixels/years/schedule"
)

func TestParsePatterns_Weekdays(t *testing.T) {
	// Every weekday term round-trips, singular and plural. The known week
	// (2024-01-14 is a Sunday) pins term -> matching day.
	weekdays := map[string]int{
		"sunday": 14, "monday": 15, "tuesday": 16, "wednesday": 17,
		"thursday": 18, "friday": 19, "saturday": 20,
	}
	for term, day := range weekdays {
		for _, input := range []string{term, term + "s"} {
			patterns, err := schedule.ParsePatterns(input)
			be.Require(t, err).To(be.Succeed(), input)
			be.Expect(t, patterns.Contains(at(day, 12, 0))).To(be.True(), input)
			be.Expect(t, patterns.Contains(at(day+1, 12, 0))).To(be.False(), input)
		}
	}
}

func TestParsePatterns_DaySets(t *testing.T) {
	patterns, err := schedule.ParsePatterns("weekends")
	be.Require(t, err).To(be.Succeed())
	be.Expect(t, patterns.Contains(at(20, 12, 0))).To(be.True())  // Saturday
	be.Expect(t, patterns.Contains(at(21, 12, 0))).To(be.True())  // Sunday
	be.Expect(t, patterns.Contains(at(19, 12, 0))).To(be.False()) // Friday

	// All three synonyms mean Monday-Friday.
	for _, term := range []string{"workingdays", "workdays", "weekdays"} {
		patterns, err = schedule.ParsePatterns(term)
		be.Require(t, err).To(be.Succeed(), term)
		be.Expect(t, patterns.Contains(at(15, 12, 0))).To(be.True(), term)  // Monday
		be.Expect(t, patterns.Contains(at(19, 12, 0))).To(be.True(), term)  // Friday
		be.Expect(t, patterns.Contains(at(20, 12, 0))).To(be.False(), term) // Saturday
	}
}

func TestParsePatterns_DayParts(t *testing.T) {
	// Part names of the default partition resolve to every-day span patterns.
	patterns, err := schedule.ParsePatterns("mornings")
	be.Require(t, err).To(be.Succeed())
	be.Expect(t, patterns.Contains(at(15, 9, 0))).To(be.True())   // Monday morning
	be.Expect(t, patterns.Contains(at(21, 9, 0))).To(be.True())   // Sunday morning
	be.Expect(t, patterns.Contains(at(15, 14, 0))).To(be.False()) // afternoon

	// Nights wrap midnight.
	patterns, err = schedule.ParsePatterns("nights")
	be.Require(t, err).To(be.Succeed())
	be.Expect(t, patterns.Contains(at(15, 23, 0))).To(be.True())
	be.Expect(t, patterns.Contains(at(15, 2, 30))).To(be.True())
	be.Expect(t, patterns.Contains(at(15, 12, 0))).To(be.False())
}

func TestParsePatterns_CaseAndWhitespace(t *testing.T) {
	patterns, err := schedule.ParsePatterns("  SUNDAYS ,  Mornings ")
	be.Require(t, err).To(be.Succeed())
	be.Require(t, patterns).To(be.HaveLength(2))
}

func TestParsePatterns_Union(t *testing.T) {
	// "sundays,mornings" = sundays all day OR mornings every day - two
	// patterns, not one merged conjunction.
	patterns, err := schedule.ParsePatterns("sundays,mornings")
	be.Require(t, err).To(be.Succeed())
	be.Require(t, patterns).To(be.HaveLength(2))

	be.Expect(t, patterns.Contains(at(21, 18, 0))).To(be.True())  // Sunday evening
	be.Expect(t, patterns.Contains(at(16, 9, 0))).To(be.True())   // Tuesday morning
	be.Expect(t, patterns.Contains(at(16, 18, 0))).To(be.False()) // Tuesday evening
}

func TestParsePatterns_Errors(t *testing.T) {
	_, err := schedule.ParsePatterns("")
	be.Expect(t, err).To(be.HaveOccurred())

	_, err = schedule.ParsePatterns("   ")
	be.Expect(t, err).To(be.HaveOccurred())

	// Unknown terms list the known vocabulary (these surface in HTTP 400s).
	_, err = schedule.ParsePatterns("cheesecake")
	be.Require(t, err).To(be.HaveOccurred())
	be.Expect(t, err.Error()).To(be_string.ContainingSubstring(`unknown pattern term "cheesecake"`))
	be.Expect(t, err.Error()).To(be_string.ContainingSubstring("known terms:"))
	be.Expect(t, err.Error()).To(be_string.ContainingSubstring("weekend"))
	be.Expect(t, err.Error()).To(be_string.ContainingSubstring("lunchtime"))

	// One bad term poisons the whole input.
	_, err = schedule.ParsePatterns("sundays,cheesecake")
	be.Expect(t, err).To(be.HaveOccurred())

	// Combined day-part grammar is reserved, not silently misparsed.
	_, err = schedule.ParsePatterns("sunday-mornings")
	be.Require(t, err).To(be.HaveOccurred())
	be.Expect(t, err.Error()).To(be_string.ContainingSubstring("not supported yet"))
}

func TestVocabulary_Extra(t *testing.T) {
	vocab := schedule.DefaultVocabulary()
	vocab.Extra = map[string]schedule.WeekPatterns{
		"OfficeHours": {{
			Days: []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday},
			Spans: []schedule.DaySlot{{
				Start: schedule.TimeOfDay{Hour: 9, Minute: 0},
				End:   schedule.TimeOfDay{Hour: 17, Minute: 0},
			}},
		}},
	}

	// Extra terms are canonicalized: lookup is case-insensitive and plural-tolerant.
	patterns, err := vocab.ParsePatterns("officehours")
	be.Require(t, err).To(be.Succeed())
	be.Expect(t, patterns.Contains(at(15, 10, 0))).To(be.True())  // Monday 10:00
	be.Expect(t, patterns.Contains(at(15, 18, 0))).To(be.False()) // Monday 18:00
	be.Expect(t, patterns.Contains(at(20, 10, 0))).To(be.False()) // Saturday 10:00

	// The built-in vocabulary still resolves alongside Extra.
	patterns, err = vocab.ParsePatterns("officehours,weekends")
	be.Require(t, err).To(be.Succeed())
	be.Expect(t, patterns.Contains(at(20, 10, 0))).To(be.True())
}

func TestVocabulary_CustomPartition(t *testing.T) {
	halves := schedule.DayPartition{
		{Name: "am", Span: schedule.DaySlot{
			Start: schedule.TimeOfDay{Hour: 0, Minute: 0},
			End:   schedule.TimeOfDay{Hour: 12, Minute: 0},
		}},
		{Name: "pm", Span: schedule.DaySlot{
			Start: schedule.TimeOfDay{Hour: 12, Minute: 0},
			End:   schedule.TimeOfDay{Hour: 24, Minute: 0},
		}},
	}
	be.Require(t, halves.Validate()).To(be.Succeed())

	vocab := schedule.Vocabulary{Partition: halves}

	patterns, err := vocab.ParsePatterns("ams")
	be.Require(t, err).To(be.Succeed())
	be.Expect(t, patterns.Contains(at(15, 3, 0))).To(be.True())
	be.Expect(t, patterns.Contains(at(15, 15, 0))).To(be.False())

	// The default partition's parts are not part of a custom vocabulary...
	_, err = vocab.ParsePatterns("mornings")
	be.Expect(t, err).To(be.HaveOccurred())

	// ...but day-set terms always are.
	_, err = vocab.ParsePatterns("weekends")
	be.Expect(t, err).To(be.Succeed())
}
