package swf

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"time"
)

// ErrorResponse models the swf json protocol.
type ErrorResponse struct {
	StatusCode int
	Type       string `json:"__type"`
	Message    string `json:"message"`
}

// constants for various SWF and Kinesis status and error codes.
const (
	StatusRegistered                              = "REGISTERED"
	StatusDeprecated                              = "DEPRECATED"
	ErrorTypeUnknownResourceFault                 = "com.amazonaws.swf.base.model#UnknownResourceFault"
	ErrorTypeWorkflowExecutionAlreadyStartedFault = "com.amazonaws.swf.base.model#WorkflowExecutionAlreadyStartedFault"
	ErrorTypeStreamNotFound                       = "ResourceNotFoundException"
)

func (err *ErrorResponse) Error() string {
	return err.Type + ": " + err.Message
}

// StartWorkflowRequest models the swf json protocol.
type StartWorkflowRequest struct {
	ChildPolicy                  string       `json:"childPolicy,omitempty"`
	Domain                       string       `json:"domain"`
	ExecutionStartToCloseTimeout string       `json:"executionStartToCloseTimeout,omitempty"`
	Input                        string       `json:"input,omitempty"`
	TagList                      []string     `json:"tagList,omitempty"`
	TaskList                     *TaskList    `json:"taskList,omitempty"`
	TaskStartToCloseTimeout      string       `json:"taskStartToCloseTimeout,omitempty"`
	WorkflowID                   string       `json:"workflowId"`
	WorkflowType                 WorkflowType `json:"workflowType"`
}

// StartWorkflowResponse models the swf json protocol.
type StartWorkflowResponse struct {
	RunID string `json:"runId"`
}

// RequestCancelWorkflowExecution models the swf json protocol.
type RequestCancelWorkflowExecution struct {
	Domain     string `json:"domain"`
	RunID      string `json:"runId,omitempty"`
	WorkflowID string `json:"workflowId"`
}

// SignalWorkflowRequest models the swf json protocol.
type SignalWorkflowRequest struct {
	Domain     string `json:"domain"`
	Input      string `json:"input,omitempty"`
	RunID      string `json:"runId,omitempty"`
	SignalName string `json:"signalName"`
	WorkflowID string `json:"workflowId"`
}

// ListWorkflowTypesRequest models the swf json protocol.
type ListWorkflowTypesRequest struct {
	Domain             string `json:"domain"`
	MaximumPageSize    int    `json:"maximumPageSize,omitempty"`
	Name               string `json:"name,omitempty"`
	NextPageToken      string `json:"nextPageToken,omitempty"`
	RegistrationStatus string `json:"registrationStatus"`
	ReverseOrder       bool   `json:"reverseOrder,omitempty"`
}

// ListWorkflowTypesResponse models the swf json protocol.
type ListWorkflowTypesResponse struct {
	NextPageToken string             `json:"nextPageToken,omitempty"`
	TypeInfos     []WorkflowTypeInfo `json:"typeInfos"`
}

// TerminateWorkflowExecution models the swf json protocol.
type TerminateWorkflowExecution struct {
	ChildPolicy string `json:"childPolicy,omitempty"`
	Details     string `json:"details,omitempty"`
	Domain      string `json:"domain"`
	Reason      string `json:"reason,omitempty"`
	RunID       string `json:"runId,omitempty"`
	WorkflowID  string `json:"workflowId,omitempty"`
}

/*DecisionWorkerProtocol*/

// PollForDecisionTaskRequest models the swf json protocol.
type PollForDecisionTaskRequest struct {
	Domain          string   `json:"domain"`
	Identity        string   `json:"identity,omitempty"`
	MaximumPageSize int      `json:"maximumPageSize,omitempty"`
	NextPageToken   string   `json:"nextPageToken,omitempty"`
	ReverseOrder    bool     `json:"reverseOrder,omitempty"`
	TaskList        TaskList `json:"taskList"`
}

