swf4go
======

go client library for amazon simple workflow service

* the Client struct implements the below interfaces

```go
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

type WorkflowAdminClient interface {
	RegisterActivityType(request RegisterActivityType) error
	DeprecateActivityType(request DeprecateActivityType) error
	RegisterWorkflowType(request RegisterWorkflowType) error
	DeprecateWorkflowType(request DeprecateWorkflowType) error
	RegisterDomain(request RegisterDomain) error
	DeprecateDomain(request DeprecateDomain) error
}
```

* create a client like so

```go
client := NewClient(MustGetenv("AWS_ACCESS_KEY_ID"), MustGetenv("AWS_SECRET_ACCESS_KEY"), USEast1)
```

need a request/response struct?
-------------------------------

* check out json/readme.md
* put a file called json/NameOfActionRequest.json and json/NameOfActionResponse.json
* run jsongen json/NameOfActionRequest.json
* add the struct, after any cleanups/deduping to protocol.go
* add a method to the right client interface and an implementation in the client