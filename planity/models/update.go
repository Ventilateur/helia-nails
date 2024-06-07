package models

type UpdateRequest struct {
	Path string         `json:"p,omitempty"`
	D    map[string]any `json:"d,omitempty"`
}
