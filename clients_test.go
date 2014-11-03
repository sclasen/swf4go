package swf

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"
)

func TestListWorkflowTypes(t *testing.T) {
	if os.Getenv("AWS_ACCESS_KEY_ID") == "" || os.Getenv("AWS_SECRET_ACCESS_KEY") == "" {
		log.Printf("WARNING: NO AWS CREDS SPECIFIED, SKIPPING CLIENTS TEST")
		return
	}

	client := NewClient(MustGetenv("AWS_ACCESS_KEY_ID"), MustGetenv("AWS_SECRET_ACCESS_KEY"), USEast1)
	client.Debug = true
	resp, err := client.ListWorkflowTypes(ListWorkflowTypesRequest{
		Domain:             "swf4go",
		RegistrationStatus: "REGISTERED",
	})

	if err != nil {
		t.Fail()
	}

	log.Printf("%+v", resp)

	count, err := client.CountOpenWorkflowExecutions(CountOpenWorkflowExecutionsRequest{
		Domain: "swf4go",
		StartTimeFilter: &TimeFilter{
			OldestDate: 0,
		},
	})

	if err != nil {
		log.Printf("%+v", err)
		t.Fail()
	}

	log.Printf("%+v", count)
}

func TestPutRecord(t *testing.T) {
	if os.Getenv("AWS_ACCESS_KEY_ID") == "" || os.Getenv("AWS_SECRET_ACCESS_KEY") == "" {
		log.Printf("WARNING: NO AWS CREDS SPECIFIED, SKIPPING CLIENTS TEST")
		return
	}

	client := NewClient(MustGetenv("AWS_ACCESS_KEY_ID"), MustGetenv("AWS_SECRET_ACCESS_KEY"), USEast1)
	client.Debug = true
	req := PutRecordRequest{
		Data:                      []byte("foo"),
		PartitionKey:              "the-key",
		SequenceNumberForOrdering: fmt.Sprintf("%d", time.Now().UnixNano()),
		StreamName:                "swf4go",
	}

	resp, err := client.PutRecord(req)

	if err != nil {
		log.Printf("%+v", err)
		t.Fail()
	}

	log.Printf("%+v", resp)

}
