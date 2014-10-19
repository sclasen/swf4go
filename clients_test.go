package swf

import (
	"testing"
	"log"
)

func TestListWorkflowTypes(t *testing.T) {

	request := ListWorkflowTypesRequest{
		Domain: "swf4go",
		RegistrationStatus: "REGISTERED",
	}

	client := NewClient(MustGetenv("AWS_ACCESS_KEY_ID"), MustGetenv("AWS_SECRET_ACCESS_KEY"), USEast1)

	resp, err := client.ListWorkflowTypes(request)

	if err != nil {
		t.Fail()
	}

	log.Printf("%+v", resp)
}
