package search

import "gopkg.in/yaml.v3"

func yamlUnmarshal(data []byte, v interface{}) error {
	return yaml.Unmarshal(data, v)
}
