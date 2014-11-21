package swf

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/bmizerany/aws4"
)

// WorkflowClient is the combined client of the ActivityWorkerClient,
// DecisionWorkerClient, KinesisClient, WorkflowAdminClient,
// WorkflowInfoClient, and WorkflowWorkerClient interfaces.
type WorkflowClient interface {
	ActivityWorkerClient
	DecisionWorkerClient
	KinesisClient
	WorkflowAdminClient
	WorkflowInfoClient
	WorkflowWorkerClient
}

// WorkflowWorkerClient specifies ActivityWorkerClient operations related to starting and stopping workflows.
type WorkflowWorkerClient interface {
	StartWorkflow(request StartWorkflowRequest) (*StartWorkflowResponse, error)
	SignalWorkflow(request SignalWorkflowRequest) error
	RequestCancelWorkflowExecution(request RequestCancelWorkflowExecution) error
	TerminateWorkflowExecution(request TerminateWorkflowExecution) error
}

// DecisionWorkerClient specifies swf client operations related to requesting and responding to decision tasks.
type DecisionWorkerClient interface {
	PollForDecisionTask(request PollForDecisionTaskRequest) (*PollForDecisionTaskResponse, error)
	RespondDecisionTaskCompleted(request RespondDecisionTaskCompletedRequest) error
}

// ActivityWorkerClient specifies swf client operations related to requesting and responding to activity tasks.
type ActivityWorkerClient interface {
	PollForActivityTask(request PollForActivityTaskRequest) (*PollForActivityTaskResponse, error)
	RecordActivityTaskHeartbeat(request RecordActivityTaskHeartbeatRequest) (*RecordActivityTaskHeartbeatResponse, error)
	RespondActivityTaskCompleted(request RespondActivityTaskCompletedRequest) error
	RespondActivityTaskFailed(request RespondActivityTaskFailedRequest) error
	RespondActivityTaskCanceled(request RespondActivityTaskCanceledRequest) error
}

// WorkflowAdminClient specifies swf client operations related to registering and deprecating domains, workflows and activities.
type WorkflowAdminClient interface {
	RegisterActivityType(request RegisterActivityType) error
	DeprecateActivityType(request DeprecateActivityType) error
	RegisterWorkflowType(request RegisterWorkflowType) error
	DeprecateWorkflowType(request DeprecateWorkflowType) error
	RegisterDomain(request RegisterDomain) error
	DeprecateDomain(request DeprecateDomain) error
}

// WorkflowInfoClient specifies swf client operations related to querying swf types and executions.
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

// WorkflowAdminInfoClient is the combined interface of WorkflowAdminClient and WorkflowInfoClient.
type WorkflowAdminInfoClient interface {
	WorkflowAdminClient
	WorkflowInfoClient
}

type KinesisClient interface {
	PutRecord(request PutRecordRequest) (*PutRecordResponse, error)
	CreateStream(request CreateStream) error
	DescribeStream(request DescribeStreamRequest) (*DescribeStreamResponse, error)
}

// Region specifies the AWS region that a client should connect to.
type Region struct {
	Name            string
	SWFEndpoint     string
	KinesisEndpoint string
}

var (
	// USEast1 is the AWS us-east-1 Region
	USEast1 = &Region{"us-east-1", "https://swf.us-east-1.amazonaws.com", "https://kinesis.us-east-1.amazonaws.com"}
	// USWest1 is the AWS us-west-1 Region
	USWest1 = &Region{"us-west-1", "https://swf.us-west-1.amazonaws.com", "https://kinesis.us-west-1.amazonaws.com"}
	// USWest2 is the AWS us-west-2 Region
	USWest2 = &Region{"us-west-2", "https://swf.us-west-2.amazonaws.com", "https://kinesis.us-west-2.amazonaws.com"}
	// EUWest1 is the AWS eu-west-1 Region
	EUWest1 = &Region{"eu-west-1", "https://swf.eu-west-1.amazonaws.com", "https://kinesis.eu-west-1.amazonaws.com"}
	// APNorthEast1 is the AWS ap-northeast-1 Region
	APNorthEast1 = &Region{"ap-northeast-1", "https://swf.ap-northeast-1.amazonaws.com", "https://kinesis.ap-northeast-1.amazonaws.com"}
	// APSouthEast1 is the AWS ap-southeast-1 Region
	APSouthEast1 = &Region{"ap-southeast-1", "https://swf.ap-southeast-1.amazonaws.com", "https://kinesis.ap-southeast-1.amazonaws.com"}
)

