package couchbasereceiver

import (
	"fmt"

	"github.com/couchbase/gocb/v2"
)

type client interface {
}

type couchbaseClient struct {
}

type couchbaseConfig struct {
	username string
	password string
}

func newCouchbaseClient(conf couchbaseConfig) (*couchbaseClient, error) {
	cluster, err := gocb.Connect(
		"localhost",
		gocb.ClusterOptions{
			Username: "otelu",
			Password: "otelpassword",
		})
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v\n", cluster)

	return &couchbaseClient{}, nil
}
