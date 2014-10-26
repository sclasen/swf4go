package swf

type ErrorResponse struct {
	StatusCode int
	Type       string `json:"__type"`
	Message    string `json:"message"`
}

func (err *ErrorResponse) Error() string {
	return err.Type + ": " + err.Message
}

/*WorkflowProtocol*/
type StartWorkflowRequest struct {
	ChildPolicy                  string       `json:"childPolicy,omitempty"`
	Domain                       string       `json:"domain"`
	ExecutionStartToCloseTimeout string       `json:"executionStartToCloseTimeout,omitempty"`
	Input                        string       `json:"input,omitempty"`
	TagList                      []string     `json:"tagList,omitempty"`
	TaskList                     *TaskList    `json:"taskList,omitempty"`
	TaskStartToCloseTimeout      string       `json:"taskStartToCloseTimeout,omitempty"`
	WorkflowId                   string       `json:"workflowId"`
	WorkflowType                 WorkflowType `json:"workflowType"`
}

type StartWorkflowResponse struct {
	RunId string `json:"runId"`
}

type RequestCancelWorkflowExecution struct {
	Domain     string `json:"domain"`
	RunId      string `json:"runId,omitempty"`
	WorkflowId string `json:"workflowId"`
}

type SignalWorkflowRequest struct {
	Domain     string `json:"domain"`
	Input      string `json:"input,omitempty"`
	RunId      string `json:"runId,omitempty"`
	SignalName string `json:"signalName"`
	WorkflowId string `json:"workflowId"`
}

type ListWorkflowTypesRequest struct {
	Domain             string `json:"domain"`
	MaximumPageSize    int    `json:"maximumPageSize,omitempty"`
	Name               string `json:"name,omitempty"`
	NextPageToken      string `json:"nextPageToken,omitempty"`
	RegistrationStatus string `json:"registrationStatus"`
	ReverseOrder       bool   `json:"reverseOrder,omitempty"`
}

type ListWorkflowTypesResponse struct {
	NextPageToken string             `json:"nextPageToken,omitempty"`
	TypeInfos     []WorkflowTypeInfo `json:"typeInfos"`
}

type TerminateWorkflowExecution struct {
	ChildPolicy string `json:"childPolicy,omitempty"`
	Details     string `json:"details,omitempty"`
	Domain      string `json:"domain"`
	Reason      string `json:"reason,omitempty"`
	RunId       string `json:"runId,omitempty"`
	WorkflowId  string `json:"workflowId,omitempty"`
}

/*DecisionWorkerProtocol*/

type PollForDecisionTaskRequest struct {
	Domain          string   `json:"domain"`
	Identity        string   `json:"identity,omitempty"`
	MaximumPageSize int      `json:"maximumPageSize,omitempty"`
	NextPageToken   string   `json:"nextPageToken,omitempty"`
	ReverseOrder    bool     `json:"reverseOrder,omitempty"`
	TaskList        TaskList `json:"taskList"`
}

type PollForDecisionTaskResponse struct {
	Events                 []HistoryEvent    `json:"events"`
	NextPageToken          string            `json:"nextPageToken"`
	PreviousStartedEventId int               `json:"previousStartedEventId"`
	StartedEventId         int               `json:"startedEventId"`
	TaskToken              string            `json:"taskToken"`
	WorkflowExecution      WorkflowExecution `json:"workflowExecution"`
	WorkflowType           WorkflowType      `json:"workflowType"`
}

// EventType := WorkflowExecutionStarted | WorkflowExecutionCancelRequested | WorkflowExecutionCompleted | CompleteWorkflowExecutionFailed | WorkflowExecutionFailed | FailWorkflowExecutionFailed |
// WorkflowExecutionTimedOut | WorkflowExecutionCanceled | CancelWorkflowExecutionFailed | WorkflowExecutionContinuedAsNew | ContinueAsNewWorkflowExecutionFailed | WorkflowExecutionTerminated |
// DecisionTaskScheduled | DecisionTaskStarted | DecisionTaskCompleted | DecisionTaskTimedOut | ActivityTaskScheduled | ScheduleActivityTaskFailed | ActivityTaskStarted | ActivityTaskCompleted |
// ActivityTaskFailed | ActivityTaskTimedOut | ActivityTaskCanceled | ActivityTaskCancelRequested | RequestCancelActivityTaskFailed | WorkflowExecutionSignaled | MarkerRecorded | RecordMarkerFailed |
// TimerStarted | StartTimerFailed | TimerFired | TimerCanceled | CancelTimerFailed | StartChildWorkflowExecutionInitiated | StartChildWorkflowExecutionFailed | ChildWorkflowExecutionStarted |
// ChildWorkflowExecutionCompleted | ChildWorkflowExecutionFailed | ChildWorkflowExecutionTimedOut | ChildWorkflowExecutionCanceled | ChildWorkflowExecutionTerminated | SignalExternalWorkflowExecutionInitiated |
// SignalExternalWorkflowExecutionFailed | ExternalWorkflowExecutionSignaled | RequestCancelExternalWorkflowExecutionInitiated | RequestCancelExternalWorkflowExecutionFailed | ExternalWorkflowExecutionCancelRequested

