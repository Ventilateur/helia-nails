package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func SendRequest[T any](client *http.Client, method string, url string, headers map[string]string, body io.Reader) (*T, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send http request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode > 299 {
		bodyStr, err := io.ReadAll(res.Body)
		return nil, fmt.Errorf("failed request: code %d, message '%s': %w", res.StatusCode, bodyStr, err)
	}

	var authInfo T
	if err := json.NewDecoder(res.Body).Decode(&authInfo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &authInfo, nil
}