// PollForDecisionTaskResponse models the swf json protocol.
type PollForDecisionTaskResponse struct {
	Events                 []HistoryEvent    `json:"events"`
	NextPageToken          string            `json:"nextPageToken"`
	PreviousStartedEventID int               `json:"previousStartedEventId"`
	StartedEventID         int               `json:"startedEventId"`
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

var eventTypes = map[string]func(HistoryEvent) interface{}{
	EventTypeWorkflowExecutionStarted:                 func(h HistoryEvent) interface{} { return h.WorkflowExecutionStartedEventAttributes },
	EventTypeWorkflowExecutionCancelRequested:         func(h HistoryEvent) interface{} { return h.WorkflowExecutionCancelRequestedEventAttributes },
	EventTypeWorkflowExecutionCompleted:               func(h HistoryEvent) interface{} { return h.WorkflowExecutionCompletedEventAttributes },
	EventTypeCompleteWorkflowExecutionFailed:          func(h HistoryEvent) interface{} { return h.CompleteWorkflowExecutionFailedEventAttributes },
	EventTypeWorkflowExecutionFailed:                  func(h HistoryEvent) interface{} { return h.WorkflowExecutionFailedEventAttributes },
	EventTypeFailWorkflowExecutionFailed:              func(h HistoryEvent) interface{} { return h.FailWorkflowExecutionFailedEventAttributes },
	EventTypeWorkflowExecutionTimedOut:                func(h HistoryEvent) interface{} { return h.WorkflowExecutionTimedOutEventAttributes },
	EventTypeWorkflowExecutionCanceled:                func(h HistoryEvent) interface{} { return h.WorkflowExecutionCanceledEventAttributes },
	EventTypeCancelWorkflowExecutionFailed:            func(h HistoryEvent) interface{} { return h.CancelWorkflowExecutionFailedEventAttributes },
	EventTypeWorkflowExecutionContinuedAsNew:          func(h HistoryEvent) interface{} { return h.WorkflowExecutionContinuedAsNewEventAttributes },
	EventTypeContinueAsNewWorkflowExecutionFailed:     func(h HistoryEvent) interface{} { return h.ContinueAsNewWorkflowExecutionFailedEventAttributes },
	EventTypeWorkflowExecutionTerminated:              func(h HistoryEvent) interface{} { return h.WorkflowExecutionTerminatedEventAttributes },
	EventTypeDecisionTaskScheduled:                    func(h HistoryEvent) interface{} { return h.DecisionTaskScheduledEventAttributes },
	EventTypeDecisionTaskStarted:                      func(h HistoryEvent) interface{} { return h.DecisionTaskStartedEventAttributes },
	EventTypeDecisionTaskCompleted:                    func(h HistoryEvent) interface{} { return h.DecisionTaskCompletedEventAttributes },
	EventTypeDecisionTaskTimedOut:                     func(h HistoryEvent) interface{} { return h.DecisionTaskTimedOutEventAttributes },
	EventTypeActivityTaskScheduled:                    func(h HistoryEvent) interface{} { return h.ActivityTaskScheduledEventAttributes },
	EventTypeScheduleActivityTaskFailed:               func(h HistoryEvent) interface{} { return h.ScheduleActivityTaskFailedEventAttributes },
	EventTypeActivityTaskStarted:                      func(h HistoryEvent) interface{} { return h.ActivityTaskStartedEventAttributes },
	EventTypeActivityTaskCompleted:                    func(h HistoryEvent) interface{} { return h.ActivityTaskCompletedEventAttributes },
	EventTypeActivityTaskFailed:                       func(h HistoryEvent) interface{} { return h.ActivityTaskFailedEventAttributes },
	EventTypeActivityTaskTimedOut:                     func(h HistoryEvent) interface{} { return h.ActivityTaskTimedOutEventAttributes },
	EventTypeActivityTaskCanceled:                     func(h HistoryEvent) interface{} { return h.ActivityTaskCanceledEventAttributes },
	EventTypeActivityTaskCancelRequested:              func(h HistoryEvent) interface{} { return h.ActivityTaskCancelRequestedEventAttributes },
	EventTypeRequestCancelActivityTaskFailed:          func(h HistoryEvent) interface{} { return h.RequestCancelActivityTaskFailedEventAttributes },
	EventTypeWorkflowExecutionSignaled:                func(h HistoryEvent) interface{} { return h.WorkflowExecutionSignaledEventAttributes },
	EventTypeMarkerRecorded:                           func(h HistoryEvent) interface{} { return h.MarkerRecordedEventAttributes },
	EventTypeRecordMarkerFailed:                       func(h HistoryEvent) interface{} { return h.RecordMarkerFailedEventAttributes },
	EventTypeTimerStarted:                             func(h HistoryEvent) interface{} { return h.TimerStartedEventAttributes },
	EventTypeStartTimerFailed:                         func(h HistoryEvent) interface{} { return h.StartTimerFailedEventAttributes },
	EventTypeTimerFired:                               func(h HistoryEvent) interface{} { return h.TimerFiredEventAttributes },
	EventTypeTimerCanceled:                            func(h HistoryEvent) interface{} { return h.TimerCanceledEventAttributes },
	EventTypeCancelTimerFailed:                        func(h HistoryEvent) interface{} { return h.CancelTimerFailedEventAttributes },
	EventTypeStartChildWorkflowExecutionInitiated:     func(h HistoryEvent) interface{} { return h.StartChildWorkflowExecutionInitiatedEventAttributes },
	EventTypeStartChildWorkflowExecutionFailed:        func(h HistoryEvent) interface{} { return h.StartChildWorkflowExecutionFailedEventAttributes },
	EventTypeChildWorkflowExecutionStarted:            func(h HistoryEvent) interface{} { return h.ChildWorkflowExecutionStartedEventAttributes },
	EventTypeChildWorkflowExecutionCompleted:          func(h HistoryEvent) interface{} { return h.ChildWorkflowExecutionCompletedEventAttributes },
	EventTypeChildWorkflowExecutionFailed:             func(h HistoryEvent) interface{} { return h.ChildWorkflowExecutionFailedEventAttributes },
	EventTypeChildWorkflowExecutionTimedOut:           func(h HistoryEvent) interface{} { return h.ChildWorkflowExecutionTimedOutEventAttributes },
	EventTypeChildWorkflowExecutionCanceled:           func(h HistoryEvent) interface{} { return h.ChildWorkflowExecutionCanceledEventAttributes },
	EventTypeChildWorkflowExecutionTerminated:         func(h HistoryEvent) interface{} { return h.ChildWorkflowExecutionTerminatedEventAttributes },
	EventTypeSignalExternalWorkflowExecutionInitiated: func(h HistoryEvent) interface{} { return h.SignalExternalWorkflowExecutionInitiatedEventAttributes },
	EventTypeSignalExternalWorkflowExecutionFailed:    func(h HistoryEvent) interface{} { return h.SignalExternalWorkflowExecutionFailedEventAttributes },
	EventTypeExternalWorkflowExecutionSignaled:        func(h HistoryEvent) interface{} { return h.ExternalWorkflowExecutionSignaledEventAttributes },
	EventTypeRequestCancelExternalWorkflowExecutionInitiated: func(h HistoryEvent) interface{} {
		return h.RequestCancelExternalWorkflowExecutionInitiatedEventAttributes
	},
	EventTypeRequestCancelExternalWorkflowExecutionFailed: func(h HistoryEvent) interface{} { return h.RequestCancelExternalWorkflowExecutionFailedEventAttributes },
	EventTypeExternalWorkflowExecutionCancelRequested:     func(h HistoryEvent) interface{} { return h.ExternalWorkflowExecutionCancelRequestedEventAttributes },
}

// HistoryEvent models the swf json protocol.
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
	EventID                                                        int                                                             `json:"eventId"`
	EventTimestamp                                                 *Date                                                           `json:"eventTimestamp"`
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

func (h HistoryEvent) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("HistoryEvent{ ")
	buffer.WriteString(fmt.Sprintf("EventId: %d,", h.EventID))
	buffer.WriteString(fmt.Sprintf("EventTimestamp: %s, ", h.EventTimestamp))
	buffer.WriteString(fmt.Sprintf("EventType:, %s", h.EventType))
	buffer.WriteString(fmt.Sprintf("%+v", eventTypes[h.EventType](h)))
	buffer.WriteString(" }")
	return buffer.String()
}

// EventTypes returns a slice containing the valid values of the HistoryEvent.EventType field.
func EventTypes() []string {
	es := make([]string, 0, len(eventTypes))
	for k := range eventTypes {
		es = append(es, k)
	}
	return es
}

// ActivityTaskCancelRequestedEventAttributes models the swf json protocol.
type ActivityTaskCancelRequestedEventAttributes struct {
	ActivityID                   string `json:"activityId"`
	DecisionTaskCompletedEventID int    `json:"decisionTaskCompletedEventId"`
}

// ActivityTaskCanceledEventAttributes models the swf json protocol.
type ActivityTaskCanceledEventAttributes struct {
	Details                      string `json:"details"`
	LatestCancelRequestedEventID int    `json:"latestCancelRequestedEventId"`
	ScheduledEventID             int    `json:"scheduledEventId"`
	StartedEventID               int    `json:"startedEventId"`
}

// ActivityTaskCompletedEventAttributes models the swf json protocol.
type ActivityTaskCompletedEventAttributes struct {
	Result           string `json:"result"`
	ScheduledEventID int    `json:"scheduledEventId"`
	StartedEventID   int    `json:"startedEventId"`
}

// ActivityTaskFailedEventAttributes models the swf json protocol.
type ActivityTaskFailedEventAttributes struct {
	Details          string `json:"details"`
	Reason           string `json:"reason"`
	ScheduledEventID int    `json:"scheduledEventId"`
	StartedEventID   int    `json:"startedEventId"`
}