// Valid values for the field EventType in the HistoryEvent struct.
const (
	EventTypeWorkflowExecutionStarted                        = "WorkflowExecutionStarted"
	EventTypeWorkflowExecutionCancelRequested                = "WorkflowExecutionCancelRequested"
	EventTypeWorkflowExecutionCompleted                      = "WorkflowExecutionCompleted"
	EventTypeCompleteWorkflowExecutionFailed                 = "CompleteWorkflowExecutionFailed"
	EventTypeWorkflowExecutionFailed                         = "WorkflowExecutionFailed"
	EventTypeFailWorkflowExecutionFailed                     = "FailWorkflowExecutionFailed"
	EventTypeWorkflowExecutionTimedOut                       = "WorkflowExecutionTimedOut"
	EventTypeWorkflowExecutionCanceled                       = "WorkflowExecutionCanceled"
	EventTypeCancelWorkflowExecutionFailed                   = "CancelWorkflowExecutionFailed"
	EventTypeWorkflowExecutionContinuedAsNew                 = "WorkflowExecutionContinuedAsNew"
	EventTypeContinueAsNewWorkflowExecutionFailed            = "ContinueAsNewWorkflowExecutionFailed"
	EventTypeWorkflowExecutionTerminated                     = "WorkflowExecutionTerminated"
	EventTypeDecisionTaskScheduled                           = "DecisionTaskScheduled"
	EventTypeDecisionTaskStarted                             = "DecisionTaskStarted"
	EventTypeDecisionTaskCompleted                           = "DecisionTaskCompleted"
	EventTypeDecisionTaskTimedOut                            = "DecisionTaskTimedOut"
	EventTypeActivityTaskScheduled                           = "ActivityTaskScheduled"
	EventTypeScheduleActivityTaskFailed                      = "ScheduleActivityTaskFailed"
	EventTypeActivityTaskStarted                             = "ActivityTaskStarted"
	EventTypeActivityTaskCompleted                           = "ActivityTaskCompleted"
	EventTypeActivityTaskFailed                              = "ActivityTaskFailed"
	EventTypeActivityTaskTimedOut                            = "ActivityTaskTimedOut"
	EventTypeActivityTaskCanceled                            = "ActivityTaskCanceled"
	EventTypeActivityTaskCancelRequested                     = "ActivityTaskCancelRequested"
	EventTypeRequestCancelActivityTaskFailed                 = "RequestCancelActivityTaskFailed"
	EventTypeWorkflowExecutionSignaled                       = "WorkflowExecutionSignaled"
	EventTypeMarkerRecorded                                  = "MarkerRecorded"
	EventTypeRecordMarkerFailed                              = "RecordMarkerFailed"
	EventTypeTimerStarted                                    = "TimerStarted"
	EventTypeStartTimerFailed                                = "StartTimerFailed"
	EventTypeTimerFired                                      = "TimerFired"
	EventTypeTimerCanceled                                   = "TimerCanceled"
	EventTypeCancelTimerFailed                               = "CancelTimerFailed"
	EventTypeStartChildWorkflowExecutionInitiated            = "StartChildWorkflowExecutionInitiated"
	EventTypeStartChildWorkflowExecutionFailed               = "StartChildWorkflowExecutionFailed"
	EventTypeChildWorkflowExecutionStarted                   = "ChildWorkflowExecutionStarted"
	EventTypeChildWorkflowExecutionCompleted                 = "ChildWorkflowExecutionCompleted"
	EventTypeChildWorkflowExecutionFailed                    = "ChildWorkflowExecutionFailed"
	EventTypeChildWorkflowExecutionTimedOut                  = "ChildWorkflowExecutionTimedOut"
	EventTypeChildWorkflowExecutionCanceled                  = "ChildWorkflowExecutionCanceled"
	EventTypeChildWorkflowExecutionTerminated                = "ChildWorkflowExecutionTerminated"
	EventTypeSignalExternalWorkflowExecutionInitiated        = "SignalExternalWorkflowExecutionInitiated"
	EventTypeSignalExternalWorkflowExecutionFailed           = "SignalExternalWorkflowExecutionFailed"
	EventTypeExternalWorkflowExecutionSignaled               = "ExternalWorkflowExecutionSignaled"
	EventTypeRequestCancelExternalWorkflowExecutionInitiated = "RequestCancelExternalWorkflowExecutionInitiated"
	EventTypeRequestCancelExternalWorkflowExecutionFailed    = "RequestCancelExternalWorkflowExecutionFailed"
	EventTypeExternalWorkflowExecutionCancelRequested        = "ExternalWorkflowExecutionCancelRequested"
)

