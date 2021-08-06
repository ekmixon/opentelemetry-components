package couchdbreceiver

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
)

var _ client = (*fakeClient)(nil)

type fakeClient struct {
	filename string
	err      error
}

func (f *fakeClient) Get() (map[string]interface{}, error) {
	if f.filename != "" {
		file, err := os.Open(path.Join("testdata", f.filename))
		if err != nil {
			return nil, err
		}
		defer file.Close()

		body, err := ioutil.ReadAll(file)
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
	return nil, f.err
}
