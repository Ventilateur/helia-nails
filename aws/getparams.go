package aws

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/Ventilateur/helia-nails/utils"
)

var httpClient = &http.Client{
	Timeout: 30 * time.Second,
}

type GetParamsResponse struct {
	Parameter struct {
		Value string `json:"Value"`
	} `json:"Parameter"`
}

func GetParam(keys ...string) (map[string]string, error) {
	m := map[string]string{}

	for _, key := range keys {
		u, err := url.Parse("http://localhost:2773/systemsmanager/parameters/get/?withDecryption=true")
		if err != nil {
			return nil, fmt.Errorf("failed to parse local parameter store server url: %w", err)
		}

		q := u.Query()
		q.Set("name", key)
		u.RawQuery = q.Encode()

		req, err := http.NewRequest(http.MethodGet, u.String(), nil)
		if err != nil {
			return nil, utils.RequestCreationErr(u.String(), err)
		}

		req.Header.Set("X-Aws-Parameters-Secrets-Token", os.Getenv("AWS_SESSION_TOKEN"))

		resp, err := httpClient.Do(req)
		if err != nil {
			return nil, err
		}

		r := &GetParamsResponse{}
		err = json.NewDecoder(resp.Body).Decode(r)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", utils.ErrUnmarshalJSON, err)
		}

		m[key] = r.Parameter.Value
	}

	return m, nil
}