type HistoryEvent struct {
	ActivityTaskCancelRequestedEventAttributes                     *ActivityTaskCancelRequestedEventAttributes                     `json:"activityTaskCancelRequestedEventAttributes,omitempty"`
	ActivityTaskCanceledEventAttributes                            *ActivityTaskCanceledEventAttributes                            `json:"activityTaskCanceledEventAttributes,omitempty"`
	ActivityTaskCompletedEventAttributes                           *ActivityTaskCompletedEventAttributes                           `json:"activityTaskCompletedEventAttributes,omitempty"`
	ActivityTaskFailedEventAttributes                              *ActivityTaskFailedEventAttributes                              `json:"activityTaskFailedEventAttributes,omitempty"`
	ActivityTaskScheduledEventAttributes                           *ActivityTaskScheduledEventAttributes                           `json:"activityTaskScheduledEventAttributes,omitempty"`
	ActivityTaskStartedEventAttributes                             *ActivityTaskStartedEventAttributes                             `json:"activityTaskStartedEventAttributes,omitempty"`
	ActivityTaskTimedOutEventAttributes                            *ActivityTaskTimedOutEventAttributes                            `json:"activityTaskTimedOutEventAttributes,omitempty"`
	CancelTimerFailedEventAttributes                               *CancelTimerFailedEventAttributes                               `json:"cancelTimerFailedEventAttributes,omitempty"`
	CancelWorkflowExecutionFailedEventAttributes                   *CancelWorkflowExecutionFailedEventAttributes                   `json:"cancelWorkflowExecutionFailedEventAttributes,omitempty"`
	ChildWorkflowExecutionCanceledEventAttributes                  *ChildWorkflowExecutionCanceledEventAttributes                  `json:"childWorkflowExecutionCanceledEventAttributes,omitempty"`
	ChildWorkflowExecutionCompletedEventAttributes                 *ChildWorkflowExecutionCompletedEventAttributes                 `json:"childWorkflowExecutionCompletedEventAttributes,omitempty"`
	ChildWorkflowExecutionFailedEventAttributes                    *CancelWorkflowExecutionFailedEventAttributes                   `json:"childWorkflowExecutionFailedEventAttributes,omitempty"`
	ChildWorkflowExecutionStartedEventAttributes                   *ChildWorkflowExecutionStartedEventAttributes                   `json:"childWorkflowExecutionStartedEventAttributes,omitempty"`
	ChildWorkflowExecutionTerminatedEventAttributes                *ChildWorkflowExecutionTerminatedEventAttributes                `json:"childWorkflowExecutionTerminatedEventAttributes,omitempty"`
	ChildWorkflowExecutionTimedOutEventAttributes                  *ChildWorkflowExecutionTimedOutEventAttributes                  `json:"childWorkflowExecutionTimedOutEventAttributes,omitempty"`
	CompleteWorkflowExecutionFailedEventAttributes                 *CompleteWorkflowExecutionFailedEventAttributes                 `json:"completeWorkflowExecutionFailedEventAttributes,omitempty"`
	ContinueAsNewWorkflowExecutionFailedEventAttributes            *ContinueAsNewWorkflowExecutionFailedEventAttributes            `json:"continueAsNewWorkflowExecutionFailedEventAttributes,omitempty"`
	DecisionTaskCompletedEventAttributes                           *DecisionTaskCompletedEventAttributes                           `json:"decisionTaskCompletedEventAttributes,omitempty"`
	DecisionTaskScheduledEventAttributes                           *DecisionTaskScheduledEventAttributes                           `json:"decisionTaskScheduledEventAttributes,omitempty"`
	DecisionTaskStartedEventAttributes                             *DecisionTaskStartedEventAttributes                             `json:"decisionTaskStartedEventAttributes,omitempty"`
	DecisionTaskTimedOutEventAttributes                            *DecisionTaskTimedOutEventAttributes                            `json:"decisionTaskTimedOutEventAttributes,omitempty"`
	EventId                                                        int                                                             `json:"eventId"`
	EventTimestamp                                                 float32                                                         `json:"eventTimestamp"`
	EventType                                                      string                                                          `json:"eventType"`
	ExternalWorkflowExecutionCancelRequestedEventAttributes        *ExternalWorkflowExecutionCancelRequestedEventAttributes        `json:"externalWorkflowExecutionCancelRequestedEventAttributes,omitempty"`
	ExternalWorkflowExecutionSignaledEventAttributes               *ExternalWorkflowExecutionSignaledEventAttributes               `json:"externalWorkflowExecutionSignaledEventAttributes,omitempty"`
	FailWorkflowExecutionFailedEventAttributes                     *FailWorkflowExecutionFailedEventAttributes                     `json:"failWorkflowExecutionFailedEventAttributes,omitempty"`
	MarkerRecordedEventAttributes                                  *MarkerRecordedEventAttributes                                  `json:"markerRecordedEventAttributes,omitempty"`
	RecordMarkerFailedEventAttributes                              *RecordMarkerFailedEventAttributes                              `json:"recordMarkerFailedEventAttributes,omitempty"`
	RequestCancelActivityTaskFailedEventAttributes                 *RequestCancelActivityTaskFailedEventAttributes                 `json:"requestCancelActivityTaskFailedEventAttributes,omitempty"`
	RequestCancelExternalWorkflowExecutionFailedEventAttributes    *RequestCancelExternalWorkflowExecutionFailedEventAttributes    `json:"requestCancelExternalWorkflowExecutionFailedEventAttributes,omitempty"`
	RequestCancelExternalWorkflowExecutionInitiatedEventAttributes *RequestCancelExternalWorkflowExecutionInitiatedEventAttributes `json:"requestCancelExternalWorkflowExecutionInitiatedEventAttributes,omitempty"`
	ScheduleActivityTaskFailedEventAttributes                      *ScheduleActivityTaskFailedEventAttributes                      `json:"scheduleActivityTaskFailedEventAttributes,omitempty"`
	SignalExternalWorkflowExecutionFailedEventAttributes           *SignalExternalWorkflowExecutionFailedEventAttributes           `json:"signalExternalWorkflowExecutionFailedEventAttributes,omitempty"`
	SignalExternalWorkflowExecutionInitiatedEventAttributes        *SignalExternalWorkflowExecutionInitiatedEventAttributes        `json:"signalExternalWorkflowExecutionInitiatedEventAttributes,omitempty"`
	StartChildWorkflowExecutionFailedEventAttributes               *StartChildWorkflowExecutionFailedEventAttributes               `json:"startChildWorkflowExecutionFailedEventAttributes,omitempty"`
	StartChildWorkflowExecutionInitiatedEventAttributes            *StartChildWorkflowExecutionInitiatedEventAttributes            `json:"startChildWorkflowExecutionInitiatedEventAttributes,omitempty"`
	StartTimerFailedEventAttributes                                *StartTimerFailedEventAttributes                                `json:"startTimerFailedEventAttributes,omitempty"`
	TimerCanceledEventAttributes                                   *TimerCanceledEventAttributes                                   `json:"timerCanceledEventAttributes,omitempty"`
	TimerFiredEventAttributes                                      *TimerFiredEventAttributes                                      `json:"timerFiredEventAttributes,omitempty"`
	TimerStartedEventAttributes                                    *TimerStartedEventAttributes                                    `json:"timerStartedEventAttributes,omitempty"`
	WorkflowExecutionCancelRequestedEventAttributes                *WorkflowExecutionCancelRequestedEventAttributes                `json:"workflowExecutionCancelRequestedEventAttributes,omitempty"`
	WorkflowExecutionCanceledEventAttributes                       *WorkflowExecutionCanceledEventAttributes                       `json:"workflowExecutionCanceledEventAttributes,omitempty"`
	WorkflowExecutionCompletedEventAttributes                      *WorkflowExecutionCompletedEventAttributes                      `json:"workflowExecutionCompletedEventAttributes,omitempty"`
	WorkflowExecutionContinuedAsNewEventAttributes                 *WorkflowExecutionContinuedAsNewEventAttributes                 `json:"workflowExecutionContinuedAsNewEventAttributes,omitempty"`
	WorkflowExecutionFailedEventAttributes                         *WorkflowExecutionFailedEventAttributes                         `json:"workflowExecutionFailedEventAttributes,omitempty"`
	WorkflowExecutionSignaledEventAttributes                       *WorkflowExecutionSignaledEventAttributes                       `json:"workflowExecutionSignaledEventAttributes,omitempty"`
	WorkflowExecutionStartedEventAttributes                        *WorkflowExecutionStartedEventAttributes                        `json:"workflowExecutionStartedEventAttributes,omitempty"`
	WorkflowExecutionTerminatedEventAttributes                     *WorkflowExecutionTerminatedEventAttributes                     `json:"workflowExecutionTerminatedEventAttributes,omitempty"`
	WorkflowExecutionTimedOutEventAttributes                       *WorkflowExecutionTimedOutEventAttributes                       `json:"workflowExecutionTimedOutEventAttributes,omitempty"`
}

