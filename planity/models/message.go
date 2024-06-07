package models

type Message[T any] struct {
	Type string                `json:"t,omitempty"`
	Desc MessageDescription[T] `json:"d,omitempty"`
}

type MessageDescription[T any] struct {
	RequestId int64  `json:"r,omitempty"`
	Action    string `json:"a,omitempty"`
	Body      T      `json:"b,omitempty"`
}
