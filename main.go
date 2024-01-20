package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
)

type Event struct {
	Name string `json:"name"`
}

func HandleRequest(ctx context.Context, event *Event) (*string, error) {
	if event == nil {
		return nil, fmt.Errorf("received nil event")
	}

	switch event.Name {
	case "sync":
		if err := syncAll(ctx); err != nil {
			return nil, fmt.Errorf("failed to sync: %w", err)
		}
	default:
		return nil, fmt.Errorf("invalid event name %s", event.Name)
	}

	return nil, nil
}

func main() {
	lambda.Start(HandleRequest)
}
