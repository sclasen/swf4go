package swf

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestListWorkflowTypes(t *testing.T) {
	if os.Getenv("AWS_ACCESS_KEY_ID") == "" || os.Getenv("AWS_SECRET_ACCESS_KEY") == "" {
		log.Printf("WARNING: NO AWS CREDS SPECIFIED, SKIPPING CLIENTS TEST")
		return
	}

	client := NewClientWithHTTPClient(MustGetenv("AWS_ACCESS_KEY_ID"), MustGetenv("AWS_SECRET_ACCESS_KEY"), USEast1, customHTTPClient())
	client.Debug = true
	resp, err := client.ListWorkflowTypes(ListWorkflowTypesRequest{
		Domain:             "swf4go",
		RegistrationStatus: "REGISTERED",
	})

	if err != nil {
		log.Printf("%+v", err)
		t.Fail()
	}

	log.Printf("%+v", resp)
	for _, i := range resp.TypeInfos {
		log.Println(i.CreationDate)
	}

	count, err := client.CountOpenWorkflowExecutions(CountOpenWorkflowExecutionsRequest{
		Domain:          "swf4go",
		StartTimeFilter: *ZeroTimeFilter(),
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

	client := NewClientWithHTTPClient(MustGetenv("AWS_ACCESS_KEY_ID"), MustGetenv("AWS_SECRET_ACCESS_KEY"), USEast1, customHTTPClient())
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

func customHTTPClient() *http.Client {
	return &http.Client{
		Transport: &LoggingRoundTripper{http.DefaultTransport},
	}
}

type LoggingRoundTripper struct {
	Transport http.RoundTripper
}

func (l *LoggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	log.Printf("request: target=%s", req.Header.Get("X-Amz-Target"))
	resp, err := l.Transport.RoundTrip(req)
	if err != nil {
		log.Printf("error %s", err)
	} else {
		log.Printf("response: status=%s", resp.Status)
	}
	return resp, err
}
