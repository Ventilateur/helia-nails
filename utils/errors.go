package utils

import (
	"errors"
	"fmt"
)

func RequestCreationErr(url string, err error) error {
	return fmt.Errorf("failed to create http request to %q: %w", url, err)
}

func DoRequestErr(url string, err error) error {
	return fmt.Errorf("failed to send http request to %q: %w", url, err)
}

func URLParseErr(url string, err error) error {
	return fmt.Errorf("failed to parse url %q: %w", url, err)
}

func UnexpectedErrorCode(code int) error {
	return fmt.Errorf("unexpected error code %d", code)
}

var (
	ErrUnmarshalJSON = errors.New("failed to unmarshal json")
)