type ActivityTaskCancelRequestedEventAttributes struct {
	ActivityId                   string `json:"activityId"`
	DecisionTaskCompletedEventId int    `json:"decisionTaskCompletedEventId"`
}
type ActivityTaskCanceledEventAttributes struct {
	Details                      string `json:"details"`
	LatestCancelRequestedEventId int    `json:"latestCancelRequestedEventId"`
	ScheduledEventId             int    `json:"scheduledEventId"`
	StartedEventId               int    `json:"startedEventId"`
}
type ActivityTaskCompletedEventAttributes struct {
	Result           string `json:"result"`
	ScheduledEventId int    `json:"scheduledEventId"`
	StartedEventId   int    `json:"startedEventId"`
}
type ActivityTaskFailedEventAttributes struct {
	Details          string `json:"details"`
	Reason           string `json:"reason"`
	ScheduledEventId int    `json:"scheduledEventId"`
	StartedEventId   int    `json:"startedEventId"`
}
type ActivityTaskScheduledEventAttributes struct {
	ActivityId                   string       `json:"activityId"`
	ActivityType                 ActivityType `json:"activityType"`
	Control                      string       `json:"control"`
	DecisionTaskCompletedEventId int          `json:"decisionTaskCompletedEventId"`
	HeartbeatTimeout             string       `json:"heartbeatTimeout"`
	Input                        string       `json:"input"`
	ScheduleToCloseTimeout       string       `json:"scheduleToCloseTimeout"`
	ScheduleToStartTimeout       string       `json:"scheduleToStartTimeout"`
	StartToCloseTimeout          string       `json:"startToCloseTimeout"`
	TaskList                     TaskList     `json:"taskList"`
}
type ActivityTaskStartedEventAttributes struct {
	Identity         string `json:"identity"`
	ScheduledEventId int    `json:"scheduledEventId"`
}
type ActivityTaskTimedOutEventAttributes struct {
	Details          string `json:"details"`
	ScheduledEventId int    `json:"scheduledEventId"`
	StartedEventId   int    `json:"startedEventId"`
	TimeoutType      string `json:"timeoutType"`
}
type CancelTimerFailedEventAttributes struct {
	Cause                        string `json:"cause"`
	DecisionTaskCompletedEventId int    `json:"decisionTaskCompletedEventId"`
	TimerId                      string `json:"timerId"`
}
type CancelWorkflowExecutionFailedEventAttributes struct {
	Cause                        string `json:"cause"`
	DecisionTaskCompletedEventId int    `json:"decisionTaskCompletedEventId"`
}
type ChildWorkflowExecutionCanceledEventAttributes struct {
	Details           string            `json:"details"`
	InitiatedEventId  int               `json:"initiatedEventId"`
	StartedEventId    int               `json:"startedEventId"`
	WorkflowExecution WorkflowExecution `json:"workflowExecution"`
	WorkflowType      WorkflowType      `json:"workflowType"`
}
type ChildWorkflowExecutionCompletedEventAttributes struct {
	InitiatedEventId  int               `json:"initiatedEventId"`
	Result            string            `json:"result"`
	StartedEventId    int               `json:"startedEventId"`
	WorkflowExecution WorkflowExecution `json:"workflowExecution"`
	WorkflowType      WorkflowType      `json:"workflowType"`
}
type ChildWorkflowExecutionFailedEventAttributes struct {
	Details           string            `json:"details"`
	InitiatedEventId  int               `json:"initiatedEventId"`
	Reason            string            `json:"reason"`
	StartedEventId    int               `json:"startedEventId"`
	WorkflowExecution WorkflowExecution `json:"workflowExecution"`
	WorkflowType      WorkflowType      `json:"workflowType"`
}
type ChildWorkflowExecutionStartedEventAttributes struct {
	InitiatedEventId  int               `json:"initiatedEventId"`
	WorkflowExecution WorkflowExecution `json:"workflowExecution"`
	WorkflowType      WorkflowType      `json:"workflowType"`
}
type ChildWorkflowExecutionTerminatedEventAttributes struct {
	InitiatedEventId  int               `json:"initiatedEventId"`
	StartedEventId    int               `json:"startedEventId"`
	WorkflowExecution WorkflowExecution `json:"workflowExecution"`
	WorkflowType      WorkflowType      `json:"workflowType"`
}
type ChildWorkflowExecutionTimedOutEventAttributes struct {
	InitiatedEventId  int               `json:"initiatedEventId"`
	StartedEventId    int               `json:"startedEventId"`
	TimeoutType       string            `json:"timeoutType"`
	WorkflowExecution WorkflowExecution `json:"workflowExecution"`
	WorkflowType      WorkflowType      `json:"workflowType"`
}
type CompleteWorkflowExecutionFailedEventAttributes struct {
	Cause                        string `json:"cause"`
	DecisionTaskCompletedEventId int    `json:"decisionTaskCompletedEventId"`
}
type ContinueAsNewWorkflowExecutionFailedEventAttributes struct {
	Cause                        string `json:"cause"`
	DecisionTaskCompletedEventId int    `json:"decisionTaskCompletedEventId"`
}
type DecisionTaskCompletedEventAttributes struct {
	ExecutionContext string `json:"executionContext"`
	ScheduledEventId int    `json:"scheduledEventId"`
	StartedEventId   int    `json:"startedEventId"`
}
type DecisionTaskScheduledEventAttributes struct {
	StartToCloseTimeout string   `json:"startToCloseTimeout"`
	TaskList            TaskList `json:"taskList"`
}
type DecisionTaskStartedEventAttributes struct {
	Identity         string `json:"identity"`
	ScheduledEventId int    `json:"scheduledEventId"`
}
type DecisionTaskTimedOutEventAttributes struct {
	ScheduledEventId int    `json:"scheduledEventId"`
	StartedEventId   int    `json:"startedEventId"`
	TimeoutType      string `json:"timeoutType"`
}
type ExternalWorkflowExecutionCancelRequestedEventAttributes struct {
	InitiatedEventId  int               `json:"initiatedEventId"`
	WorkflowExecution WorkflowExecution `json:"workflowExecution"`
}
type ExternalWorkflowExecutionSignaledEventAttributes struct {
	InitiatedEventId  int               `json:"initiatedEventId"`
	WorkflowExecution WorkflowExecution `json:"workflowExecution"`
}
type FailWorkflowExecutionFailedEventAttributes struct {
	Cause                        string `json:"cause"`
	DecisionTaskCompletedEventId int    `json:"decisionTaskCompletedEventId"`
}
type MarkerRecordedEventAttributes struct {
	DecisionTaskCompletedEventId int    `json:"decisionTaskCompletedEventId"`
	Details                      string `json:"details"`
	MarkerName                   string `json:"markerName"`
}
type RecordMarkerFailedEventAttributes struct {
	Cause                        string `json:"cause"`
	DecisionTaskCompletedEventId int    `json:"decisionTaskCompletedEventId"`
	MarkerName                   string `json:"markerName"`
}
type RequestCancelActivityTaskFailedEventAttributes struct {
	ActivityId                   string `json:"activityId"`
	Cause                        string `json:"cause"`
	DecisionTaskCompletedEventId int    `json:"decisionTaskCompletedEventId"`
}
type RequestCancelExternalWorkflowExecutionFailedEventAttributes struct {
	Cause                        string `json:"cause"`
	Control                      string `json:"control"`
	DecisionTaskCompletedEventId int    `json:"decisionTaskCompletedEventId"`
	InitiatedEventId             int    `json:"initiatedEventId"`
	RunId                        string `json:"runId"`
	WorkflowId                   string `json:"workflowId"`
}
type RequestCancelExternalWorkflowExecutionInitiatedEventAttributes struct {
	Control                      string `json:"control"`
	DecisionTaskCompletedEventId int    `json:"decisionTaskCompletedEventId"`
	RunId                        string `json:"runId"`
	WorkflowId                   string `json:"workflowId"`
}
type ScheduleActivityTaskFailedEventAttributes struct {
	ActivityId                   string `json:"activityId"`
	ActivityType                 ActivityType
	Cause                        string `json:"cause"`
	DecisionTaskCompletedEventId int    `json:"decisionTaskCompletedEventId"`
}
type SignalExternalWorkflowExecutionFailedEventAttributes struct {
	Cause                        string `json:"cause"`
	Control                      string `json:"control"`
	DecisionTaskCompletedEventId int    `json:"decisionTaskCompletedEventId"`
	InitiatedEventId             int    `json:"initiatedEventId"`
	RunId                        string `json:"runId"`
	WorkflowId                   string `json:"workflowId"`
}
type SignalExternalWorkflowExecutionInitiatedEventAttributes struct {
	Control                      string `json:"control"`
	DecisionTaskCompletedEventId int    `json:"decisionTaskCompletedEventId"`
	Input                        string `json:"input"`
	RunId                        string `json:"runId"`
	SignalName                   string `json:"signalName"`
	WorkflowId                   string `json:"workflowId"`
}
type StartChildWorkflowExecutionFailedEventAttributes struct {
	Cause                        string       `json:"cause"`
	Control                      string       `json:"control"`
	DecisionTaskCompletedEventId int          `json:"decisionTaskCompletedEventId"`
	InitiatedEventId             int          `json:"initiatedEventId"`
	WorkflowId                   string       `json:"workflowId"`
	WorkflowType                 WorkflowType `json:"workflowType"`
}
type StartChildWorkflowExecutionInitiatedEventAttributes struct {
	ChildPolicy                  string       `json:"childPolicy"`
	Control                      string       `json:"control"`
	DecisionTaskCompletedEventId int          `json:"decisionTaskCompletedEventId"`
	ExecutionStartToCloseTimeout string       `json:"executionStartToCloseTimeout"`
	Input                        string       `json:"input"`
	TagList                      []string     `json:"tagList"`
	TaskList                     TaskList     `json:"taskList"`
	TaskStartToCloseTimeout      string       `json:"taskStartToCloseTimeout"`
	WorkflowId                   string       `json:"workflowId"`
	WorkflowType                 WorkflowType `json:"workflowType"`
}
type StartTimerFailedEventAttributes struct {
	Cause                        string `json:"cause"`
	DecisionTaskCompletedEventId int    `json:"decisionTaskCompletedEventId"`
	TimerId                      string `json:"timerId"`
}
type TimerCanceledEventAttributes struct {
	DecisionTaskCompletedEventId int    `json:"decisionTaskCompletedEventId"`
	StartedEventId               int    `json:"startedEventId"`
	TimerId                      string `json:"timerId"`
}
type TimerFiredEventAttributes struct {
	StartedEventId int    `json:"startedEventId"`
	TimerId        string `json:"timerId"`
}
type TimerStartedEventAttributes struct {
	Control                      string `json:"control"`
	DecisionTaskCompletedEventId int    `json:"decisionTaskCompletedEventId"`
	StartToFireTimeout           string `json:"startToFireTimeout"`
	TimerId                      string `json:"timerId"`
}
type WorkflowExecutionCancelRequestedEventAttributes struct {
	Cause                     string            `json:"cause"`
	ExternalInitiatedEventId  int               `json:"externalInitiatedEventId"`
	ExternalWorkflowExecution WorkflowExecution `json:"externalWorkflowExecution"`
}
type WorkflowExecutionCanceledEventAttributes struct {
	DecisionTaskCompletedEventId int    `json:"decisionTaskCompletedEventId"`
	Details                      string `json:"details"`
}
type WorkflowExecutionCompletedEventAttributes struct {
	DecisionTaskCompletedEventId int    `json:"decisionTaskCompletedEventId"`
	Result                       string `json:"result"`
}
type WorkflowExecutionContinuedAsNewEventAttributes struct {
	ChildPolicy                  string       `json:"childPolicy"`
	DecisionTaskCompletedEventId int          `json:"decisionTaskCompletedEventId"`
	ExecutionStartToCloseTimeout string       `json:"executionStartToCloseTimeout"`
	Input                        string       `json:"input"`
	NewExecutionRunId            string       `json:"newExecutionRunId"`
	TagList                      []string     `json:"tagList"`
	TaskList                     TaskList     `json:"taskList"`
	TaskStartToCloseTimeout      string       `json:"taskStartToCloseTimeout"`
	WorkflowType                 WorkflowType `json:"workflowType"`
}
type WorkflowExecutionFailedEventAttributes struct {
	DecisionTaskCompletedEventId int    `json:"decisionTaskCompletedEventId"`
	Details                      string `json:"details"`
	Reason                       string `json:"reason"`
}
type WorkflowExecutionSignaledEventAttributes struct {
	ExternalInitiatedEventId  int               `json:"externalInitiatedEventId"`
	ExternalWorkflowExecution WorkflowExecution `json:"externalWorkflowExecution"`
	Input                     string            `json:"input"`
	SignalName                string            `json:"signalName"`
}
type WorkflowExecutionStartedEventAttributes struct {
	ChildPolicy                  string            `json:"childPolicy"`
	ContinuedExecutionRunId      string            `json:"continuedExecutionRunId"`
	ExecutionStartToCloseTimeout string            `json:"executionStartToCloseTimeout"`
	Input                        string            `json:"input"`
	ParentInitiatedEventId       int               `json:"parentInitiatedEventId"`
	ParentWorkflowExecution      WorkflowExecution `json:"parentWorkflowExecution"`
	TagList                      []string          `json:"tagList"`
	TaskList                     TaskList          `json:"taskList"`
	TaskStartToCloseTimeout      string            `json:"taskStartToCloseTimeout"`
	WorkflowType                 WorkflowType      `json:"workflowType"`
}
type WorkflowExecutionTerminatedEventAttributes struct {
	Cause       string `json:"cause"`
	ChildPolicy string `json:"childPolicy"`
	Details     string `json:"details"`
	Reason      string `json:"reason"`
}
type WorkflowExecutionTimedOutEventAttributes struct {
	ChildPolicy string `json:"childPolicy"`
	TimeoutType string `json:"timeoutType"`
}