// ActivityTaskScheduledEventAttributes models the swf json protocol.
type ActivityTaskScheduledEventAttributes struct {
	ActivityID                   string       `json:"activityId"`
	ActivityType                 ActivityType `json:"activityType"`
	Control                      string       `json:"control"`
	DecisionTaskCompletedEventID int          `json:"decisionTaskCompletedEventId"`
	HeartbeatTimeout             string       `json:"heartbeatTimeout"`
	Input                        string       `json:"input"`
	ScheduleToCloseTimeout       string       `json:"scheduleToCloseTimeout"`
	ScheduleToStartTimeout       string       `json:"scheduleToStartTimeout"`
	StartToCloseTimeout          string       `json:"startToCloseTimeout"`
	TaskList                     TaskList     `json:"taskList"`
}

// ActivityTaskStartedEventAttributes models the swf json protocol.
type ActivityTaskStartedEventAttributes struct {
	Identity         string `json:"identity"`
	ScheduledEventID int    `json:"scheduledEventId"`
}

// ActivityTaskTimedOutEventAttributes models the swf json protocol.
type ActivityTaskTimedOutEventAttributes struct {
	Details          string `json:"details"`
	ScheduledEventID int    `json:"scheduledEventId"`
	StartedEventID   int    `json:"startedEventId"`
	TimeoutType      string `json:"timeoutType"`
}

// CancelTimerFailedEventAttributes models the swf json protocol.
type CancelTimerFailedEventAttributes struct {
	Cause                        string `json:"cause"`
	DecisionTaskCompletedEventID int    `json:"decisionTaskCompletedEventId"`
	TimerID                      string `json:"timerId"`
}

// CancelWorkflowExecutionFailedEventAttributes models the swf json protocol.
type CancelWorkflowExecutionFailedEventAttributes struct {
	Cause                        string `json:"cause"`
	DecisionTaskCompletedEventID int    `json:"decisionTaskCompletedEventId"`
}

// ChildWorkflowExecutionCanceledEventAttributes models the swf json protocol.
type ChildWorkflowExecutionCanceledEventAttributes struct {
	Details           string            `json:"details"`
	InitiatedEventID  int               `json:"initiatedEventId"`
	StartedEventID    int               `json:"startedEventId"`
	WorkflowExecution WorkflowExecution `json:"workflowExecution"`
	WorkflowType      WorkflowType      `json:"workflowType"`
}

// ChildWorkflowExecutionCompletedEventAttributes models the swf json protocol.
type ChildWorkflowExecutionCompletedEventAttributes struct {
	InitiatedEventID  int               `json:"initiatedEventId"`
	Result            string            `json:"result"`
	StartedEventID    int               `json:"startedEventId"`
	WorkflowExecution WorkflowExecution `json:"workflowExecution"`
	WorkflowType      WorkflowType      `json:"workflowType"`
}

// ChildWorkflowExecutionFailedEventAttributes models the swf json protocol.
type ChildWorkflowExecutionFailedEventAttributes struct {
	Details           string            `json:"details"`
	InitiatedEventID  int               `json:"initiatedEventId"`
	Reason            string            `json:"reason"`
	StartedEventID    int               `json:"startedEventId"`
	WorkflowExecution WorkflowExecution `json:"workflowExecution"`
	WorkflowType      WorkflowType      `json:"workflowType"`
}

// ChildWorkflowExecutionStartedEventAttributes models the swf json protocol.
type ChildWorkflowExecutionStartedEventAttributes struct {
	InitiatedEventID  int               `json:"initiatedEventId"`
	WorkflowExecution WorkflowExecution `json:"workflowExecution"`
	WorkflowType      WorkflowType      `json:"workflowType"`
}

// ChildWorkflowExecutionTerminatedEventAttributes models the swf json protocol.
type ChildWorkflowExecutionTerminatedEventAttributes struct {
	InitiatedEventID  int               `json:"initiatedEventId"`
	StartedEventID    int               `json:"startedEventId"`
	WorkflowExecution WorkflowExecution `json:"workflowExecution"`
	WorkflowType      WorkflowType      `json:"workflowType"`
}

// ChildWorkflowExecutionTimedOutEventAttributes models the swf json protocol.
type ChildWorkflowExecutionTimedOutEventAttributes struct {
	InitiatedEventID  int               `json:"initiatedEventId"`
	StartedEventID    int               `json:"startedEventId"`
	TimeoutType       string            `json:"timeoutType"`
	WorkflowExecution WorkflowExecution `json:"workflowExecution"`
	WorkflowType      WorkflowType      `json:"workflowType"`
}

// CompleteWorkflowExecutionFailedEventAttributes models the swf json protocol.
type CompleteWorkflowExecutionFailedEventAttributes struct {
	Cause                        string `json:"cause"`
	DecisionTaskCompletedEventID int    `json:"decisionTaskCompletedEventId"`
}

// ContinueAsNewWorkflowExecutionFailedEventAttributes models the swf json protocol.
type ContinueAsNewWorkflowExecutionFailedEventAttributes struct {
	Cause                        string `json:"cause"`
	DecisionTaskCompletedEventID int    `json:"decisionTaskCompletedEventId"`
}

// DecisionTaskCompletedEventAttributes models the swf json protocol.
type DecisionTaskCompletedEventAttributes struct {
	ExecutionContext string `json:"executionContext"`
	ScheduledEventID int    `json:"scheduledEventId"`
	StartedEventID   int    `json:"startedEventId"`
}

// DecisionTaskScheduledEventAttributes models the swf json protocol.
type DecisionTaskScheduledEventAttributes struct {
	StartToCloseTimeout string   `json:"startToCloseTimeout"`
	TaskList            TaskList `json:"taskList"`
}

// DecisionTaskStartedEventAttributes models the swf json protocol.
type DecisionTaskStartedEventAttributes struct {
	Identity         string `json:"identity"`
	ScheduledEventID int    `json:"scheduledEventId"`
}

// DecisionTaskTimedOutEventAttributes models the swf json protocol.
type DecisionTaskTimedOutEventAttributes struct {
	ScheduledEventID int    `json:"scheduledEventId"`
	StartedEventID   int    `json:"startedEventId"`
	TimeoutType      string `json:"timeoutType"`
}

// ExternalWorkflowExecutionCancelRequestedEventAttributes models the swf json protocol.
type ExternalWorkflowExecutionCancelRequestedEventAttributes struct {
	InitiatedEventID  int               `json:"initiatedEventId"`
	WorkflowExecution WorkflowExecution `json:"workflowExecution"`
}

