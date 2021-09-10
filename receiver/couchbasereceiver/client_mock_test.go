package couchbasereceiver

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
)

var _ client = (*fakeClient)(nil)

type fakeClient struct {
	nodeStatsFilename    string
	bucketsStatsFilename string
	err                  error
}

func (f *fakeClient) getNodeStats() (*NodeStats, error) {
	if f.nodeStatsFilename != "" {
		file, err := os.Open(path.Join("testdata", f.nodeStatsFilename))
		if err != nil {
			return nil, err
		}
		defer file.Close()

		body, err := ioutil.ReadAll(file)
		if err != nil {
			return nil, err
		}

		nodeStats := &NodeStats{}
		err = json.Unmarshal(body, nodeStats)
		if err != nil {
			return nil, err
		}

		return nodeStats, nil
	}
	return nil, f.err
}

func (f *fakeClient) getBucketsStats() (*BucketsStats, error) {
	if f.bucketsStatsFilename != "" {
		file, err := os.Open(path.Join("testdata", f.bucketsStatsFilename))
		if err != nil {
			return nil, err
		}
		defer file.Close()

		body, err := ioutil.ReadAll(file)
		if err != nil {
			return nil, err
		}

		bucketsStats := &BucketsStats{}
		err = json.Unmarshal(body, bucketsStats)
		if err != nil {
			return nil, err
		}

		return bucketsStats, nil
	}
	return nil, f.err
}

func (f *fakeClient) Get() (*Stats, error) {
	nodeStats, err := f.getNodeStats()
	if err != nil {
		return nil, err
	}

	bucketsStats, err := f.getBucketsStats()
	if err != nil {
		return nil, err
	}

	return &Stats{
		NodeStats:    *nodeStats,
		BucketsStats: *bucketsStats,
	}, nil
}
