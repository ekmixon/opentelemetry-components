package couchdbreceiver

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"go.opentelemetry.io/collector/component"
	"go.uber.org/zap"
)

type client interface {
	Get() (map[string]interface{}, error)
}

type couchdbClient struct {
	client *http.Client
	cfg    *Config
	logger *zap.Logger
}

var _ client = (*couchdbClient)(nil)

func newCouchDBClient(host component.Host, cfg *Config, logger *zap.Logger) (*couchdbClient, error) {
	client, err := cfg.ToClient(host.GetExtensions())
	if err != nil {
		return nil, err
	}

	return &couchdbClient{
		client: client,
		cfg:    cfg,
		logger: logger,
	}, nil
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func (c *couchdbClient) Get() (map[string]interface{}, error) {
	req, err := http.NewRequest("GET", c.cfg.Endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Basic "+basicAuth(c.cfg.Username, c.cfg.Password))
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.Error("failed to close client response", zap.Error(err))
		}
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var bodyParsed interface{}
	err = json.Unmarshal(body, &bodyParsed)
	if err != nil {
		return nil, err
	}

	fields, ok := bodyParsed.(map[string]interface{})
	if !ok {
		return nil, err
	}
	return fields, nil
}