//DecisionType := ScheduleActivityTask | RequestCancelActivityTask | CompleteWorkflowExecution | FailWorkflowExecution | CancelWorkflowExecution | ContinueAsNewWorkflowExecution | RecordMarker | StartTimer | CancelTimer | SignalExternalWorkflowExecution | RequestCancelExternalWorkflowExecution | StartChildWorkflowExecution

// Valid values for the field DecisionType in the Decision struct.
const (
	DecisionTypeScheduleActivityTask                   = "ScheduleActivityTask"
	DecisionTypeRequestCancelActivityTask              = "RequestCancelActivityTask"
	DecisionTypeCompleteWorkflowExecution              = "CompleteWorkflowExecution"
	DecisionTypeFailWorkflowExecution                  = "FailWorkflowExecution"
	DecisionTypeCancelWorkflowExecution                = "CancelWorkflowExecution"
	DecisionTypeContinueAsNewWorkflowExecution         = "ContinueAsNewWorkflowExecution"
	DecisionTypeRecordMarker                           = "RecordMarker"
	DecisionTypeStartTimer                             = "StartTimer"
	DecisionTypeCancelTimer                            = "CancelTimer"
	DecisionTypeSignalExternalWorkflowExecution        = "SignalExternalWorkflowExecution"
	DecisionTypeRequestCancelExternalWorkflowExecution = "RequestCancelExternalWorkflowExecution"
	DecisionTypeStartChildWorkflowExecution            = "StartChildWorkflowExecution"
)

