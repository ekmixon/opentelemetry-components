package couchbasereceiver

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"go.opentelemetry.io/collector/component"
	"go.uber.org/zap"
)

type client interface {
	Post([]Metric) error
}

type couchbaseClient struct {
	client *http.Client
	cfg    *Config
	logger *zap.Logger
}

var _ client = (*couchbaseClient)(nil)

func newCouchbaseClient(host component.Host, cfg *Config, logger *zap.Logger) (*couchbaseClient, error) {
	client, err := cfg.ToClient(host.GetExtensions())
	if err != nil {
		return nil, err
	}

	return &couchbaseClient{
		client: client,
		cfg:    cfg,
		logger: logger,
	}, nil
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

type MetricFields struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type PostRequest struct {
	Metric          []MetricFields `json:"metric"`
	ApplyFunctions  []string       `json:"applyFunctions"`
	AlignTimestamps bool           `json:"alignTimestamps"`
	Step            int            `json:"step"`
	Start           int            `json:"start"`
}

// NewPostRequest fills in metric names for the fields that are to be retrived via post request.
func NewPostRequest(value string) PostRequest {
	return PostRequest{
		Metric: []MetricFields{
			{
				Label: "name",
				Value: value,
			},
		},
		ApplyFunctions: []string{"avg"},
		Step:           60,
		Start:          -60,
	}
}

// NewPostRequests creates a NewPostRequest for each metric in Metrics.
func NewPostRequests(metrics []Metric) []PostRequest {
	postRequests := []PostRequest{}
	for _, metric := range metrics {
		postRequest := NewPostRequest(metric.Name)
		postRequests = append(postRequests, postRequest)
	}
	return postRequests
}

type ResponseBody []struct {
	Data []struct {
		Metric struct {
			Nodes []string `json:"nodes"`
		} `json:"metric"`
		Values [][]interface{} `json:"values"`
	} `json:"data"`
	Errors         []interface{} `json:"errors"`
	StartTimestamp int           `json:"startTimestamp"`
	EndTimestamp   int           `json:"endTimestamp"`
}

func (c *couchbaseClient) Post(metrics []Metric) error {

	postRequest := NewPostRequests(Metrics)
	jsonStr, err := json.Marshal(postRequest)
	if err != nil {
		panic("failed to marshal post request")
	}
	// fmt.Println(string(jsonStr))

	req, err := http.NewRequest("POST", c.cfg.Endpoint, bytes.NewBuffer(jsonStr))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Authorization", "Basic "+basicAuth(c.cfg.Username, c.cfg.Password))
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.Error("failed to close client response", zap.Error(err))
		}
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	responseBody := ResponseBody{}
	err = json.Unmarshal([]byte(body), &responseBody)
	if err != nil {
		panic(err)
	}

	PopulateMetrics(responseBody, Metrics)
	// fmt.Printf("\n\nMetrics!: %#v\n", Metrics)

	return nil
}

func cleanBrackets(rawString string) string {
	removedLeftBrackets := strings.ReplaceAll(rawString, "[", "")
	return strings.ReplaceAll(removedLeftBrackets, "]", "")
}

// PopulateMetrics populates the metrics with the responseBody.
// Since there are no labels, this function makes sure that the metric order
// remains the same.
// The values extracted create a time series, but we are only interested in the
// latest timestamp, therefore, we take the last index.
func PopulateMetrics(responseBody ResponseBody, metrics []Metric) {
	for i := 0; i < len(metrics); i++ {
		if len(responseBody[i].Errors) > 0 {
			errors := strings.Split(fmt.Sprintf("%v", responseBody[i].Errors), " ")
			metrics[i].ErrMessage = cleanBrackets(errors[len(errors)-1]) // gets the last timestamp value
		}

		if len(responseBody[i].Data[0].Values) > 0 {
			values := strings.Split(fmt.Sprintf("%v", responseBody[i].Data[0].Values), " ")
			metrics[i].Value = cleanBrackets(values[len(values)-1])
		}

		metrics[i].Timestamp = responseBody[i].EndTimestamp
	}
}
