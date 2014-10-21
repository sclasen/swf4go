package swf

import (
	"bytes"
	"encoding/json"
	"strings"
)

type StateSerializer interface {
	Serialize(state interface{}) (string, error)
	Deserialize(serialized string, state interface{}) error
}

type JsonStateSerializer struct{}

func (j JsonStateSerializer) Serialize(state interface{}) (string, error) {
	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(state); err != nil {
		return "", err
	}
	return b.String(), nil
}

func (j JsonStateSerializer) Deserialize(serialized string, state interface{}) error {
	err := json.NewDecoder(strings.NewReader(serialized)).Decode(state)
	return err
}