type Decision struct {
	CancelTimerDecisionAttributes                            *CancelTimerDecisionAttributes                            `json:"cancelTimerDecisionAttributes,omitempty"`
	CancelWorkflowExecutionDecisionAttributes                *CancelWorkflowExecutionDecisionAttributes                `json:"cancelWorkflowExecutionDecisionAttributes,omitempty"`
	CompleteWorkflowExecutionDecisionAttributes              *CompleteWorkflowExecutionDecisionAttributes              `json:"completeWorkflowExecutionDecisionAttributes,omitempty"`
	ContinueAsNewWorkflowExecutionDecisionAttributes         *ContinueAsNewWorkflowExecutionDecisionAttributes         `json:"continueAsNewWorkflowExecutionDecisionAttributes,omitempty"`
	DecisionType                                             string                                                    `json:"decisionType"`
	FailWorkflowExecutionDecisionAttributes                  *FailWorkflowExecutionDecisionAttributes                  `json:"failWorkflowExecutionDecisionAttributes,omitempty"`
	RecordMarkerDecisionAttributes                           *RecordMarkerDecisionAttributes                           `json:"recordMarkerDecisionAttributes,omitempty"`
	RequestCancelActivityTaskDecisionAttributes              *RequestCancelActivityTaskDecisionAttributes              `json:"requestCancelActivityTaskDecisionAttributes,omitempty"`
	RequestCancelExternalWorkflowExecutionDecisionAttributes *RequestCancelExternalWorkflowExecutionDecisionAttributes `json:"requestCancelExternalWorkflowExecutionDecisionAttributes,omitempty"`
	ScheduleActivityTaskDecisionAttributes                   *ScheduleActivityTaskDecisionAttributes                   `json:"scheduleActivityTaskDecisionAttributes,omitempty"`
	SignalExternalWorkflowExecutionDecisionAttributes        *SignalExternalWorkflowExecutionDecisionAttributes        `json:"signalExternalWorkflowExecutionDecisionAttributes,omitempty"`
	StartChildWorkflowExecutionDecisionAttributes            *StartChildWorkflowExecutionDecisionAttributes            `json:"startChildWorkflowExecutionDecisionAttributes,omitempty"`
	StartTimerDecisionAttributes                             *StartTimerDecisionAttributes                             `json:"startTimerDecisionAttributes,omitempty"`
}

type CancelTimerDecisionAttributes struct {
	TimerId string `json:"timerId"`
}
type CancelWorkflowExecutionDecisionAttributes struct {
	Details string `json:"details"`
}
type CompleteWorkflowExecutionDecisionAttributes struct {
	Result string `json:"result"`
}
type ContinueAsNewWorkflowExecutionDecisionAttributes struct {
	ChildPolicy                  string   `json:"childPolicy"`
	ExecutionStartToCloseTimeout string   `json:"executionStartToCloseTimeout"`
	Input                        string   `json:"input"`
	TagList                      []string `json:"tagList"`
	TaskList                     TaskList `json:"taskList"`
	TaskStartToCloseTimeout      string   `json:"taskStartToCloseTimeout"`
	WorkflowTypeVersion          string   `json:"workflowTypeVersion"`
}
type FailWorkflowExecutionDecisionAttributes struct {
	Details string `json:"details"`
	Reason  string `json:"reason"`
}
type RecordMarkerDecisionAttributes struct {
	Details    string `json:"details"`
	MarkerName string `json:"markerName"`
}
type RequestCancelActivityTaskDecisionAttributes struct {
	ActivityId string `json:"activityId"`
}
type RequestCancelExternalWorkflowExecutionDecisionAttributes struct {
	Control    string `json:"control"`
	RunId      string `json:"runId"`
	WorkflowId string `json:"workflowId"`
}
type ScheduleActivityTaskDecisionAttributes struct {
	ActivityId             string       `json:"activityId"`
	ActivityType           ActivityType `json:"activityType"`
	Control                string       `json:"control"`
	HeartbeatTimeout       string       `json:"heartbeatTimeout,omitempty"`
	Input                  string       `json:"input,omitempty"`
	ScheduleToCloseTimeout string       `json:"scheduleToCloseTimeout,omitempty"`
	ScheduleToStartTimeout string       `json:"scheduleToStartTimeout,omitempty"`
	StartToCloseTimeout    string       `json:"startToCloseTimeout,omitempty"`
	TaskList               TaskList     `json:"taskList,omitempty"`
}
type SignalExternalWorkflowExecutionDecisionAttributes struct {
	Control    string `json:"control"`
	Input      string `json:"input"`
	RunId      string `json:"runId"`
	SignalName string `json:"signalName"`
	WorkflowId string `json:"workflowId"`
}
type StartChildWorkflowExecutionDecisionAttributes struct {
	ChildPolicy                  string       `json:"childPolicy"`
	Control                      string       `json:"control"`
	ExecutionStartToCloseTimeout string       `json:"executionStartToCloseTimeout"`
	Input                        string       `json:"input"`
	TagList                      []string     `json:"tagList"`
	TaskList                     TaskList     `json:"taskList"`
	TaskStartToCloseTimeout      string       `json:"taskStartToCloseTimeout"`
	WorkflowId                   string       `json:"workflowId"`
	WorkflowType                 WorkflowType `json:"workflowType"`
}
type StartTimerDecisionAttributes struct {
	Control            string `json:"control"`
	StartToFireTimeout string `json:"startToFireTimeout"`
	TimerId            string `json:"timerId"`
}

type RespondDecisionTaskCompletedRequest struct {
	Decisions        []*Decision `json:"decisions"`
	ExecutionContext string      `json:"executionContext"`
	TaskToken        string      `json:"taskToken"`
}

/*ActivityWorkerProtocol*/

type PollForActivityTaskRequest struct {
	Domain   string   `json:"domain"`
	Identity string   `json:"identity,omitempty"`
	TaskList TaskList `json:"taskList"`
}

