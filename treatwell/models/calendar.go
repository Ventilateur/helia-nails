package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Ventilateur/helia-nails/core/models"
	ics "github.com/arran4/golang-ical"
)

type Calendar struct {
	Appointments []Appointment `json:"appointments"`

	// Unknown types for now
	//AppointmentGroups []interface{} `json:"appointmentGroups"`
	//Blocks            []interface{} `json:"blocks"`
}

type Sku struct {
	SkuId         int    `json:"skuId"`
	SkuName       string `json:"skuName"`
	NameInherited bool   `json:"nameInherited"`
}

type Appointment struct {
	Id                                 int           `json:"id"`
	AppointmentDate                    string        `json:"appointmentDate"`
	AppointmentStatusCode              string        `json:"appointmentStatusCode"`
	BookingActor                       string        `json:"bookingActor"`
	ChannelCode                        string        `json:"channelCode"`
	Platform                           string        `json:"platform"`
	CancellationPeriodEndDate          time.Time     `json:"cancellationPeriodEndDate"`
	ConsumerId                         int           `json:"consumerId"`
	VenueCustomerId                    int           `json:"venueCustomerId"`
	ConsumerName                       string        `json:"consumerName"`
	ConsumerFirstName                  string        `json:"consumerFirstName,omitempty"`
	ConsumerEmail                      string        `json:"consumerEmail,omitempty"`
	ConsumerNotes                      string        `json:"consumerNotes,omitempty"`
	Created                            time.Time     `json:"created"`
	CreatedByName                      string        `json:"createdByName"`
	Amount                             float64       `json:"amount"`
	CurrencyCode                       string        `json:"currencyCode"`
	EmployeeId                         int           `json:"employeeId"`
	EmployeeName                       string        `json:"employeeName"`
	StartTime                          string        `json:"startTime"`
	EndTime                            string        `json:"endTime"`
	NoShowTimeLimit                    time.Time     `json:"noShowTimeLimit"`
	OfferId                            int           `json:"offerId"`
	OfferName                          string        `json:"offerName"`
	TaxRate                            float64       `json:"taxRate"`
	PaymentProtected                   bool          `json:"paymentProtected"`
	PaymentProtectionApplied           bool          `json:"paymentProtectionApplied"`
	Skus                               []Sku         `json:"skus"`
	SkuId                              int           `json:"skuId"`
	OrderItemIds                       []interface{} `json:"orderItemIds"`
	AppointmentGroupIds                []interface{} `json:"appointmentGroupIds"`
	TreatwellFeePercentage             int           `json:"treatwellFeePercentage"`
	ConsumerConsentedForMarketingComms bool          `json:"consumerConsentedForMarketingComms"`
	ConsumerPrepaymentRequired         bool          `json:"consumerPrepaymentRequired"`
	WalkIn                             bool          `json:"walkIn"`
	PayAtVenue                         bool          `json:"payAtVenue"`
	NotifyConsumer                     bool          `json:"notifyConsumer"`
	Anonymous                          bool          `json:"anonymous"`
	EmployeeSelected                   bool          `json:"employeeSelected"`
	FirstTimeCustomer                  bool          `json:"firstTimeCustomer"`
	HasEvoucher                        bool          `json:"hasEvoucher"`
	AnonymousNote                      string        `json:"anonymousNote,omitempty"`
	Notes                              string        `json:"notes,omitempty"`
	ServiceId                          int           `json:"serviceId"`
}

func (a *Appointment) StartAt() (time.Time, error) {
	return time.Parse(time.DateTime, fmt.Sprintf("%s %s:00", a.AppointmentDate, a.StartTime))
}

func (a *Appointment) EndAt() (time.Time, error) {
	return time.Parse(time.DateTime, fmt.Sprintf("%s %s:00", a.AppointmentDate, a.EndTime))
}

func (c *Calendar) ICal() (string, error) {
	cal := ics.NewCalendar()

	for _, appointment := range c.Appointments {
		cal.Events()
		event := cal.AddEvent(fmt.Sprintf("%d", appointment.Id))

		start, err := appointment.StartAt()
		if err != nil {
			return "", err
		}

		end, err := appointment.EndAt()
		if err != nil {
			return "", err
		}

		event.SetStartAt(start)
		event.SetEndAt(end)

		b, err := json.Marshal(appointment)
		if err != nil {
			return "", err
		}

		event.SetLocation("TREATWELL")

		event.SetDescription(string(b))
	}

	return cal.Serialize(), nil
}

func (c *Calendar) CommonCalendar() (models.Calendar, error) {
	ret := models.Calendar{
		Appointments: make([]models.Appointment, len(c.Appointments)),
	}

	for _, a := range c.Appointments {
		start, err := a.StartAt()
		if err != nil {
			return ret, err
		}

		end, err := a.EndAt()
		if err != nil {
			return ret, err
		}

		ret.Appointments = append(ret.Appointments, models.Appointment{
			Id:           a.Id,
			Platform:     "TREATWELL",
			EmployeeId:   a.EmployeeId,
			EmployeeName: a.EmployeeName,
			StartTime:    start,
			EndTime:      end,
			OfferId:      a.OfferId,
			OfferName:    a.OfferName,
		})
	}

	return ret, nil
}