// ExternalWorkflowExecutionSignaledEventAttributes models the swf json protocol.
type ExternalWorkflowExecutionSignaledEventAttributes struct {
	InitiatedEventID  int               `json:"initiatedEventId"`
	WorkflowExecution WorkflowExecution `json:"workflowExecution"`
}

// FailWorkflowExecutionFailedEventAttributes models the swf json protocol.
type FailWorkflowExecutionFailedEventAttributes struct {
	Cause                        string `json:"cause"`
	DecisionTaskCompletedEventID int    `json:"decisionTaskCompletedEventId"`
}

// MarkerRecordedEventAttributes models the swf json protocol.
type MarkerRecordedEventAttributes struct {
	DecisionTaskCompletedEventID int    `json:"decisionTaskCompletedEventId"`
	Details                      string `json:"details"`
	MarkerName                   string `json:"markerName"`
}

// RecordMarkerFailedEventAttributes models the swf json protocol.
type RecordMarkerFailedEventAttributes struct {
	Cause                        string `json:"cause"`
	DecisionTaskCompletedEventID int    `json:"decisionTaskCompletedEventId"`
	MarkerName                   string `json:"markerName"`
}

// RequestCancelActivityTaskFailedEventAttributes models the swf json protocol.
type RequestCancelActivityTaskFailedEventAttributes struct {
	ActivityID                   string `json:"activityId"`
	Cause                        string `json:"cause"`
	DecisionTaskCompletedEventID int    `json:"decisionTaskCompletedEventId"`
}

// RequestCancelExternalWorkflowExecutionFailedEventAttributes models the swf json protocol.
type RequestCancelExternalWorkflowExecutionFailedEventAttributes struct {
	Cause                        string `json:"cause"`
	Control                      string `json:"control"`
	DecisionTaskCompletedEventID int    `json:"decisionTaskCompletedEventId"`
	InitiatedEventID             int    `json:"initiatedEventId"`
	RunID                        string `json:"runId"`
	WorkflowID                   string `json:"workflowId"`
}

// RequestCancelExternalWorkflowExecutionInitiatedEventAttributes models the swf json protocol.
type RequestCancelExternalWorkflowExecutionInitiatedEventAttributes struct {
	Control                      string `json:"control"`
	DecisionTaskCompletedEventID int    `json:"decisionTaskCompletedEventId"`
	RunID                        string `json:"runId"`
	WorkflowID                   string `json:"workflowId"`
}

// ScheduleActivityTaskFailedEventAttributes models the swf json protocol.
type ScheduleActivityTaskFailedEventAttributes struct {
	ActivityID                   string `json:"activityId"`
	ActivityType                 ActivityType
	Cause                        string `json:"cause"`
	DecisionTaskCompletedEventID int    `json:"decisionTaskCompletedEventId"`
}

// SignalExternalWorkflowExecutionFailedEventAttributes models the swf json protocol.
type SignalExternalWorkflowExecutionFailedEventAttributes struct {
	Cause                        string `json:"cause"`
	Control                      string `json:"control"`
	DecisionTaskCompletedEventID int    `json:"decisionTaskCompletedEventId"`
	InitiatedEventID             int    `json:"initiatedEventId"`
	RunID                        string `json:"runId"`
	WorkflowID                   string `json:"workflowId"`
}

// SignalExternalWorkflowExecutionInitiatedEventAttributes models the swf json protocol.
type SignalExternalWorkflowExecutionInitiatedEventAttributes struct {
	Control                      string `json:"control"`
	DecisionTaskCompletedEventID int    `json:"decisionTaskCompletedEventId"`
	Input                        string `json:"input"`
	RunID                        string `json:"runId"`
	SignalName                   string `json:"signalName"`
	WorkflowID                   string `json:"workflowId"`
}

// StartChildWorkflowExecutionFailedEventAttributes models the swf json protocol.
type StartChildWorkflowExecutionFailedEventAttributes struct {
	Cause                        string       `json:"cause"`
	Control                      string       `json:"control"`
	DecisionTaskCompletedEventID int          `json:"decisionTaskCompletedEventId"`
	InitiatedEventID             int          `json:"initiatedEventId"`
	WorkflowID                   string       `json:"workflowId"`
	WorkflowType                 WorkflowType `json:"workflowType"`
}

// StartChildWorkflowExecutionInitiatedEventAttributes models the swf json protocol.
type StartChildWorkflowExecutionInitiatedEventAttributes struct {
	ChildPolicy                  string       `json:"childPolicy"`
	Control                      string       `json:"control"`
	DecisionTaskCompletedEventID int          `json:"decisionTaskCompletedEventId"`
	ExecutionStartToCloseTimeout string       `json:"executionStartToCloseTimeout"`
	Input                        string       `json:"input"`
	TagList                      []string     `json:"tagList"`
	TaskList                     TaskList     `json:"taskList"`
	TaskStartToCloseTimeout      string       `json:"taskStartToCloseTimeout"`
	WorkflowID                   string       `json:"workflowId"`
	WorkflowType                 WorkflowType `json:"workflowType"`
}

// StartTimerFailedEventAttributes models the swf json protocol.
type StartTimerFailedEventAttributes struct {
	Cause                        string `json:"cause"`
	DecisionTaskCompletedEventID int    `json:"decisionTaskCompletedEventId"`
	TimerID                      string `json:"timerId"`
}

// TimerCanceledEventAttributes models the swf json protocol.
type TimerCanceledEventAttributes struct {
	DecisionTaskCompletedEventID int    `json:"decisionTaskCompletedEventId"`
	StartedEventID               int    `json:"startedEventId"`
	TimerID                      string `json:"timerId"`
}

// TimerFiredEventAttributes models the swf json protocol.
type TimerFiredEventAttributes struct {
	StartedEventID int    `json:"startedEventId"`
	TimerID        string `json:"timerId"`
}

// TimerStartedEventAttributes models the swf json protocol.
type TimerStartedEventAttributes struct {
	Control                      string `json:"control"`
	DecisionTaskCompletedEventID int    `json:"decisionTaskCompletedEventId"`
	StartToFireTimeout           string `json:"startToFireTimeout"`
	TimerID                      string `json:"timerId"`
}

// WorkflowExecutionCancelRequestedEventAttributes models the swf json protocol.
type WorkflowExecutionCancelRequestedEventAttributes struct {
	Cause                     string            `json:"cause"`
	ExternalInitiatedEventID  int               `json:"externalInitiatedEventId"`
	ExternalWorkflowExecution WorkflowExecution `json:"externalWorkflowExecution"`
}

// WorkflowExecutionCanceledEventAttributes models the swf json protocol.
type WorkflowExecutionCanceledEventAttributes struct {
	DecisionTaskCompletedEventID int    `json:"decisionTaskCompletedEventId"`
	Details                      string `json:"details"`
}

// WorkflowExecutionCompletedEventAttributes models the swf json protocol.
type WorkflowExecutionCompletedEventAttributes struct {
	DecisionTaskCompletedEventID int    `json:"decisionTaskCompletedEventId"`
	Result                       string `json:"result"`
}

