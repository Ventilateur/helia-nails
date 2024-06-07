package models

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Ventilateur/helia-nails/config"
	coremodels "github.com/Ventilateur/helia-nails/core/models"
	"github.com/Ventilateur/helia-nails/utils"
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
	TreatwellFeePercentage             float64       `json:"treatwellFeePercentage"`
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

func (a *Appointment) StartAt() time.Time {
	t, err := time.Parse(time.DateTime, fmt.Sprintf("%s %s:00", a.AppointmentDate, a.StartTime))
	if err != nil {
		panic(fmt.Errorf("invalid date time for [%s %s]: %w", a.AppointmentDate, a.StartTime, err))
	}

	return t
}

func (a *Appointment) EndAt() time.Time {
	t, err := time.Parse(time.DateTime, fmt.Sprintf("%s %s:00", a.AppointmentDate, a.EndTime))
	if err != nil {
		panic(fmt.Errorf("invalid date time for [%s %s]: %w", a.AppointmentDate, a.EndTime, err))
	}

	return t
}

func (a *Appointment) CoreModel(config *config.Config) coremodels.Appointment {
	source, id := utils.ParseCustomID(a.Notes)

	return coremodels.Appointment{
		Source: source,
		Ids: coremodels.AppointmentIds{
			Treatwell: strconv.Itoa(a.Id),
			Planity: func() string {
				if source == coremodels.SourcePlanity {
					return id
				}
				return ""
			}(),
			Classpass: func() string {
				if source == coremodels.SourceClassPass {
					return id
				}
				return ""
			}(),
		},
		Employee:   config.GetEmployee(coremodels.SourceTreatwell, strconv.Itoa(a.Id)),
		Service:    config.GetService(coremodels.SourceTreatwell, strconv.Itoa(a.OfferId), strconv.Itoa(a.SkuId)),
		StartTime:  a.StartAt(),
		EndTime:    a.EndAt(),
		ClientName: a.ConsumerName,
		Notes:      a.Notes,
	}
}
