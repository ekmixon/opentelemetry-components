package couchbasereceiver

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCouchbaseClient(t *testing.T) {
	cfg := couchbaseConfig{}
	client, err := newCouchbaseClient(cfg)
	assert.Nil(t, err)
	fmt.Printf("%#v\n", client)

}