// WorkflowExecutionContinuedAsNewEventAttributes models the swf json protocol.
type WorkflowExecutionContinuedAsNewEventAttributes struct {
	ChildPolicy                  string       `json:"childPolicy"`
	DecisionTaskCompletedEventID int          `json:"decisionTaskCompletedEventId"`
	ExecutionStartToCloseTimeout string       `json:"executionStartToCloseTimeout"`
	Input                        string       `json:"input"`
	NewExecutionRunID            string       `json:"newExecutionRunId"`
	TagList                      []string     `json:"tagList"`
	TaskList                     TaskList     `json:"taskList"`
	TaskStartToCloseTimeout      string       `json:"taskStartToCloseTimeout"`
	WorkflowType                 WorkflowType `json:"workflowType"`
}

// WorkflowExecutionFailedEventAttributes models the swf json protocol.
type WorkflowExecutionFailedEventAttributes struct {
	DecisionTaskCompletedEventID int    `json:"decisionTaskCompletedEventId"`
	Details                      string `json:"details"`
	Reason                       string `json:"reason"`
}

// WorkflowExecutionSignaledEventAttributes models the swf json protocol.
type WorkflowExecutionSignaledEventAttributes struct {
	ExternalInitiatedEventID  int               `json:"externalInitiatedEventId"`
	ExternalWorkflowExecution WorkflowExecution `json:"externalWorkflowExecution"`
	Input                     string            `json:"input"`
	SignalName                string            `json:"signalName"`
}

// WorkflowExecutionStartedEventAttributes models the swf json protocol.
type WorkflowExecutionStartedEventAttributes struct {
	ChildPolicy                  string            `json:"childPolicy"`
	ContinuedExecutionRunID      string            `json:"continuedExecutionRunId"`
	ExecutionStartToCloseTimeout string            `json:"executionStartToCloseTimeout"`
	Input                        string            `json:"input"`
	ParentInitiatedEventID       int               `json:"parentInitiatedEventId"`
	ParentWorkflowExecution      WorkflowExecution `json:"parentWorkflowExecution"`
	TagList                      []string          `json:"tagList"`
	TaskList                     TaskList          `json:"taskList"`
	TaskStartToCloseTimeout      string            `json:"taskStartToCloseTimeout"`
	WorkflowType                 WorkflowType      `json:"workflowType"`
}

// WorkflowExecutionTerminatedEventAttributes models the swf json protocol.
type WorkflowExecutionTerminatedEventAttributes struct {
	Cause       string `json:"cause"`
	ChildPolicy string `json:"childPolicy"`
	Details     string `json:"details"`
	Reason      string `json:"reason"`
}

// WorkflowExecutionTimedOutEventAttributes models the swf json protocol.
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

var decisionTypes = map[string]func(Decision) interface{}{
	DecisionTypeScheduleActivityTask:                   func(d Decision) interface{} { return d.ScheduleActivityTaskDecisionAttributes },
	DecisionTypeRequestCancelActivityTask:              func(d Decision) interface{} { return d.RequestCancelActivityTaskDecisionAttributes },
	DecisionTypeCompleteWorkflowExecution:              func(d Decision) interface{} { return d.CompleteWorkflowExecutionDecisionAttributes },
	DecisionTypeFailWorkflowExecution:                  func(d Decision) interface{} { return d.FailWorkflowExecutionDecisionAttributes },
	DecisionTypeCancelWorkflowExecution:                func(d Decision) interface{} { return d.CancelWorkflowExecutionDecisionAttributes },
	DecisionTypeContinueAsNewWorkflowExecution:         func(d Decision) interface{} { return d.ContinueAsNewWorkflowExecutionDecisionAttributes },
	DecisionTypeRecordMarker:                           func(d Decision) interface{} { return d.RecordMarkerDecisionAttributes },
	DecisionTypeStartTimer:                             func(d Decision) interface{} { return d.StartTimerDecisionAttributes },
	DecisionTypeCancelTimer:                            func(d Decision) interface{} { return d.CancelTimerDecisionAttributes },
	DecisionTypeSignalExternalWorkflowExecution:        func(d Decision) interface{} { return d.SignalExternalWorkflowExecutionDecisionAttributes },
	DecisionTypeRequestCancelExternalWorkflowExecution: func(d Decision) interface{} { return d.RequestCancelExternalWorkflowExecutionDecisionAttributes },
	DecisionTypeStartChildWorkflowExecution:            func(d Decision) interface{} { return d.StartChildWorkflowExecutionDecisionAttributes },
}

// Decision models the swf json protocol.
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

func (d Decision) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("Decision{ ")
	buffer.WriteString(fmt.Sprintf("DecisionType: %s,", d.DecisionType))
	buffer.WriteString(fmt.Sprintf("%+v", decisionTypes[d.DecisionType](d)))
	buffer.WriteString(" }")
	return buffer.String()
}

// DecisionTypes returns a slice containing the valid values of the Decision.DecisionType field.
func DecisionTypes() []string {
	ds := make([]string, 0, len(decisionTypes))
	for k := range eventTypes {
		ds = append(ds, k)
	}
	return ds
}

// CancelTimerDecisionAttributes models the swf json protocol.
type CancelTimerDecisionAttributes struct {
	TimerID string `json:"timerId"`
}

// CancelWorkflowExecutionDecisionAttributes models the swf json protocol.
type CancelWorkflowExecutionDecisionAttributes struct {
	Details string `json:"details"`
}

// CompleteWorkflowExecutionDecisionAttributes models the swf json protocol.
type CompleteWorkflowExecutionDecisionAttributes struct {
	Result string `json:"result"`
}

// ContinueAsNewWorkflowExecutionDecisionAttributes models the swf json protocol.
type ContinueAsNewWorkflowExecutionDecisionAttributes struct {
	ChildPolicy                  string   `json:"childPolicy"`
	ExecutionStartToCloseTimeout string   `json:"executionStartToCloseTimeout"`
	Input                        string   `json:"input"`
	TagList                      []string `json:"tagList"`
	TaskList                     TaskList `json:"taskList"`
	TaskStartToCloseTimeout      string   `json:"taskStartToCloseTimeout"`
	WorkflowTypeVersion          string   `json:"workflowTypeVersion"`
}

// FailWorkflowExecutionDecisionAttributes models the swf json protocol.
type FailWorkflowExecutionDecisionAttributes struct {
	Details string `json:"details"`
	Reason  string `json:"reason"`
}

// RecordMarkerDecisionAttributes models the swf json protocol.
type RecordMarkerDecisionAttributes struct {
	Details    string `json:"details"`
	MarkerName string `json:"markerName"`
}

