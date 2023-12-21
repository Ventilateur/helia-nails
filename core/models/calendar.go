package models

import "time"

type Source int

const (
	SourceTreatwell Source = iota
)

type Calendar struct {
	Appointments []Appointment `json:"appointments"`
}

type Appointment struct {
	Id           int       `json:"id"`
	Source       Source    `json:"source"`
	Platform     string    `json:"platform"`
	EmployeeId   int       `json:"employeeId"`
	EmployeeName string    `json:"employeeName"`
	StartTime    time.Time `json:"startTime"`
	EndTime      time.Time `json:"endTime"`
	OfferId      int       `json:"offerId"`
	OfferName    string    `json:"offerName"`
	Notes        string    `json:"notes,omitempty"`
	Extra        string
}
