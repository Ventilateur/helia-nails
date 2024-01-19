package utils

import (
	"regexp"
	"slices"
	"time"

	"github.com/Ventilateur/helia-nails/core/models"
)

const (
	defaultTimeFormat = "2006-01-02T15:04:05"
)

var (
	customIDRegex = regexp.MustCompile(`\${(\w+):([\dA-Za-z-_]+)}`)
)

func ParseFromToTimes(t1Str, t2Str string) (time.Time, time.Time, error) {
	t1, err := time.Parse(defaultTimeFormat, t1Str)
	if err != nil {
		t1, err = time.Parse(time.RFC3339, t1Str)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
	}

	t2, err := time.Parse(defaultTimeFormat, t2Str)
	if err != nil {
		t2, err = time.Parse(time.RFC3339, t2Str)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
	}

	return t1, t2, err
}

func ParseCustomID(s string) (models.Source, string) {
	matches := customIDRegex.FindStringSubmatch(s)

	if len(matches) > 0 {
		return models.Source(matches[1]), matches[2]
	}

	return models.SourceTreatwell, ""
}

func BoD(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

func EoD(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 23, 59, 59, 0, t.Location())
}

func MapToOrderedSlice(m map[string]models.Appointment) []models.Appointment {
	var ret []models.Appointment

	for _, a := range m {
		ret = append(ret, a)
	}

	slices.SortFunc(ret, func(a1, a2 models.Appointment) int {
		return a1.StartTime.Compare(a2.StartTime)
	})

	return ret
}
