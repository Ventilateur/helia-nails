package models

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/Ventilateur/helia-nails/utils"
)

type BookRequest struct {
	Path string                 `json:"p,omitempty"`
	D    map[string]BookDetails `json:"d,omitempty"`
}

type BookDetails struct {
	Id         string            `json:"sq,omitempty"`
	Title      string            `json:"t,omitempty"`
	Note       string            `json:"c,omitempty"`
	Start      string            `json:"s,omitempty"`
	Duration   int64             `json:"d,omitempty"`
	Position   int64             `json:"st,omitempty"` // number of minutes since 00:00
	EmployeeId string            `json:"ca,omitempty"`
	Cby        string            `json:"cby,omitempty"`
	Uat        map[string]string `json:"uat,omitempty"`
	Client     map[string]bool   `json:"cu,omitempty"`
	Cat        map[string]string `json:"cat,omitempty"`
}

func NewBookRequest(reqId int64, employeeId string, start time.Time, end time.Time, title string, note string) (string, *Message[BookRequest]) {
	id := randId()
	startOfDay := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
	return id, &Message[BookRequest]{
		Type: "d",
		Desc: MessageDescription[BookRequest]{
			RequestId: reqId,
			Action:    "m",
			Body: BookRequest{
				Path: "/",
				D: map[string]BookDetails{
					fmt.Sprintf("calendar_vevents/%s/%s", employeeId, id): {
						Id:         id,
						Title:      title,
						Note:       note,
						Start:      start.Format(utils.PlanityTimeFormat),
						Duration:   int64(end.Sub(start).Minutes()),
						Position:   int64(start.Sub(startOfDay).Minutes()),
						EmployeeId: employeeId,
						Cby:        "accès",
						Uat:        svTimestamp,
						Cat:        svTimestamp,
						Client: map[string]bool{
							"d": true,
						},
					},
				},
			},
		},
	}
}

func randId() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seed := rand.NewSource(time.Now().UnixNano())
	random := rand.New(seed)

	result := make([]byte, 16)
	for i := range result {
		result[i] = charset[random.Intn(len(charset))]
	}
	return fmt.Sprintf("-Nz_%s", string(result))
}
