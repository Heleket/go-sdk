package transport

import "encoding/json"

type JSONAdapter struct{}

func (JSONAdapter) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (JSONAdapter) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
