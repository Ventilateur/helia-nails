package models

import "time"

type Source string

const (
	SourceTreatwell Source = "Treatwell"
	SourceClassPass Source = "ClassPass"
)

type Appointment struct {
	Id        string    `json:"id"`
	Source    Source    `json:"source"`
	Employee  string    `json:"employee"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
	Offer     string    `json:"offer"`
	Notes     string    `json:"notes,omitempty"`
}
