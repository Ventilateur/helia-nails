package models

import "fmt"

type DeleteRequest struct {
	Path string         `json:"p,omitempty"`
	D    map[string]any `json:"d,omitempty"`
}

func NewDeleteRequest(reqId int64, userId string, employeeId string, appointmentId string) *Message[DeleteRequest] {
	return &Message[DeleteRequest]{
		Type: "d",
		Desc: MessageDescription[DeleteRequest]{
			RequestId: reqId,
			Action:    "m",
			Body: DeleteRequest{
				Path: "/",
				D: map[string]any{
					fmt.Sprintf("calendar_vevents/%s/%s/dat", employeeId, appointmentId): svTimestamp,
					fmt.Sprintf("calendar_vevents/%s/%s/uat", employeeId, appointmentId): svTimestamp,
					fmt.Sprintf("calendar_vevents/%s/%s/dby", employeeId, appointmentId): userId,
				},
			},
		},
	}
}
