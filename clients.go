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
)

type WorkflowClient interface {
	StartWorkflow(request StartWorkflowRequest) (*StartWorkflowResponse, error)
	SignalWorkflow(request SignalWorkflowRequest) error
}

type DecisionWorkerClient interface {
	PollForDecisionTask(request PollForDecisionTaskRequest) (*PollForDecisionTaskResponse, error)
	RespondDecisionTaskCompleted(request RespondDecisionTaskCompletedRequest) error
}

type ActivityWorkerClient interface {
	PollForActivityTask(request PollForActivityTaskRequest) (*PollForActivityTaskResponse, error)
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
	USEast1      *Region = &Region{"us-east-1", "https://dynamodb.us-east-1.amazonaws.com"}
	USWest1      *Region = &Region{"us-west-1", "https://dynamodb.us-west-1.amazonaws.com"}
	USWest2      *Region = &Region{"us-west-2", "https://dynamodb.us-west-2.amazonaws.com"}
	EUWest1      *Region = &Region{"eu-west-1", "https://dynamodb.eu-west-1.amazonaws.com"}
	APNorthEast1 *Region = &Region{"ap-northeast-1", "https://dynamodb.ap-northeast-1.amazonaws.com"}
	APSouthEast1 *Region = &Region{"ap-southeast-1", "https://dynamodb.ap-southeast-1.amazonaws.com"}
	Local        *Region = &Region{"local", "http://localhost:8000"}
)

type Client struct {
	keys   *Keys
	Region *Region
	log    *Logger
}

func (c *Client) Service() *Service {
	return &Service{Name: "swf", Region: c.Region.Name}
}

func (c *Client) StartWorkflow(request StartWorkflowRequest) (*StartWorkflowResponse, error) {
	resp := &StartWorkflowResponse{}
	err := c.swfReqWithResponse("SStartWorkflowExecution", request, resp)
	return resp, err
}

func (c *Client) SignalWorkflow(request SignalWorkflowRequest) error {
	err := c.swfReqNoResponse("SignalWorkflowExecution", request)
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
	req.Header.Set("X-Amz-Target", "com.amazonaws.swf.service.model.SimpleWorkflowService."+operation)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("Connection", "Keep-Alive")

	err = c.Sign(req)
	if err != nil {
		return err
	}

	if c.log != nil {
		out, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			return err
		}
		c.log.Multiln(string(out))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if c.log != nil {
		out, err := httputil.DumpResponse(resp, true)
		if err != nil {
			return err
		}
		c.log.Multiln(string(out))
	}

	if resp.StatusCode != 200 {
		return errors.New("non 200")
	}

	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(response)
	return err
}

func (c *Client) swfReqNoResponse(operation string, request interface{}) error {
	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(request); err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "", &b)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return errors.New("not 200")
	}

	return nil
}

func (c *Client) Sign(request *http.Request) error {
	return c.Service().Sign(c.keys, request)
}

type Logger log.Logger

func (l *Logger) Multiln(s string) {
	lines := strings.Split(s, "\n")
	for _, line := range lines {
		ln := strings.TrimRightFunc(line, unicode.IsSpace)
		(*log.Logger)(l).Printf("line=%q\n", ln)
	}
}
