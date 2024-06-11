package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/Ventilateur/helia-nails/utils"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
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
			return nil, fmt.Errorf("failed to unmarshal json: key=%s: %w", key, err)
		}

		m[key] = r.Parameter.Value
	}

	return m, nil
}

func SetParam(key, value string) error {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return err
	}

	c := ssm.NewFromConfig(cfg)
	_, err = c.PutParameter(ctx, &ssm.PutParameterInput{
		Name:      aws.String(key),
		Value:     aws.String(value),
		DataType:  aws.String("text"),
		Overwrite: aws.Bool(true),
		Type:      types.ParameterTypeSecureString,
	})

	if err != nil {
		return err
	}

	return nil
}
