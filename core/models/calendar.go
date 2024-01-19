package models

import (
	"fmt"
	"time"
)

type Source string

const (
	SourceTreatwell Source = "Treatwell"
	SourceClassPass Source = "ClassPass"
)

type Platform string

const (
	PlatformGoogle    Platform = "Google"
	PlatformTreatwell Platform = "Treatwell"
)

type Appointment struct {
	Id         string    `json:"id"`
	Source     Source    `json:"source"`
	Employee   string    `json:"employee"`
	StartTime  time.Time `json:"startTime"`
	EndTime    time.Time `json:"endTime"`
	Offer      string    `json:"offer"`
	ClientName string    `json:"clientName"`
	Notes      string    `json:"notes"`

	OriginalPlatform Platform
	OriginalID       string
}

func (a Appointment) String() string {
	return fmt.Sprintf(
		"[Source:%s] [%s] [%s - %s] [Employee:%s] [Client:%s] [Offer:%s]",
		a.Source, a.StartTime.Format(time.DateOnly),
		a.StartTime.Format(time.TimeOnly), a.EndTime.Format(time.TimeOnly),
		a.Employee, a.ClientName, a.Offer,
	)
}
