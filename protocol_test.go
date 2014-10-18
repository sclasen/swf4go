package swf

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestJson(t *testing.T) {

	request := &StartWorkflowRequest{
		Domain: "foo",
	}

	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(request); err != nil {
		t.Errorf("no encode")
	}
	decoded := &StartWorkflowRequest{}
	json.NewDecoder(bytes.NewReader(b.Bytes())).Decode(decoded)

	if decoded.Domain != "foo" {
		t.Errorf("not foo")
	}
}
