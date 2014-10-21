package swf

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"time"
	"unicode"
)

type WorkflowClient interface {
	StartWorkflow(request StartWorkflowRequest) (*StartWorkflowResponse, error)
	SignalWorkflow(request SignalWorkflowRequest) error
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

type WorkflowAdminClient interface {
	RegisterActivityType(request RegisterActivityType) error
	DeprecateActivityType(request DeprecateActivityType) error
	RegisterWorkflowType(request RegisterWorkflowType) error
	DeprecateWorkflowType(request DeprecateWorkflowType) error
	RegisterDomain(request RegisterDomain) error
	DeprecateDomain(request DeprecateDomain) error
}

type WorkflowInfoClient interface {
	CountClosedWorkflowExecutions(request CountClosedWorkflowExecutionsRequest) (*CountResponse, error)
	CountOpenWorkflowExecutions(request CountOpenWorkflowExecutionsRequest) (*CountResponse, error)
	CountPendingActivityTasks(request CountPendingActivityTasksRequest) (*CountResponse, error)
	CountPendingDecisionTasks(request CountPendingDecisionTasksRequest) (*CountResponse, error)
	ListWorkflowTypes(request ListWorkflowTypesRequest) (*ListWorkflowTypesResponse, error)
	ListActivityTypes(request ListActivityTypesRequest) (*ListActivityTypesResponse, error)
	DescribeActivityType(request DescribeActivityTypeRequest) (*DescribeActivityTypeResponse, error)
	DescribeDomain(request DescribeDomainRequest) (*DescribeDomainResponse, error)
	DescribeWorkflowType(request DescribeWorkflowTypeRequest) (*DescribeWorkflowTypeResponse, error)
	DescribeWorkflowExecution(request DescribeWorkflowExecutionRequest) (*DescribeWorkflowExecutionResponse, error)
	ListOpenWorkflowExecutions(request ListOpenWorkflowExecutionsRequest) (*ListOpenWorkflowExecutionsResponse, error)
	ListClosedWorkflowExecutions(request ListClosedWorkflowExecutionsRequest) (*ListClosedWorkflowExecutionsResponse, error)
	GetWorkflowExecutionHistory(request GetWorkflowExecutionHistoryRequest) (*GetWorkflowExecutionHistoryResponse, error)
	ListDomains(request ListDomainsRequest) (*ListDomainsResponse, error)
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
		keys:   &Keys{AccessKey: key, SecretKey: secret},
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

func (c *Client) RegisterActivityType(request RegisterActivityType) error {
	err := c.swfReqNoResponse("RegisterActivityType", request)
	return err
}
func (c *Client) DeprecateActivityType(request DeprecateActivityType) error {
	err := c.swfReqNoResponse("DeprecateActivityType", request)
	return err
}
func (c *Client) RegisterWorkflowType(request RegisterWorkflowType) error {
	err := c.swfReqNoResponse("RegisterWorkflowType", request)
	return err
}
func (c *Client) DeprecateWorkflowType(request DeprecateWorkflowType) error {
	err := c.swfReqNoResponse("DeprecateWorkflowType", request)
	return err
}
func (c *Client) RegisterDomain(request RegisterDomain) error {
	err := c.swfReqNoResponse("RegisterDomain", request)
	return err
}
func (c *Client) DeprecateDomain(request DeprecateDomain) error {
	err := c.swfReqNoResponse("DeprecateDomain", request)
	return err
}

func (c *Client) CountClosedWorkflowExecutions(request CountClosedWorkflowExecutionsRequest) (*CountResponse, error) {
	resp := &CountResponse{}
	err := c.swfReqWithResponse("CountClosedWorkflowExecutions", request, resp)
	return resp, err
}
func (c *Client) CountOpenWorkflowExecutions(request CountOpenWorkflowExecutionsRequest) (*CountResponse, error) {
	resp := &CountResponse{}
	err := c.swfReqWithResponse("CountOpenWorkflowExecutions", request, resp)
	return resp, err
}
func (c *Client) CountPendingActivityTasks(request CountPendingActivityTasksRequest) (*CountResponse, error) {
	resp := &CountResponse{}
	err := c.swfReqWithResponse("CountPendingActivityTasks", request, resp)
	return resp, err
}
func (c *Client) CountPendingDecisionTasks(request CountPendingDecisionTasksRequest) (*CountResponse, error) {
	resp := &CountResponse{}
	err := c.swfReqWithResponse("CountPendingDecisionTasks", request, resp)
	return resp, err
}

func (c *Client) DescribeActivityType(request DescribeActivityTypeRequest) (*DescribeActivityTypeResponse, error) {
	resp := &DescribeActivityTypeResponse{}
	err := c.swfReqWithResponse("DescribeActivityType", request, resp)
	return resp, err
}
func (c *Client) DescribeDomain(request DescribeDomainRequest) (*DescribeDomainResponse, error) {
	resp := &DescribeDomainResponse{}
	err := c.swfReqWithResponse("DescribeDomain", request, resp)
	return resp, err
}
func (c *Client) DescribeWorkflowType(request DescribeWorkflowTypeRequest) (*DescribeWorkflowTypeResponse, error) {
	resp := &DescribeWorkflowTypeResponse{}
	err := c.swfReqWithResponse("DescribeWorkflowType", request, resp)
	return resp, err
}

func (c *Client) DescribeWorkflowExecution(request DescribeWorkflowExecutionRequest) (*DescribeWorkflowExecutionResponse, error) {
	resp := &DescribeWorkflowExecutionResponse{}
	err := c.swfReqWithResponse("DescribeWorkflowExecution", request, resp)
	return resp, err
}

func (c *Client) ListWorkflowTypes(request ListWorkflowTypesRequest) (*ListWorkflowTypesResponse, error) {
	resp := &ListWorkflowTypesResponse{}
	err := c.swfReqWithResponse("ListWorkflowTypes", request, resp)
	return resp, err
}

func (c *Client) ListOpenWorkflowExecutions(request ListOpenWorkflowExecutionsRequest) (*ListOpenWorkflowExecutionsResponse, error) {
	resp := &ListOpenWorkflowExecutionsResponse{}
	err := c.swfReqWithResponse("ListOpenWorkflowExecutions", request, resp)
	return resp, err
}

func (c *Client) ListClosedWorkflowExecutions(request ListClosedWorkflowExecutionsRequest) (*ListClosedWorkflowExecutionsResponse, error) {
	resp := &ListClosedWorkflowExecutionsResponse{}
	err := c.swfReqWithResponse("ListClosedWorkflowExecutions", request, resp)
	return resp, err
}

func (c *Client) GetWorkflowExecutionHistory(request GetWorkflowExecutionHistoryRequest) (*GetWorkflowExecutionHistoryResponse, error) {
	resp := &GetWorkflowExecutionHistoryResponse{}
	err := c.swfReqWithResponse("GetWorkflowExecutionHistory", request, resp)
	return resp, err
}

func (c *Client) ListDomains(request ListDomainsRequest) (*ListDomainsResponse, error) {
	resp := &ListDomainsResponse{}
	err := c.swfReqWithResponse("ListDomains", request, resp)
	return resp, err
}

func (c *Client) swfReqWithResponse(operation string, request interface{}, response interface{}) error {
	resp, err := c.prepareAndExecuteRequest(operation, request)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return errors.New("non 200")
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(response)
	pretty, _ := json.MarshalIndent(response, "", "    ")
	log.Println(string(pretty))
	return err
}

func (c *Client) swfReqNoResponse(operation string, request interface{}) error {
	resp, err := c.prepareAndExecuteRequest(operation, request)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return errors.New("non 200")
	}
	return err
}

func (c *Client) prepareAndExecuteRequest(operation string, request interface{}) (*http.Response, error) {
	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(request); err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", c.Region.URL(), &b)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Date", time.Now().UTC().Format(http.TimeFormat))
	req.Header.Set("X-Amz-Target", "SimpleWorkflowService."+operation)
	req.Header.Set("Content-Type", "application/x-amz-json-1.0")
	req.Header.Set("Connection", "Keep-Alive")

	err = c.Sign(req)
	if err != nil {
		return nil, err
	}

	out, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		return nil, err
	}
	Multiln(string(out))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	out, err = httputil.DumpResponse(resp, true)
	if err != nil {
		return nil, err
	}
	Multiln(string(out))

	return resp, nil
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