// RequestCancelActivityTaskDecisionAttributes models the swf json protocol.
type RequestCancelActivityTaskDecisionAttributes struct {
	ActivityID string `json:"activityId"`
}

// RequestCancelExternalWorkflowExecutionDecisionAttributes models the swf json protocol.
type RequestCancelExternalWorkflowExecutionDecisionAttributes struct {
	Control    string `json:"control"`
	RunID      string `json:"runId"`
	WorkflowID string `json:"workflowId"`
}

// ScheduleActivityTaskDecisionAttributes models the swf json protocol.
type ScheduleActivityTaskDecisionAttributes struct {
	ActivityID             string       `json:"activityId"`
	ActivityType           ActivityType `json:"activityType"`
	Control                string       `json:"control"`
	HeartbeatTimeout       string       `json:"heartbeatTimeout,omitempty"`
	Input                  string       `json:"input,omitempty"`
	ScheduleToCloseTimeout string       `json:"scheduleToCloseTimeout,omitempty"`
	ScheduleToStartTimeout string       `json:"scheduleToStartTimeout,omitempty"`
	StartToCloseTimeout    string       `json:"startToCloseTimeout,omitempty"`
	TaskList               *TaskList    `json:"taskList,omitempty"`
}

// SignalExternalWorkflowExecutionDecisionAttributes models the swf json protocol.
type SignalExternalWorkflowExecutionDecisionAttributes struct {
	Control    string `json:"control"`
	Input      string `json:"input"`
	RunID      string `json:"runId"`
	SignalName string `json:"signalName"`
	WorkflowID string `json:"workflowId"`
}

// StartChildWorkflowExecutionDecisionAttributes models the swf json protocol.
type StartChildWorkflowExecutionDecisionAttributes struct {
	ChildPolicy                  string       `json:"childPolicy"`
	Control                      string       `json:"control"`
	ExecutionStartToCloseTimeout string       `json:"executionStartToCloseTimeout"`
	Input                        string       `json:"input"`
	TagList                      []string     `json:"tagList"`
	TaskList                     TaskList     `json:"taskList"`
	TaskStartToCloseTimeout      string       `json:"taskStartToCloseTimeout"`
	WorkflowID                   string       `json:"workflowId"`
	WorkflowType                 WorkflowType `json:"workflowType"`
}

// StartTimerDecisionAttributes models the swf json protocol.
type StartTimerDecisionAttributes struct {
	Control            string `json:"control"`
	StartToFireTimeout string `json:"startToFireTimeout"`
	TimerID            string `json:"timerId"`
}

// RespondDecisionTaskCompletedRequest models the swf json protocol.
type RespondDecisionTaskCompletedRequest struct {
	Decisions        []Decision `json:"decisions"`
	ExecutionContext string     `json:"executionContext"`
	TaskToken        string     `json:"taskToken"`
}

/*ActivityWorkerProtocol*/

// PollForActivityTaskRequest models the swf json protocol.
type PollForActivityTaskRequest struct {
	Domain   string   `json:"domain"`
	Identity string   `json:"identity,omitempty"`
	TaskList TaskList `json:"taskList"`
}

// PollForActivityTaskResponse models the swf json protocol.
type PollForActivityTaskResponse struct {
	ActivityID        string            `json:"activityId"`
	ActivityType      ActivityType      `json:"activityType"`
	Input             string            `json:"input"`
	StartedEventID    int               `json:"startedEventId"`
	TaskToken         string            `json:"taskToken"`
	WorkflowExecution WorkflowExecution `json:"workflowExecution"`
}

// RespondActivityTaskCompletedRequest models the swf json protocol.
type RespondActivityTaskCompletedRequest struct {
	Result    string `json:"result,omitempty"`
	TaskToken string `json:"taskToken"`
}

// RespondActivityTaskFailedRequest models the swf json protocol.
type RespondActivityTaskFailedRequest struct {
	Details   string `json:"details,omitempty"`
	Reason    string `json:"reason,omitempty"`
	TaskToken string `json:"taskToken"`
}

// RespondActivityTaskCanceledRequest models the swf json protocol.
type RespondActivityTaskCanceledRequest struct {
	Details   string `json:"details,omitempty"`
	TaskToken string `json:"taskToken"`
}

// RecordActivityTaskHeartbeatRequest models the swf json protocol.
type RecordActivityTaskHeartbeatRequest struct {
	Details   string `json:"details,omitempty"`
	TaskToken string `json:"taskToken"`
}

// RecordActivityTaskHeartbeatResponse models the swf json protocol.
type RecordActivityTaskHeartbeatResponse struct {
	CancelRequested bool `json:"cancelRequested"`
}

// DeprecateActivityType models the swf json protocol.
type DeprecateActivityType struct {
	ActivityType ActivityType `json:"activityType"`
	Domain       string       `json:"domain"`
}

// DeprecateWorkflowType models the swf json protocol.
type DeprecateWorkflowType struct {
	Domain       string       `json:"domain"`
	WorkflowType WorkflowType `json:"workflowType"`
}

// DeprecateDomain models the swf json protocol.
type DeprecateDomain struct {
	Name string `json:"name"`
}

// RegisterActivityType models the swf json protocol.
type RegisterActivityType struct {
	DefaultTaskHeartbeatTimeout       string    `json:"defaultTaskHeartbeatTimeout,omitempty"`
	DefaultTaskList                   *TaskList `json:"defaultTaskList,omitempty"`
	DefaultTaskScheduleToCloseTimeout string    `json:"defaultTaskScheduleToCloseTimeout,omitempty"`
	DefaultTaskScheduleToStartTimeout string    `json:"defaultTaskScheduleToStartTimeout,omitempty"`
	DefaultTaskStartToCloseTimeout    string    `json:"defaultTaskStartToCloseTimeout,omitempty"`
	Description                       string    `json:"description,omitempty"`
	Domain                            string    `json:"domain"`
	Name                              string    `json:"name"`
	Version                           string    `json:"version"`
}

// RegisterDomain models the swf json protocol.
type RegisterDomain struct {
	Description                            string `json:"description,omitempty"`
	Name                                   string `json:"name"`
	WorkflowExecutionRetentionPeriodInDays string `json:"workflowExecutionRetentionPeriodInDays"`
}

// RegisterWorkflowType models the swf json protocol.
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

// CountClosedWorkflowExecutionsRequest models the swf json protocol.
type CountClosedWorkflowExecutionsRequest struct {
	CloseStatusFilter *StatusFilter    `json:"closeStatusFilter,omitempty"`
	CloseTimeFilter   *TimeFilter      `json:"closeTimeFilter,omitempty"`
	Domain            string           `json:"domain"`
	ExecutionFilter   *ExecutionFilter `json:"executionFilter,omitempty"`
	StartTimeFilter   *TimeFilter      `json:"startTimeFilter,omitempty"`
	TagFilter         *TagFilter       `json:"tagFilter,omitempty"`
	TypeFilter        *TypeFilter      `json:"typeFilter,omitempty"`
}

