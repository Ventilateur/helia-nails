package utils

import (
	"time"

	"github.com/Ventilateur/helia-nails/core/models"
)

const (
	PlanityTimeFormat = "2006-01-02 15:04"
	DefaultTimeFormat = "2006-01-02T15:04:05"
	DefaultIanaTz     = "Europe/Paris"
)

var (
	DefaultLocation *time.Location
)

func init() {
	var err error
	DefaultLocation, err = time.LoadLocation(DefaultIanaTz)
	if err != nil {
		panic(err)
	}
}

func ParseTimes(t1Str, t2Str string) (time.Time, time.Time, error) {
	t1, err := time.ParseInLocation(DefaultTimeFormat, t1Str, DefaultLocation)
	if err != nil {
		t1, err = time.ParseInLocation(time.RFC3339, t1Str, DefaultLocation)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
	}

	t2, err := time.ParseInLocation(DefaultTimeFormat, t2Str, DefaultLocation)
	if err != nil {
		t2, err = time.ParseInLocation(time.RFC3339, t2Str, DefaultLocation)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
	}

	return t1, t2, err
}

func ParseCustomID(s string) (models.Source, string) {
	matches := models.CustomIDRegex.FindStringSubmatch(s)

	if len(matches) > 0 {
		return models.Source(matches[1]), matches[2]
	}

	return "", ""
}

func BoD(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

func EoD(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 23, 59, 59, 0, t.Location())
}

func TimeWithLocation(in time.Time) time.Time {
	return time.Date(in.Year(), in.Month(), in.Day(), in.Hour(), in.Minute(), 0, 0, DefaultLocation)
}
