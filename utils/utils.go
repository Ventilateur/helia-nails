package utils

import (
	"regexp"
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

	return t1.In(time.Local), t2.In(time.Local), err
}

func ParseCustomID(s string) (models.Source, string) {
	matches := customIDRegex.FindStringSubmatch(s)

	if len(matches) > 0 {
		return models.Source(matches[1]), matches[2]
	}

	return models.SourceTreatwell, ""
}