// CountOpenWorkflowExecutionsRequest models the swf json protocol.
type CountOpenWorkflowExecutionsRequest struct {
	Domain          string           `json:"domain"`
	ExecutionFilter *ExecutionFilter `json:"executionFilter,omitempty"`
	StartTimeFilter TimeFilter       `json:"startTimeFilter,omitempty"`
	TagFilter       *TagFilter       `json:"tagFilter,omitempty"`
	TypeFilter      *TypeFilter      `json:"typeFilter,omitempty"`
}

// CountPendingActivityTasksRequest models the swf json protocol.
type CountPendingActivityTasksRequest struct {
	Domain   string   `json:"domain"`
	TaskList TaskList `json:"taskList"`
}

// CountPendingDecisionTasksRequest models the swf json protocol.
type CountPendingDecisionTasksRequest struct {
	Domain   string   `json:"domain"`
	TaskList TaskList `json:"taskList"`
}

// DescribeActivityTypeRequest models the swf json protocol.
type DescribeActivityTypeRequest struct {
	ActivityType ActivityType `json:"activityType"`
	Domain       string       `json:"domain"`
}

// DescribeActivityTypeResponse models the swf json protocol.
type DescribeActivityTypeResponse struct {
	Configuration ActivityTypeConfiguration `json:"configuration"`
	TypeInfo      ActivityTypeInfo          `json:"typeInfo"`
}

// ActivityTypeConfiguration models the swf json protocol.
type ActivityTypeConfiguration struct {
	DefaultTaskHeartbeatTimeout       string   `json:"defaultTaskHeartbeatTimeout"`
	DefaultTaskList                   TaskList `json:"defaultTaskList"`
	DefaultTaskScheduleToCloseTimeout string   `json:"defaultTaskScheduleToCloseTimeout"`
	DefaultTaskScheduleToStartTimeout string   `json:"defaultTaskScheduleToStartTimeout"`
	DefaultTaskStartToCloseTimeout    string   `json:"defaultTaskStartToCloseTimeout"`
}

// DescribeDomainRequest models the swf json protocol.
type DescribeDomainRequest struct {
	Name string `json:"name"`
}

// DescribeDomainResponse models the swf json protocol.
type DescribeDomainResponse struct {
	Configuration DomainConfiguration `json:"configuration"`
	DomainInfo    DomainInfo          `json:"domainInfo"`
}

// DomainConfiguration models the swf json protocol.
type DomainConfiguration struct {
	WorkflowExecutionRetentionPeriodInDays string `json:"workflowExecutionRetentionPeriodInDays"`
}

// DomainInfo models the swf json protocol.
type DomainInfo struct {
	Description string `json:"description"`
	Name        string `json:"name"`
	Status      string `json:"status"`
}

// DescribeWorkflowExecutionRequest models the swf json protocol.
type DescribeWorkflowExecutionRequest struct {
	Domain    string            `json:"domain"`
	Execution WorkflowExecution `json:"execution"`
}

// DescribeWorkflowExecutionResponse models the swf json protocol.
type DescribeWorkflowExecutionResponse struct {
	ExecutionConfiguration      ExecutionConfiguration `json:"executionConfiguration"`
	ExecutionInfo               WorkflowExecutionInfo  `json:"executionInfo"`
	LatestActivityTaskTimestamp *Date                  `json:"latestActivityTaskTimestamp"`
	LatestExecutionContext      string                 `json:"latestExecutionContext"`
	OpenCounts                  OpenCounts             `json:"openCounts"`
}

// ExecutionConfiguration models the swf json protocol.
type ExecutionConfiguration struct {
	ChildPolicy                  string   `json:"childPolicy"`
	ExecutionStartToCloseTimeout string   `json:"executionStartToCloseTimeout"`
	TaskList                     TaskList `json:"taskList"`
	TaskStartToCloseTimeout      string   `json:"taskStartToCloseTimeout"`
}

// OpenCounts models the swf json protocol.
type OpenCounts struct {
	OpenActivityTasks           string `json:"openActivityTasks"`
	OpenChildWorkflowExecutions string `json:"openChildWorkflowExecutions"`
	OpenDecisionTasks           string `json:"openDecisionTasks"`
	OpenTimers                  string `json:"openTimers"`
}

// DescribeWorkflowTypeRequest models the swf json protocol.
type DescribeWorkflowTypeRequest struct {
	Domain       string       `json:"domain"`
	WorkflowType WorkflowType `json:"workflowType"`
}

// DescribeWorkflowTypeResponse models the swf json protocol.
type DescribeWorkflowTypeResponse struct {
	Configuration WorkflowConfiguration `json:"configuration"`
	TypeInfo      WorkflowTypeInfo      `json:"typeInfo"`
}

// WorkflowConfiguration models the swf json protocol.
type WorkflowConfiguration struct {
	DefaultChildPolicy                  string   `json:"defaultChildPolicy"`
	DefaultExecutionStartToCloseTimeout string   `json:"defaultExecutionStartToCloseTimeout"`
	DefaultTaskList                     TaskList `json:"defaultTaskList"`
	DefaultTaskStartToCloseTimeout      string   `json:"defaultTaskStartToCloseTimeout"`
}

// GetWorkflowExecutionHistoryRequest models the swf json protocol.
type GetWorkflowExecutionHistoryRequest struct {
	Domain          string            `json:"domain"`
	Execution       WorkflowExecution `json:"execution"`
	MaximumPageSize int               `json:"maximumPageSize,omitempty"`
	NextPageToken   string            `json:"nextPageToken,omitempty"`
	ReverseOrder    bool              `json:"reverseOrder,omitempty"`
}

// GetWorkflowExecutionHistoryResponse models the swf json protocol.
type GetWorkflowExecutionHistoryResponse struct {
	Events        []HistoryEvent `json:"events"`
	NextPageToken string         `json:"nextPageToken,omitempty"`
}

// ListActivityTypesRequest models the swf json protocol.
type ListActivityTypesRequest struct {
	Domain             string `json:"domain"`
	MaximumPageSize    int    `json:"maximumPageSize,omitempty"`
	Name               string `json:"name,omitempty"`
	NextPageToken      string `json:"nextPageToken,omitempty"`
	RegistrationStatus string `json:"registrationStatus"`
	ReverseOrder       bool   `json:"reverseOrder,omitempty"`
}

// ListActivityTypesResponse models the swf json protocol.
type ListActivityTypesResponse struct {
	NextPageToken *string            `json:"nextPageToken"`
	TypeInfos     []ActivityTypeInfo `json:"typeInfos"`
}

// ListClosedWorkflowExecutionsRequest models the swf json protocol.
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

