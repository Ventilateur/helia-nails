package models

import (
	"fmt"
	"strings"
	"time"
)

type GetCalendarRequest struct {
	Path  string                  `json:"p,omitempty"`
	Query GetCalendarRequestQuery `json:"q,omitempty"`
	T     int64                   `json:"t,omitempty"`
	H     string                  `json:"h,omitempty"`
}

type GetCalendarRequestQuery struct {
	Start string `json:"sp,omitempty"`
	End   string `json:"ep,omitempty"`
	Sin   bool   `json:"sin,omitempty"`
	Ein   bool   `json:"ein,omitempty"`
	I     string `json:"i,omitempty"`
}

func NewGetCalendarRequest(reqId int64, employeeId string, from, to time.Time) *Message[GetCalendarRequest] {
	return &Message[GetCalendarRequest]{
		Type: "d",
		Desc: MessageDescription[GetCalendarRequest]{
			RequestId: reqId,
			Action:    "q",
			Body: GetCalendarRequest{
				Path: fmt.Sprintf("/calendar_vevents/%s", employeeId),
				Query: GetCalendarRequestQuery{
					Start: from.Format(time.DateOnly),
					End:   fmt.Sprintf("%s\uF8FF", to.Format(time.DateOnly)),
					Sin:   true,
					Ein:   true,
					I:     "s",
				},
				T: 1, // TODO???
				H: "",
			},
		},
	}
}

type GetCalendarResponse struct {
	Path string `json:"p,omitempty"`
	D    map[string]Appointment
}

func (r GetCalendarResponse) EmployeeId() string {
	return strings.TrimSpace(strings.TrimPrefix(r.Path, "calendar_vevents/"))
}

type Appointment struct {
	Title      string `json:"t,omitempty"`
	Notes      string `json:"c,omitempty"`
	Start      string `json:"s,omitempty"`
	Duration   int64  `json:"d,omitempty"`
	ServiceId  string `json:"se,omitempty"`
	Client     Client `json:"cu,omitempty"`
	EmployeeId string `json:"ca,omitempty"`
	DeletedAt  *int64 `json:"dat,omitempty"`
}

type Client struct {
	Id   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}
