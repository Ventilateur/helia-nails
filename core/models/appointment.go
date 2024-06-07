package models

import (
	"fmt"
	"time"
)

type Source string

const (
	SourceTreatwell Source = "Treatwell"
	SourceClassPass Source = "ClassPass"
	SourcePlanity   Source = "Planity"
)

type Service struct {
	Name      string           `yaml:"name"`
	Treatwell TreatwellService `yaml:"treatwell"`
	Planity   PlanityService   `yaml:"planity"`
	Classpass ClasspassService `yaml:"classpass"`
}

type TreatwellService struct {
	OfferId int `yaml:"offerId"`
	SkuId   int `yaml:"skuId"`
}

type PlanityService struct {
	Id string `yaml:"id"`
}

type ClasspassService struct {
	PossibleNames []string `yaml:"possibleNames"`
}

type Appointment struct {
	Source Source

	Ids AppointmentIds

	Employee Employee
	Service  Service

	StartTime time.Time
	EndTime   time.Time

	ClientName string
	Notes      string
}

type AppointmentIds struct {
	Treatwell string
	Planity   string
	Classpass string
}

func (a Appointment) CustomNotes() string {
	return fmt.Sprintf("${%s:%s}\n%s", string(a.Source), a.SourceId(), a.Notes)
}

func (a Appointment) SourceId() string {
	return a.Id(a.Source)
}

func (a Appointment) Id(source Source) string {
	switch source {
	case SourceTreatwell:
		return a.Ids.Treatwell
	case SourcePlanity:
		return a.Ids.Planity
	case SourceClassPass:
		return a.Ids.Classpass
	default:
		return ""
	}
}

func (a Appointment) String() string {
	return fmt.Sprintf("[%s] [%s %s-%s] [Source: %s] [%s]",
		a.Employee.Name,
		a.StartTime.Format(time.DateOnly),
		a.StartTime.Format(time.TimeOnly),
		a.EndTime.Format(time.TimeOnly),
		a.Source,
		a.Service.Name,
	)
}