// ListClosedWorkflowExecutionsResponse models the swf json protocol.
type ListClosedWorkflowExecutionsResponse struct {
	ExecutionInfos []WorkflowExecutionInfo `json:"executionInfos"`
	NextPageToken  string                  `json:"nextPageToken,omitempty"`
}

// ListDomainsRequest models the swf json protocol.
type ListDomainsRequest struct {
	MaximumPageSize    int    `json:"maximumPageSize,omitempty"`
	NextPageToken      string `json:"nextPageToken,omitempty"`
	RegistrationStatus string `json:"registrationStatus"`
	ReverseOrder       bool   `json:"reverseOrder,omitempty"`
}

// ListDomainsResponse models the swf json protocol.
type ListDomainsResponse struct {
	DomainInfos   []DomainInfo `json:"domainInfos"`
	NextPageToken string       `json:"nextPageToken,omitempty"`
}

// ListOpenWorkflowExecutionsRequest models the swf json protocol.
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

// ListOpenWorkflowExecutionsResponse models the swf json protocol.
type ListOpenWorkflowExecutionsResponse struct {
	ExecutionInfos []WorkflowExecutionInfo `json:"executionInfos"`
	NextPageToken  string                  `json:"nextPageToken,omitempty"`
}

// WorkflowExecutionInfo models the swf json protocol.
type WorkflowExecutionInfo struct {
	CancelRequested bool              `json:"cancelRequested"`
	CloseStatus     string            `json:"closeStatus"`
	CloseTimestamp  *Date             `json:"closeTimestamp"`
	Execution       WorkflowExecution `json:"execution"`
	ExecutionStatus string            `json:"executionStatus"`
	Parent          WorkflowExecution `json:"parent"`
	StartTimestamp  *Date             `json:"startTimestamp"`
	TagList         []string          `json:"tagList"`
	WorkflowType    WorkflowType      `json:"workflowType"`
}

// StatusFilter models the swf json protocol.
type StatusFilter struct {
	Status string `json:"status"`
}

// CountResponse models the swf json protocol.
type CountResponse struct {
	Count     int  `json:"count"`
	Truncated bool `json:"truncated"`
}

// TimeFilter models the swf json protocol.
type TimeFilter struct {
	LatestDate *Date `json:"latestDate,omitempty"`
	OldestDate *Date `json:"oldestDate"`
}

// ZeroTimeFilter returns a TimeFiter with the OldestDate set to 0 and the LatestDate nil
func ZeroTimeFilter() *TimeFilter {
	return &TimeFilter{
		OldestDate: &Date{time.Unix(0, 0)},
	}
}

// ExecutionFilter models the swf json protocol.
type ExecutionFilter struct {
	WorkflowID string `json:"workflowId"`
}

// TagFilter models the swf json protocol.
type TagFilter struct {
	Tag string `json:"tag"`
}

// TypeFilter models the swf json protocol.
type TypeFilter struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

/*common types*/

// TaskList models the swf json protocol.
type TaskList struct {
	Name string `json:"name"`
}

// WorkflowType models the swf json protocol.
type WorkflowType struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// WorkflowExecution models the swf json protocol.
type WorkflowExecution struct {
	RunID      string `json:"runId"`
	WorkflowID string `json:"workflowId"`
}

// ActivityType models the swf json protocol.
type ActivityType struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// WorkflowTypeInfo models the swf json protocol.
type WorkflowTypeInfo struct {
	CreationDate    *Date        `json:"creationDate"`
	DeprecationDate *Date        `json:"deprecationDate"`
	Description     string       `json:"description"`
	Status          string       `json:"status"`
	WorkflowType    WorkflowType `json:"workflowType"`
}

// ActivityTypeInfo models the swf json protocol.
type ActivityTypeInfo struct {
	CreationDate    *Date        `json:"creationDate"`
	DeprecationDate *Date        `json:"deprecationDate"`
	Description     string       `json:"description"`
	Status          string       `json:"status"`
	ActivityType    ActivityType `json:"activityType"`
}

// PutRecordRequest models the kinesis json protocol.
type PutRecordRequest struct {
	Data                      []byte
	ExplicitHashKey           string `json:"ExplicitHashKey,omitempty"`
	PartitionKey              string
	SequenceNumberForOrdering string `json:"SequenceNumberForOrdering,omitempty"`
	StreamName                string
}

// PutRecordResponse models the kinesis json protocol.
type PutRecordResponse struct {
	SequenceNumber string
	ShardID        string `json:"ShardId"`
}

// GetRecordsRequest models the kinesis json protocol.
type GetRecordsRequest struct {
	Limit         int `json:"Limit,omitempty"`
	ShardIterator string
}

// GetRecordsResponse models the kinesis json protocol.
type GetRecordsResponse struct {
	NextShardIterator string
	Records           []struct {
		Data           []byte
		PartitionKey   string
		SequenceNumber string
	}
}

// GetShardIteratorRequest models the kinesis json protocol.
type GetShardIteratorRequest struct {
	StreamName             string
	ShardID                string `json:"ShardId"`
	ShardIteratorType      string
	StartingSequenceNumber string `json:"StartingSequenceNumber,omitempty"`
}

// GetShardIteratorResponse models the kinesis json protocol.
type GetShardIteratorResponse struct {
	ShardIterator string
}

// CreateStream models the kinesis json protocol.
type CreateStream struct {
	ShardCount int
	StreamName string
}

// DescribeStreamRequest models the kinesis json protocol.
type DescribeStreamRequest struct {
	ExclusiveStartShardID *string `json:"ExclusiveStartShardId"`
	Limit                 *int
	StreamName            string
}

// DescribeStreamResponse models the kinesis json protocol.
type DescribeStreamResponse struct {
	StreamDescription struct {
		HasMoreShards bool
		Shards        []struct {
			AdjacentParentShardID string `json:"AdjacentParentShardId"`
			HashKeyRange          struct {
				EndingHashKey   string
				StartingHashKey string
			}
			ParentShardID       string `json:"ParentShardId"`
			SequenceNumberRange struct {
				EndingSequenceNumber   string
				StartingSequenceNumber string
			}
			ShardID string `json:"ShardId"`
		}
		StreamARN    string
		StreamName   string
		StreamStatus string
	}
}

//Date is a wrapper struct around time.Time that provides json de/serialization in swf's expected format.
type Date struct{ time.Time }

//UnmarshalJSON parses the swf representation of time.
func (s *Date) UnmarshalJSON(b []byte) error {
	timestamp, err := strconv.ParseFloat(string(b), 64)
	if err != nil {
		return err
	}
	s.Time = time.Unix(int64(timestamp), 0)
	return nil
}

//MarshalJSON formats a time.Time into swf's expected format.
func (s *Date) MarshalJSON() ([]byte, error) {
	timestamp := s.Time.Unix()
	return []byte(strconv.FormatFloat(float64(timestamp), 'g', -1, 64)), nil
}