// Client is the implementation of the WorkflowClient, DecisionWorkerClient, ActivityWorkerClient, WorkflowAdminClient, and WorkflowInfoClient interfaces.
type Client struct {
	keys       *aws4.Keys
	httpClient *http.Client
	Region     *Region
	Debug      bool
}

// NewClient creates a new Client which uses the given credentials to talk to the given region with http.DefaultClient.
func NewClient(key string, secret string, region *Region) *Client {
	return &Client{
		keys:       &aws4.Keys{AccessKey: key, SecretKey: secret},
		httpClient: http.DefaultClient,
		Region:     region,
	}
}

// NewClientWithHTTPClient creates a new Client which uses the given credentials to talk to the given region with the specified http.Client.
func NewClientWithHTTPClient(key string, secret string, region *Region, client *http.Client) *Client {
	return &Client{
		keys:       &aws4.Keys{AccessKey: key, SecretKey: secret},
		httpClient: client,
		Region:     region,
	}
}

func (c *Client) service(svc string) *aws4.Service {
	return &aws4.Service{Name: svc, Region: c.Region.Name}
}

// StartWorkflow executes http://docs.aws.amazon.com/amazonswf/latest/apireference/API_StartWorkflowExecution.html
func (c *Client) StartWorkflow(request StartWorkflowRequest) (*StartWorkflowResponse, error) {
	resp := &StartWorkflowResponse{}
	err := c.swfReqWithResponse("StartWorkflowExecution", request, resp)
	return resp, err
}

// SignalWorkflow executes http://docs.aws.amazon.com/amazonswf/latest/apireference/API_SignalWorkflowExecution.html
func (c *Client) SignalWorkflow(request SignalWorkflowRequest) error {
	err := c.swfReqNoResponse("SignalWorkflowExecution", request)
	return err
}

// RequestCancelWorkflowExecution executes http://docs.aws.amazon.com/amazonswf/latest/apireference/API_RequestCancelWorkflowExecution.html
func (c *Client) RequestCancelWorkflowExecution(request RequestCancelWorkflowExecution) error {
	err := c.swfReqNoResponse("RequestCancelWorkflowExecution", request)
	return err
}

// TerminateWorkflowExecution executes http://docs.aws.amazon.com/amazonswf/latest/apireference/API_TerminateWorkflowExecution.html
func (c *Client) TerminateWorkflowExecution(request TerminateWorkflowExecution) error {
	err := c.swfReqNoResponse("TerminateWorkflowExecution", request)
	return err
}

// PollForDecisionTask executes http://docs.aws.amazon.com/amazonswf/latest/apireference/API_PollForDecisionTask.html
func (c *Client) PollForDecisionTask(request PollForDecisionTaskRequest) (*PollForDecisionTaskResponse, error) {
	resp := &PollForDecisionTaskResponse{}
	err := c.swfReqWithResponse("PollForDecisionTask", request, resp)
	return resp, err
}

// RespondDecisionTaskCompleted executes http://docs.aws.amazon.com/amazonswf/latest/apireference/API_RespondDecisionTaskCompleted.html
func (c *Client) RespondDecisionTaskCompleted(request RespondDecisionTaskCompletedRequest) error {
	err := c.swfReqNoResponse("RespondDecisionTaskCompleted", request)
	return err
}

// PollForActivityTask executes http://docs.aws.amazon.com/amazonswf/latest/apireference/API_PollForActivityTask.html
func (c *Client) PollForActivityTask(request PollForActivityTaskRequest) (*PollForActivityTaskResponse, error) {
	resp := &PollForActivityTaskResponse{}
	err := c.swfReqWithResponse("PollForActivityTask", request, resp)
	return resp, err
}

// RespondActivityTaskCompleted executes http://docs.aws.amazon.com/amazonswf/latest/apireference/API_RespondActivityTaskCompleted.html
func (c *Client) RespondActivityTaskCompleted(request RespondActivityTaskCompletedRequest) error {
	err := c.swfReqNoResponse("RespondActivityTaskCompleted", request)
	return err
}

