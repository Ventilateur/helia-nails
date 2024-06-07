package models

var (
	svTimestamp = map[string]string{
		".sv": "timestamp",
	}
)

type GenericResponse struct {
	Status string `json:"s,omitempty"`
}
