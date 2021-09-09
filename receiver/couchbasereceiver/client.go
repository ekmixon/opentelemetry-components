package couchbasereceiver

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.opentelemetry.io/collector/component"
	"go.uber.org/zap"
)

type client interface {
	Get() (*Stats, error)
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

// Stats is implemented to to group NodeStats and BucketsStats since BucketsStats depends upon NodeStats
type Stats struct {
	NodeStats
	BucketsStats
}

// NodeStats contains node stats about about all the nodes, which related to all the buckets. The field URI in Buckets struck contains a uuid, which is needed to call the getbucketsStats.
type NodeStats struct {
	Nodes []struct {
		SystemStats struct {
			CPUUtilizationRate *float64 `json:"cpu_utilization_rate"`
			SwapTotal          *float64 `json:"swap_total"`
			SwapUsed           *float64 `json:"swap_used"`
			MemTotal           *float64 `json:"mem_total"`
			MemFree            *float64 `json:"mem_free"`
		} `json:"systemStats"`
		InterestingStats struct {
			CurrItems     *float64 `json:"curr_items"`
			CurrItemsTot  *float64 `json:"curr_items_tot"`
			EpBgFetched   *float64 `json:"ep_bg_fetched"`
			MemUsed       *float64 `json:"mem_used"`
			CmdGet        *float64 `json:"cmd_get"`
			GetHits       *float64 `json:"get_hits"`
			Ops           *float64 `json:"ops"`
			IndexDataSize *float64 `json:"index_data_size"`
			IndexDiskSize *float64 `json:"index_disk_size"`
		} `json:"interestingStats"`
		Uptime string `json:"uptime"`
	} `json:"nodes"`
	Buckets struct {
		URI string `json:"uri"`
	} `json:"buckets"`
}

// BucketsStats contains a stats for each bucket instance.
type BucketsStats []struct {
	Name       string `json:"name"`
	BasicStats struct {
		QuotaPercentUsed *float64 `json:"quotaPercentUsed"`
		OpsPerSec        *float64 `json:"opsPerSec"`
		DiskFetches      *float64 `json:"diskFetches"`
		ItemCount        *float64 `json:"itemCount"`
		DiskUsed         *float64 `json:"diskUsed"`
		DataUsed         *float64 `json:"dataUsed"`
		MemUsed          *float64 `json:"memUsed"`
	} `json:"basicStats"`
}

// Get is exposesd by the client interface and returns the NodeStats and BucketsStats.
func (c *couchbaseClient) Get() (*Stats, error) {
	nodeStats, err := c.getNodeStats()
	if err != nil {
		return nil, err
	}
	bucketsStats, err := c.getBucketsStats(nodeStats.Buckets.URI)
	if err != nil {
		return nil, err
	}
	stats := Stats{
		NodeStats:    *nodeStats,
		BucketsStats: *bucketsStats}

	fmt.Printf("%#v\n", stats)
	return &stats, nil
}

func (c *couchbaseClient) getNodeStats() (*NodeStats, error) {
	url := fmt.Sprintf("%s%s", c.cfg.Endpoint, "/pools/default")
	req, err := http.NewRequest("GET", url, nil)
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	nodeStats := &NodeStats{}
	err = json.Unmarshal(body, nodeStats)
	if err != nil {
		panic("failed to marshal post request")
	}

	return nodeStats, nil
}

func (c *couchbaseClient) getBucketsStats(uri string) (*BucketsStats, error) {
	url := fmt.Sprintf("%s%s", c.cfg.Endpoint, uri)
	req, err := http.NewRequest("GET", url, nil)
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	bucketsStats := &BucketsStats{}
	err = json.Unmarshal(body, bucketsStats)
	if err != nil {
		panic("failed to marshal post request")
	}

	return bucketsStats, nil
}