type PollForActivityTaskResponse struct {
	ActivityId        string            `json:"activityId"`
	ActivityType      ActivityType      `json:"activityType"`
	Input             string            `json:"input"`
	StartedEventId    int               `json:"startedEventId"`
	TaskToken         string            `json:"taskToken"`
	WorkflowExecution WorkflowExecution `json:"workflowExecution"`
}

type RespondActivityTaskCompletedRequest struct {
	Result    string `json:"result,omitempty"`
	TaskToken string `json:"taskToken"`
}

type RespondActivityTaskFailedRequest struct {
	Details   string `json:"details,omitempty"`
	Reason    string `json:"reason,omitempty"`
	TaskToken string `json:"taskToken"`
}

type RespondActivityTaskCanceledRequest struct {
	Details   string `json:"details,omitempty"`
	TaskToken string `json:"taskToken"`
}

type RecordActivityTaskHeartbeatRequest struct {
	Details   string `json:"details,omitempty"`
	TaskToken string `json:"taskToken"`
}

type RecordActivityTaskHeartbeatResponse struct {
	CancelRequested bool `json:"cancelRequested"`
}

/*admin protocol*/
type DeprecateActivityType struct {
	ActivityType ActivityType `json:"activityType"`
	Domain       string       `json:"domain"`
}

type DeprecateWorkflowType struct {
	Domain       string       `json:"domain"`
	WorkflowType WorkflowType `json:"workflowType"`
}

type DeprecateDomain struct {
	Name string `json:"name"`
}

type RegisterActivityType struct {
	DefaultTaskHeartbeatTimeout       string    `json:"defaultTaskHeartbeatTimeout,omitempty"`
	DefaultTaskList                   *TaskList `json:"defaultTaskList",omitempty`
	DefaultTaskScheduleToCloseTimeout string    `json:"defaultTaskScheduleToCloseTimeout,omitempty"`
	DefaultTaskScheduleToStartTimeout string    `json:"defaultTaskScheduleToStartTimeout,omitempty"`
	DefaultTaskStartToCloseTimeout    string    `json:"defaultTaskStartToCloseTimeout,omitempty"`
	Description                       string    `json:"description,omitempty"`
	Domain                            string    `json:"domain"`
	Name                              string    `json:"name"`
	Version                           string    `json:"version"`
}

type RegisterDomain struct {
	Description                            string `json:"description,omitempty"`
	Name                                   string `json:"name"`
	WorkflowExecutionRetentionPeriodInDays string `json:"workflowExecutionRetentionPeriodInDays"`
}

type RegisterWorkflowType struct {
	DefaultChildPolicy                  string    `json:"defaultChildPolicy,omitempty"`
	DefaultExecutionStartToCloseTimeout string    `json:"defaultExecutionStartToCloseTimeout,omitempty"`
	DefaultTaskList                     *TaskList `json:"defaultTaskList,omitempty"`
	DefaultTaskStartToCloseTimeout      string    `json:"defaultTaskStartToCloseTimeout,omitempty"`
	Description                         string    `json:"description,omitempty"`
	Domain                              string    `json:"domain"`
	Name                                string    `json:"name"`
	Version                             string    `json:"version"`
}

/*WorkflowInfoProtocol*/
type CountClosedWorkflowExecutionsRequest struct {
	CloseStatusFilter *StatusFilter    `json:"closeStatusFilter,omitempty"`
	CloseTimeFilter   *TimeFilter      `json:"closeTimeFilter,omitempty"`
	Domain            string           `json:"domain"`
	ExecutionFilter   *ExecutionFilter `json:"executionFilter,omitempty"`
	StartTimeFilter   *TimeFilter      `json:"startTimeFilter,omitempty"`
	TagFilter         *TagFilter       `json:"tagFilter,omitempty"`
	TypeFilter        *TypeFilter      `json:"typeFilter,omitempty"`
}

type CountOpenWorkflowExecutionsRequest struct {
	Domain          string           `json:"domain"`
	ExecutionFilter *ExecutionFilter `json:"executionFilter,omitempty"`
	StartTimeFilter *TimeFilter      `json:"startTimeFilter,omitempty"`
	TagFilter       *TagFilter       `json:"tagFilter,omitempty"`
	TypeFilter      *TypeFilter      `json:"typeFilter,omitempty"`
}

type CountPendingActivityTasksRequest struct {
	Domain   string   `json:"domain"`
	TaskList TaskList `json:"taskList"`
}

type CountPendingDecisionTasksRequest struct {
	Domain   string   `json:"domain"`
	TaskList TaskList `json:"taskList"`
}

type DescribeActivityTypeRequest struct {
	ActivityType ActivityType `json:"activityType"`
	Domain       string       `json:"domain"`
}

type DescribeActivityTypeResponse struct {
	Configuration ActivityTypeConfiguration `json:"configuration"`
	TypeInfo      ActivityTypeInfo          `json:"typeInfo"`
}

type ActivityTypeConfiguration struct {
	DefaultTaskHeartbeatTimeout       string   `json:"defaultTaskHeartbeatTimeout"`
	DefaultTaskList                   TaskList `json:"defaultTaskList"`
	DefaultTaskScheduleToCloseTimeout string   `json:"defaultTaskScheduleToCloseTimeout"`
	DefaultTaskScheduleToStartTimeout string   `json:"defaultTaskScheduleToStartTimeout"`
	DefaultTaskStartToCloseTimeout    string   `json:"defaultTaskStartToCloseTimeout"`
}

type DescribeDomainRequest struct {
	Name string `json:"name"`
}

type DescribeDomainResponse struct {
	Configuration DomainConfiguration `json:"configuration"`
	DomainInfo    DomainInfo          `json:"domainInfo"`
}

type DomainConfiguration struct {
	WorkflowExecutionRetentionPeriodInDays string `json:"workflowExecutionRetentionPeriodInDays"`
}

type DomainInfo struct {
	Description string `json:"description"`
	Name        string `json:"name"`
	Status      string `json:"status"`
}

type DescribeWorkflowExecutionRequest struct {
	Domain    string            `json:"domain"`
	Execution WorkflowExecution `json:"execution"`
}

type DescribeWorkflowExecutionResponse struct {
	ExecutionConfiguration      ExecutionConfiguration `json:"executionConfiguration"`
	ExecutionInfo               WorkflowExecutionInfo  `json:"executionInfo"`
	LatestActivityTaskTimestamp float32                `json:"latestActivityTaskTimestamp"`
	LatestExecutionContext      string                 `json:"latestExecutionContext"`
	OpenCounts                  OpenCounts             `json:"openCounts"`
}

