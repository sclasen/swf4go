package swf

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"
	"unicode"
	"os"
)

type WorkflowClient interface {
	StartWorkflow(request StartWorkflowRequest) (*StartWorkflowResponse, error)
	SignalWorkflow(request SignalWorkflowRequest) error
	ListWorkflowTypes(request ListWorkflowTypesRequest) (*ListWorkflowTypesResponse, error)
	RequestCancelWorkflowExecution(request RequestCancelWorkflowExecution) error
	TerminateWorkflowExecution(request TerminateWorkflowExecution) error
}

type DecisionWorkerClient interface {
	PollForDecisionTask(request PollForDecisionTaskRequest) (*PollForDecisionTaskResponse, error)
	RespondDecisionTaskCompleted(request RespondDecisionTaskCompletedRequest) error
}

type ActivityWorkerClient interface {
	PollForActivityTask(request PollForActivityTaskRequest) (*PollForActivityTaskResponse, error)
	RecordActivityTaskHeartbeat(request RecordActivityTaskHeartbeatRequest) (*RecordActivityTaskHeartbeatResponse, error)
	RespondActivityTaskCompleted(request RespondActivityTaskCompletedRequest) error
	RespondActivityTaskFailed(request RespondActivityTaskFailedRequest) error
	RespondActivityTaskCanceled(request RespondActivityTaskFailedRequest) error
}

type Region struct {
	Name     string
	Endpoint string
}

func (r *Region) URL() string {
	return r.Endpoint
}

var (
	USEast1      *Region = &Region{"us-east-1", "https://swf.us-east-1.amazonaws.com"}
	USWest1      *Region = &Region{"us-west-1", "https://swf.us-west-1.amazonaws.com"}
	USWest2      *Region = &Region{"us-west-2", "https://swf.us-west-2.amazonaws.com"}
	EUWest1      *Region = &Region{"eu-west-1", "https://swf.eu-west-1.amazonaws.com"}
	APNorthEast1 *Region = &Region{"ap-northeast-1", "https://swf.ap-northeast-1.amazonaws.com"}
	APSouthEast1 *Region = &Region{"ap-southeast-1", "https://swf.ap-southeast-1.amazonaws.com"}
)

type Client struct {
	keys   *Keys
	Region *Region
}

func NewClient(key string, secret string, region *Region) *Client {
	return &Client{
		keys: &Keys{AccessKey:key, SecretKey:secret},
		Region: region,
	}
}

func (c *Client) Service() *Service {
	return &Service{Name: "swf", Region: c.Region.Name}
}

func (c *Client) StartWorkflow(request StartWorkflowRequest) (*StartWorkflowResponse, error) {
	resp := &StartWorkflowResponse{}
	err := c.swfReqWithResponse("StartWorkflowExecution", request, resp)
	return resp, err
}

func (c *Client) SignalWorkflow(request SignalWorkflowRequest) error {
	err := c.swfReqNoResponse("SignalWorkflowExecution", request)
	return err
}

func (c *Client) ListWorkflowTypes(request ListWorkflowTypesRequest) (*ListWorkflowTypesResponse, error) {
	resp := new(ListWorkflowTypesResponse)
	err := c.swfReqWithResponse("ListWorkflowTypes", request, resp)
	return resp, err
}

func (c *Client) RequestCancelWorkflowExecution(request RequestCancelWorkflowExecution) error {
	err := c.swfReqNoResponse("RequestCancelWorkflowExecution", request)
	return err
}

func (c *Client) TerminateWorkflowExecution(request TerminateWorkflowExecution) error {
	err := c.swfReqNoResponse("TerminateWorkflowExecution", request)
	return err
}

func (c *Client) PollForDecisionTask(request PollForDecisionTaskRequest) (*PollForDecisionTaskResponse, error) {
	resp := &PollForDecisionTaskResponse{}
	err := c.swfReqWithResponse("PollForDecisionTask", request, resp)
	return resp, err
}

func (c *Client) RespondDecisionTaskCompleted(request RespondDecisionTaskCompletedRequest) error {
	err := c.swfReqNoResponse("RespondDecisionTaskCompleted", request)
	return err
}

func (c *Client) PollForActivityTask(request PollForActivityTaskRequest) (*PollForActivityTaskResponse, error) {
	resp := &PollForActivityTaskResponse{}
	err := c.swfReqWithResponse("PollForActivityTask", request, resp)
	return resp, err
}

func (c *Client) RespondActivityTaskCompleted(request RespondActivityTaskCompletedRequest) error {
	err := c.swfReqNoResponse("RespondActivityTaskCompleted", request)
	return err
}

func (c *Client) RespondActivityTaskFailed(request RespondActivityTaskFailedRequest) error {
	err := c.swfReqNoResponse("RespondActivityTaskFailed", request)
	return err
}

func (c *Client) RespondActivityTaskCanceled(request RespondActivityTaskFailedRequest) error {
	err := c.swfReqNoResponse("RespondActivityTaskCanceled", request)
	return err
}

func (c *Client) RecordActivityTaskHeartbeat(request RecordActivityTaskHeartbeatRequest) (*RecordActivityTaskHeartbeatResponse, error) {
	resp := &RecordActivityTaskHeartbeatResponse{}
	err := c.swfReqWithResponse("RecordActivityTaskHeartbeat", request, resp)
	return resp, err
}

func (c *Client) swfReqWithResponse(operation string, request interface{}, response interface{}) error {
	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(request); err != nil {
		return err
	}
	req, err := http.NewRequest("POST", c.Region.URL(), &b)
	if err != nil {
		return err
	}

	req.Header.Set("Date", time.Now().UTC().Format(http.TimeFormat))
	req.Header.Set("X-Amz-Target", "SimpleWorkflowService."+operation)
	req.Header.Set("Content-Type", "application/x-amz-json-1.0")
	req.Header.Set("Connection", "Keep-Alive")

	err = c.Sign(req)
	if err != nil {
		return err
	}

	out, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		return err
	}
	Multiln(string(out))


	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	out, err = httputil.DumpResponse(resp, true)
	if err != nil {
		return err
	}
	Multiln(string(out))


	if resp.StatusCode != 200 {
		return errors.New("non 200")
	}

	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(response)
	log.Printf("%+v", response)
	return err
}

func (c *Client) swfReqNoResponse(operation string, request interface{}) error {
	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(request); err != nil {
		return err
	}
	req, err := http.NewRequest("POST", c.Region.URL(), &b)
	if err != nil {
		return err
	}

	req.Header.Set("Date", time.Now().UTC().Format(http.TimeFormat))
	req.Header.Set("X-Amz-Target", "SimpleWorkflowService."+operation)
	req.Header.Set("Content-Type", "application/x-amz-json-1.0")
	req.Header.Set("Connection", "Keep-Alive")

	err = c.Sign(req)
	if err != nil {
		return err
	}

	out, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		return err
	}
	Multiln(string(out))


	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	out, err = httputil.DumpResponse(resp, true)
	if err != nil {
		return err
	}
	Multiln(string(out))


	if resp.StatusCode != 200 {
		return errors.New("non 200")
	}

	return nil
}

func (c *Client) Sign(request *http.Request) error {
	return c.Service().Sign(c.keys, request)
}


func Multiln(s string) {
	lines := strings.Split(s, "\n")
	for _, line := range lines {
		ln := strings.TrimRightFunc(line, unicode.IsSpace)
		log.Printf("line=%q\n", ln)
	}
}

func MustGetenv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic("Missing key " + key)
	}

	return value
}
