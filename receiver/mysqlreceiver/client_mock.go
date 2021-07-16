package mysqlreceiver

import (
	"bufio"
	"os"
	"path"
	"strings"
)

var _ client = (*fakeClient)(nil)

type fakeClient struct {
}

func readFile(fname string) (map[string]string, error) {
	var stats = map[string]string{}
	file, err := os.Open(path.Join("testdata", fname+".txt"))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := strings.Split(scanner.Text(), "\t")
		stats[text[0]] = text[1]
	}
	return stats, nil
}

func (c *fakeClient) getGlobalStats() (map[string]string, error) {
	return readFile("global_stats")
}

func (c *fakeClient) getInnodbStats() (map[string]string, error) {
	return readFile("innodb_stats")
}

func (c *fakeClient) Closed() bool {
	return false
}

func (c *fakeClient) Close() error {
	return nil
}
