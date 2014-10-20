swf4go
======

go client library for amazon simple workflow service

* the `swf.Client` struct implements the below interfaces

```go
package swf

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
```

* create a client like so

```go
client := swf.NewClient(swf.MustGetenv("AWS_ACCESS_KEY_ID"), swf.MustGetenv("AWS_SECRET_ACCESS_KEY"), swf.USEast1)
```