// RespondActivityTaskFailed executes http://docs.aws.amazon.com/amazonswf/latest/apireference/API_RespondActivityTaskFailed.html
func (c *Client) RespondActivityTaskFailed(request RespondActivityTaskFailedRequest) error {
	err := c.swfReqNoResponse("RespondActivityTaskFailed", request)
	return err
}

// RespondActivityTaskCanceled executes http://docs.aws.amazon.com/amazonswf/latest/apireference/API_RespondActivityTaskCanceled.html
func (c *Client) RespondActivityTaskCanceled(request RespondActivityTaskCanceledRequest) error {
	err := c.swfReqNoResponse("RespondActivityTaskCanceled", request)
	return err
}

// RecordActivityTaskHeartbeat executes http://docs.aws.amazon.com/amazonswf/latest/apireference/API_RecordActivityTaskHeartbeat.html
func (c *Client) RecordActivityTaskHeartbeat(request RecordActivityTaskHeartbeatRequest) (*RecordActivityTaskHeartbeatResponse, error) {
	resp := &RecordActivityTaskHeartbeatResponse{}
	err := c.swfReqWithResponse("RecordActivityTaskHeartbeat", request, resp)
	return resp, err
}

// RegisterActivityType executes http://docs.aws.amazon.com/amazonswf/latest/apireference/API_RegisterActivityType.html
func (c *Client) RegisterActivityType(request RegisterActivityType) error {
	err := c.swfReqNoResponse("RegisterActivityType", request)
	return err
}

// DeprecateActivityType executes http://docs.aws.amazon.com/amazonswf/latest/apireference/API_DeprecateActivityType.html
func (c *Client) DeprecateActivityType(request DeprecateActivityType) error {
	err := c.swfReqNoResponse("DeprecateActivityType", request)
	return err
}

// RegisterWorkflowType executes http://docs.aws.amazon.com/amazonswf/latest/apireference/API_RegisterWorkflowType.html
func (c *Client) RegisterWorkflowType(request RegisterWorkflowType) error {
	err := c.swfReqNoResponse("RegisterWorkflowType", request)
	return err
}

// DeprecateWorkflowType executes http://docs.aws.amazon.com/amazonswf/latest/apireference/API_DeprecateWorkflowType.html
func (c *Client) DeprecateWorkflowType(request DeprecateWorkflowType) error {
	err := c.swfReqNoResponse("DeprecateWorkflowType", request)
	return err
}

// RegisterDomain executes http://docs.aws.amazon.com/amazonswf/latest/apireference/API_RegisterDomain.html
func (c *Client) RegisterDomain(request RegisterDomain) error {
	err := c.swfReqNoResponse("RegisterDomain", request)
	return err
}

// DeprecateDomain executes http://docs.aws.amazon.com/amazonswf/latest/apireference/API_DeprecateDomain.html
func (c *Client) DeprecateDomain(request DeprecateDomain) error {
	err := c.swfReqNoResponse("DeprecateDomain", request)
	return err
}

// CountClosedWorkflowExecutions executes http://docs.aws.amazon.com/amazonswf/latest/apireference/API_CountClosedWorkflowExecutions.html
func (c *Client) CountClosedWorkflowExecutions(request CountClosedWorkflowExecutionsRequest) (*CountResponse, error) {
	resp := &CountResponse{}
	err := c.swfReqWithResponse("CountClosedWorkflowExecutions", request, resp)
	return resp, err
}

// CountOpenWorkflowExecutions executes http://docs.aws.amazon.com/amazonswf/latest/apireference/API_CountOpenWorkflowExecutions.html
func (c *Client) CountOpenWorkflowExecutions(request CountOpenWorkflowExecutionsRequest) (*CountResponse, error) {
	resp := &CountResponse{}
	err := c.swfReqWithResponse("CountOpenWorkflowExecutions", request, resp)
	return resp, err
}

// CountPendingActivityTasks executes http://docs.aws.amazon.com/amazonswf/latest/apireference/API_CountPendingActivityTasks.html
func (c *Client) CountPendingActivityTasks(request CountPendingActivityTasksRequest) (*CountResponse, error) {
	resp := &CountResponse{}
	err := c.swfReqWithResponse("CountPendingActivityTasks", request, resp)
	return resp, err
}

