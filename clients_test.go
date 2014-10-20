package swf

import (
	"log"
	"testing"
)

func TestListWorkflowTypes(t *testing.T) {

	client := NewClient(MustGetenv("AWS_ACCESS_KEY_ID"), MustGetenv("AWS_SECRET_ACCESS_KEY"), USEast1)

	resp, err := client.ListWorkflowTypes(ListWorkflowTypesRequest{
		Domain:             "swf4go",
		RegistrationStatus: "REGISTERED",
	})

	if err != nil {
		t.Fail()
	}

	log.Printf("%+v", resp)

	count, err := client.CountOpenWorkflowExecutions(CountOpenWorkflowExecutionsRequest{
		Domain:"swf4go",
    })

	if err != nil {
		t.Fail()
	}

	log.Printf("%+v", count)
}
