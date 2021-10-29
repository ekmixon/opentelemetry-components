package mongodbreceiver

import (
	"context"
	"fmt"
	"io/ioutil"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.uber.org/zap"
)

var _ client = (*fakeClient)(nil)

type fakeClient struct{}

func createFakeClient(config *Config, logger *zap.Logger) (client, error) {
	return &fakeClient{}, nil
}

func (c *fakeClient) Disconnect(context.Context) error {
	return nil
}

func (c *fakeClient) query(ctx context.Context, database string, command bson.M) (bson.M, error) {
	var doc bson.M
	if database == "admin" {
		admin, err := ioutil.ReadFile("./testdata/admin.json")
		if err != nil {
			return nil, err
		}
		err = bson.UnmarshalExtJSON(admin, true, &doc)
		if err != nil {
			return nil, err
		}
		return doc, nil
	} else {
		for k := range command {
			if k == "dbStats" {
				dbStats, err := ioutil.ReadFile("./testdata/dbstats.json")
				if err != nil {
					return nil, err
				}
				err = bson.UnmarshalExtJSON(dbStats, true, &doc)
				if err != nil {
					return nil, err
				}
				return doc, nil
			} else if k == "serverStatus" {
				serverStatus, err := ioutil.ReadFile("./testdata/serverstatus.json")
				if err != nil {
					return nil, err
				}
				err = bson.UnmarshalExtJSON(serverStatus, true, &doc)
				if err != nil {
					return nil, err
				}
				return doc, nil
			}
		}
	}
	return nil, fmt.Errorf("document could not be found")
}

func (c *fakeClient) ListDatabaseNames(_ context.Context, _ interface{}, _ ...*options.ListDatabasesOptions) ([]string, error) {
	return []string{"fakedatabase"}, nil
}

func (c *fakeClient) Connect(_ context.Context) error {
	return nil
}

func (c *fakeClient) Ping(ctx context.Context, rp *readpref.ReadPref) error {
	return nil
}