// CountPendingDecisionTasks executes http://docs.aws.amazon.com/amazonswf/latest/apireference/API_CountPendingDecisionTasks.html
func (c *Client) CountPendingDecisionTasks(request CountPendingDecisionTasksRequest) (*CountResponse, error) {
	resp := &CountResponse{}
	err := c.swfReqWithResponse("CountPendingDecisionTasks", request, resp)
	return resp, err
}

// DescribeActivityType executes http://docs.aws.amazon.com/amazonswf/latest/apireference/API_DescribeActivityType.html
func (c *Client) DescribeActivityType(request DescribeActivityTypeRequest) (*DescribeActivityTypeResponse, error) {
	resp := &DescribeActivityTypeResponse{}
	err := c.swfReqWithResponse("DescribeActivityType", request, resp)
	return resp, err
}

// DescribeDomain executes http://docs.aws.amazon.com/amazonswf/latest/apireference/API_DescribeDomain.html
func (c *Client) DescribeDomain(request DescribeDomainRequest) (*DescribeDomainResponse, error) {
	resp := &DescribeDomainResponse{}
	err := c.swfReqWithResponse("DescribeDomain", request, resp)
	return resp, err
}

// DescribeWorkflowType executes http://docs.aws.amazon.com/amazonswf/latest/apireference/API_DescribeWorkflowType.html
func (c *Client) DescribeWorkflowType(request DescribeWorkflowTypeRequest) (*DescribeWorkflowTypeResponse, error) {
	resp := &DescribeWorkflowTypeResponse{}
	err := c.swfReqWithResponse("DescribeWorkflowType", request, resp)
	return resp, err
}

// DescribeWorkflowExecution executes http://docs.aws.amazon.com/amazonswf/latest/apireference/API_DescribeWorkflowExecution.html
func (c *Client) DescribeWorkflowExecution(request DescribeWorkflowExecutionRequest) (*DescribeWorkflowExecutionResponse, error) {
	resp := &DescribeWorkflowExecutionResponse{}
	err := c.swfReqWithResponse("DescribeWorkflowExecution", request, resp)
	return resp, err
}

// ListActivityTypes executes http://docs.aws.amazon.com/amazonswf/latest/apireference/API_ListActivityTypes.html
func (c *Client) ListActivityTypes(request ListActivityTypesRequest) (*ListActivityTypesResponse, error) {
	resp := &ListActivityTypesResponse{}
	err := c.swfReqWithResponse("ListActivityTypes", request, resp)
	return resp, err
}

// ListWorkflowTypes executes http://docs.aws.amazon.com/amazonswf/latest/apireference/API_ListWorkflowTypes.html
func (c *Client) ListWorkflowTypes(request ListWorkflowTypesRequest) (*ListWorkflowTypesResponse, error) {
	resp := &ListWorkflowTypesResponse{}
	err := c.swfReqWithResponse("ListWorkflowTypes", request, resp)
	return resp, err
}

// ListOpenWorkflowExecutions executes http://docs.aws.amazon.com/amazonswf/latest/apireference/API_ListOpenWorkflowExecutions.html
func (c *Client) ListOpenWorkflowExecutions(request ListOpenWorkflowExecutionsRequest) (*ListOpenWorkflowExecutionsResponse, error) {
	resp := &ListOpenWorkflowExecutionsResponse{}
	err := c.swfReqWithResponse("ListOpenWorkflowExecutions", request, resp)
	return resp, err
}

// ListClosedWorkflowExecutions executes http://docs.aws.amazon.com/amazonswf/latest/apireference/API_ListClosedWorkflowExecutions.html
func (c *Client) ListClosedWorkflowExecutions(request ListClosedWorkflowExecutionsRequest) (*ListClosedWorkflowExecutionsResponse, error) {
	resp := &ListClosedWorkflowExecutionsResponse{}
	err := c.swfReqWithResponse("ListClosedWorkflowExecutions", request, resp)
	return resp, err
}