type ExecutionConfiguration struct {
	ChildPolicy                  string   `json:"childPolicy"`
	ExecutionStartToCloseTimeout string   `json:"executionStartToCloseTimeout"`
	TaskList                     TaskList `json:"taskList"`
	TaskStartToCloseTimeout      string   `json:"taskStartToCloseTimeout"`
}

type OpenCounts struct {
	OpenActivityTasks           string `json:"openActivityTasks"`
	OpenChildWorkflowExecutions string `json:"openChildWorkflowExecutions"`
	OpenDecisionTasks           string `json:"openDecisionTasks"`
	OpenTimers                  string `json:"openTimers"`
}

type DescribeWorkflowTypeRequest struct {
	Domain       string       `json:"domain"`
	WorkflowType WorkflowType `json:"workflowType"`
}

type DescribeWorkflowTypeResponse struct {
	Configuration WorkflowConfiguration `json:"configuration"`
	TypeInfo      WorkflowTypeInfo      `json:"typeInfo"`
}

type WorkflowConfiguration struct {
	DefaultChildPolicy                  string   `json:"defaultChildPolicy"`
	DefaultExecutionStartToCloseTimeout string   `json:"defaultExecutionStartToCloseTimeout"`
	DefaultTaskList                     TaskList `json:"defaultTaskList"`
	DefaultTaskStartToCloseTimeout      string   `json:"defaultTaskStartToCloseTimeout"`
}

type GetWorkflowExecutionHistoryRequest struct {
	Domain          string            `json:"domain"`
	Execution       WorkflowExecution `json:"execution"`
	MaximumPageSize int               `json:"maximumPageSize,omitempty"`
	NextPageToken   string            `json:"nextPageToken,omitempty"`
	ReverseOrder    bool              `json:"reverseOrder,omitempty"`
}

type GetWorkflowExecutionHistoryResponse struct {
	Events        []HistoryEvent `json:"events"`
	NextPageToken string         `json:"nextPageToken,omitempty"`
}

type ListActivityTypesRequest struct {
	Domain             string `json:"domain"`
	MaximumPageSize    int    `json:"maximumPageSize,omitempty"`
	Name               string `json:"name,omitempty"`
	NextPageToken      string `json:"nextPageToken,omitempty"`
	RegistrationStatus string `json:"registrationStatus"`
	ReverseOrder       bool   `json:"reverseOrder,omitempty"`
}

type ListActivityTypesResponse struct {
	NextPageToken *string            `json:"nextPageToken"`
	TypeInfos     []ActivityTypeInfo `json:"typeInfos"`
}

type ListClosedWorkflowExecutionsRequest struct {
	CloseStatusFilter *StatusFilter    `json:"closeStatusFilter,omitempty"`
	CloseTimeFilter   *TimeFilter      `json:"closeTimeFilter,omitempty"`
	Domain            string           `json:"domain"`
	ExecutionFilter   *ExecutionFilter `json:"executionFilter,omitempty"`
	MaximumPageSize   int              `json:"maximumPageSize,omitempty"`
	NextPageToken     string           `json:"nextPageToken,omitempty"`
	ReverseOrder      bool             `json:"reverseOrde,omitemptyr"`
	StartTimeFilter   *TimeFilter      `json:"startTimeFilter,omitempty"`
	TagFilter         *TagFilter       `json:"tagFilter,omitempty"`
	TypeFilter        *TypeFilter      `json:"typeFilter,omitempty"`
}

type ListClosedWorkflowExecutionsResponse struct {
	ExecutionInfos []WorkflowExecutionInfo `json:"executionInfos"`
	NextPageToken  string                  `json:"nextPageToken,omitempty"`
}

type ListDomainsRequest struct {
	MaximumPageSize    int    `json:"maximumPageSize,omitempty"`
	NextPageToken      string `json:"nextPageToken,omitempty"`
	RegistrationStatus string `json:"registrationStatus"`
	ReverseOrder       bool   `json:"reverseOrder,omitempty"`
}

type ListDomainsResponse struct {
	DomainInfos   []DomainInfo `json:"domainInfos"`
	NextPageToken string       `json:"nextPageToken,omitempty"`
}

type ListOpenWorkflowExecutionsRequest struct {
	Domain          string           `json:"domain"`
	ExecutionFilter *ExecutionFilter `json:"executionFilter,omitempty"`
	MaximumPageSize int              `json:"maximumPageSize,omitempty"`
	NextPageToken   string           `json:"nextPageToken,omitempty"`
	ReverseOrder    bool             `json:"reverseOrder,omitempty"`
	StartTimeFilter TimeFilter       `json:"startTimeFilter"`
	TagFilter       *TagFilter       `json:"tagFilter,omitempty"`
	TypeFilter      *TypeFilter      `json:"typeFilter,omitempty"`
}

type ListOpenWorkflowExecutionsResponse struct {
	ExecutionInfos []WorkflowExecutionInfo `json:"executionInfos"`
	NextPageToken  string                  `json:"nextPageToken,omitempty"`
}

type WorkflowExecutionInfo struct {
	CancelRequested bool              `json:"cancelRequested"`
	CloseStatus     string            `json:"closeStatus"`
	CloseTimestamp  float32           `json:"closeTimestamp"`
	Execution       WorkflowExecution `json:"execution"`
	ExecutionStatus string            `json:"executionStatus"`
	Parent          WorkflowExecution `json:"parent"`
	StartTimestamp  float32           `json:"startTimestamp"`
	TagList         []string          `json:"tagList"`
	WorkflowType    WorkflowType      `json:"workflowType"`
}

type StatusFilter struct {
	Status string `json:"status"`
}

type CountResponse struct {
	Count     int  `json:"count"`
	Truncated bool `json:"truncated"`
}
type TimeFilter struct {
	LatestDate float32 `json:"latestDate,omitempty"`
	OldestDate float32 `json:"oldestDate"`
}

type ExecutionFilter struct {
	WorkflowId string `json:"workflowId"`
}

type TagFilter struct {
	Tag string `json:"tag"`
}

type TypeFilter struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

/*common types*/

type TaskList struct {
	Name string `json:"name"`
}
type WorkflowType struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type WorkflowExecution struct {
	RunId      string `json:"runId"`
	WorkflowId string `json:"workflowId"`
}

type ActivityType struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type WorkflowTypeInfo struct {
	CreationDate    float32      `json:"creationDate"`
	DeprecationDate float32      `json:"deprecationDate"`
	Description     string       `json:"description"`
	Status          string       `json:"status"`
	WorkflowType    WorkflowType `json:"workflowType"`
}

type ActivityTypeInfo struct {
	CreationDate    float32      `json:"creationDate"`
	DeprecationDate float32      `json:"deprecationDate"`
	Description     string       `json:"description"`
	Status          string       `json:"status"`
	ActivityType    ActivityType `json:"activityType"`
}