// GetWorkflowExecutionHistory executes http://docs.aws.amazon.com/amazonswf/latest/apireference/API_GetWorkflowExecutionHistory.html
func (c *Client) GetWorkflowExecutionHistory(request GetWorkflowExecutionHistoryRequest) (*GetWorkflowExecutionHistoryResponse, error) {
	resp := &GetWorkflowExecutionHistoryResponse{}
	err := c.swfReqWithResponse("GetWorkflowExecutionHistory", request, resp)
	return resp, err
}

// ListDomains executes http://docs.aws.amazon.com/amazonswf/latest/apireference/API_ListDomains.html
func (c *Client) ListDomains(request ListDomainsRequest) (*ListDomainsResponse, error) {
	resp := &ListDomainsResponse{}
	err := c.swfReqWithResponse("ListDomains", request, resp)
	return resp, err
}

func (c *Client) PutRecord(request PutRecordRequest) (*PutRecordResponse, error) {
	resp := &PutRecordResponse{}
	err := c.kinesisReqWithResponse("PutRecord", request, resp)
	return resp, err
}

func (c *Client) CreateStream(request CreateStream) error {
	err := c.kinesisReqNoResponse("CreateStream", request)
	return err
}

func (c *Client) DescribeStream(request DescribeStreamRequest) (*DescribeStreamResponse, error) {
	resp := &DescribeStreamResponse{}
	err := c.kinesisReqWithResponse("DescribeStream", request, resp)
	return resp, err
}

func (c *Client) swfReqNoResponse(operation string, request interface{}) error {
	resp, err := c.prepareAndExecuteRequest("swf", c.Region.SWFEndpoint, "SimpleWorkflowService."+operation, "application/x-amz-json-1.0", request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		response := new(ErrorResponse)
		if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
			return err
		}
		return response
	}
	return err
}

func (c *Client) swfReqWithResponse(operation string, request interface{}, response interface{}) error {
	resp, err := c.prepareAndExecuteRequest("swf", c.Region.SWFEndpoint, "SimpleWorkflowService."+operation, "application/x-amz-json-1.0", request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		errResp := new(ErrorResponse)
		if err := json.NewDecoder(resp.Body).Decode(errResp); err != nil {
			return err
		}
		return errResp
	}

	err = json.NewDecoder(resp.Body).Decode(response)
	if c.Debug {
		pretty, _ := json.MarshalIndent(response, "", "    ")
		log.Println(string(pretty))
	}
	return err
}

func (c *Client) kinesisReqNoResponse(operation string, request interface{}) error {
	resp, err := c.prepareAndExecuteRequest("kinesis", c.Region.KinesisEndpoint, "Kinesis_20131202."+operation, "application/x-amz-json-1.1", request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		errResp := new(ErrorResponse)
		if err := json.NewDecoder(resp.Body).Decode(errResp); err != nil {
			return err
		}
		return errResp
	}
	return err
}

func (c *Client) kinesisReqWithResponse(operation string, request interface{}, response interface{}) error {
	resp, err := c.prepareAndExecuteRequest("kinesis", c.Region.KinesisEndpoint, "Kinesis_20131202."+operation, "application/x-amz-json-1.1", request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		errResp := new(ErrorResponse)
		if err := json.NewDecoder(resp.Body).Decode(errResp); err != nil {
			return err
		}
		return errResp
	}

	err = json.NewDecoder(resp.Body).Decode(response)
	if c.Debug {
		pretty, _ := json.MarshalIndent(response, "", "    ")
		log.Println(string(pretty))
	}
	return err
}

func (c *Client) prepareAndExecuteRequest(service string, url string, target string, contentType string, request interface{}) (*http.Response, error) {
	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(request); err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Date", time.Now().UTC().Format(http.TimeFormat))
	req.Header.Set("X-Amz-Target", target)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Connection", "Keep-Alive")

	err = c.sign(service, req)
	if err != nil {
		return nil, err
	}

	if c.Debug {
		out, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			return nil, err
		}
		multiln(string(out))
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if c.Debug {
		defer resp.Body.Close()
		out, err := httputil.DumpResponse(resp, true)
		if err != nil {
			return nil, err
		}
		multiln(string(out))
	}

	return resp, nil
}

func (c *Client) sign(svc string, request *http.Request) error {
	return c.service(svc).Sign(c.keys, request)
}

func multiln(s string) {
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
