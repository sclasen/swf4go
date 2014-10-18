package swf

type StartWorkflowRequest struct {
	ChildPolicy                  string       `json:"childPolicy"`
	Domain                       string       `json:"domain"`
	ExecutionStartToCloseTimeout string       `json:"executionStartToCloseTimeout"`
	Input                        string       `json:"input"`
	TagList                      []string     `json:"tagList"`
	TaskList                     TaskList     `json:"taskList"`
	TaskStartToCloseTimeout      string       `json:"taskStartToCloseTimeout"`
	WorkflowId                   string       `json:"workflowId"`
	WorkflowType                 WorkflowType `json:"workflowType"`
}

type StartWorkflowResponse struct {
	RunId string `json:"runId"`
}

type PollForDecisionTaskRequest struct {
	Domain          string   `json:"domain"`
	Identity        string   `json:"identity"`
	MaximumPageSize string   `json:"maximumPageSize"`
	NextPageToken   string   `json:"nextPageToken"`
	ReverseOrder    string   `json:"reverseOrder"`
	TaskList        TaskList `json:"taskList"`
}

type HistoryEvent struct {
	ActivityTaskCancelRequestedEventAttributes struct {
		ActivityId                   string `json:"activityId"`
		DecisionTaskCompletedEventId string `json:"decisionTaskCompletedEventId"`
	} `json:"activityTaskCancelRequestedEventAttributes"`
	ActivityTaskCanceledEventAttributes struct {
		Details                      string `json:"details"`
		LatestCancelRequestedEventId string `json:"latestCancelRequestedEventId"`
		ScheduledEventId             string `json:"scheduledEventId"`
		StartedEventId               string `json:"startedEventId"`
	} `json:"activityTaskCanceledEventAttributes"`
	ActivityTaskCompletedEventAttributes struct {
		Result           string `json:"result"`
		ScheduledEventId string `json:"scheduledEventId"`
		StartedEventId   string `json:"startedEventId"`
	} `json:"activityTaskCompletedEventAttributes"`
	ActivityTaskFailedEventAttributes struct {
		Details          string `json:"details"`
		Reason           string `json:"reason"`
		ScheduledEventId string `json:"scheduledEventId"`
		StartedEventId   string `json:"startedEventId"`
	} `json:"activityTaskFailedEventAttributes"`
	ActivityTaskScheduledEventAttributes struct {
		ActivityId                   string `json:"activityId"`
		ActivityType struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"activityType"`
		Control                      string `json:"control"`
		DecisionTaskCompletedEventId string `json:"decisionTaskCompletedEventId"`
		HeartbeatTimeout             string `json:"heartbeatTimeout"`
		Input                        string `json:"input"`
		ScheduleToCloseTimeout       string `json:"scheduleToCloseTimeout"`
		ScheduleToStartTimeout       string `json:"scheduleToStartTimeout"`
		StartToCloseTimeout          string `json:"startToCloseTimeout"`
		TaskList                     struct {
			Name string `json:"name"`
		} `json:"taskList"`
	} `json:"activityTaskScheduledEventAttributes"`
	ActivityTaskStartedEventAttributes struct {
		Identity         string `json:"identity"`
		ScheduledEventId string `json:"scheduledEventId"`
	} `json:"activityTaskStartedEventAttributes"`
	ActivityTaskTimedOutEventAttributes struct {
		Details          string `json:"details"`
		ScheduledEventId string `json:"scheduledEventId"`
		StartedEventId   string `json:"startedEventId"`
		TimeoutType      string `json:"timeoutType"`
	} `json:"activityTaskTimedOutEventAttributes"`
	CancelTimerFailedEventAttributes struct {
		Cause                        string `json:"cause"`
		DecisionTaskCompletedEventId string `json:"decisionTaskCompletedEventId"`
		TimerId                      string `json:"timerId"`
	} `json:"cancelTimerFailedEventAttributes"`
	CancelWorkflowExecutionFailedEventAttributes struct {
		Cause                        string `json:"cause"`
		DecisionTaskCompletedEventId string `json:"decisionTaskCompletedEventId"`
	} `json:"cancelWorkflowExecutionFailedEventAttributes"`
	ChildWorkflowExecutionCanceledEventAttributes struct {
		Details           string `json:"details"`
		InitiatedEventId  string `json:"initiatedEventId"`
		StartedEventId    string `json:"startedEventId"`
		WorkflowExecution struct {
			RunId      string `json:"runId"`
			WorkflowId string `json:"workflowId"`
		} `json:"workflowExecution"`
		WorkflowType struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"workflowType"`
	} `json:"childWorkflowExecutionCanceledEventAttributes"`
	ChildWorkflowExecutionCompletedEventAttributes struct {
		InitiatedEventId  string `json:"initiatedEventId"`
		Result            string `json:"result"`
		StartedEventId    string `json:"startedEventId"`
		WorkflowExecution struct {
			RunId      string `json:"runId"`
			WorkflowId string `json:"workflowId"`
		} `json:"workflowExecution"`
		WorkflowType struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"workflowType"`
	} `json:"childWorkflowExecutionCompletedEventAttributes"`
	ChildWorkflowExecutionFailedEventAttributes struct {
		Details           string `json:"details"`
		InitiatedEventId  string `json:"initiatedEventId"`
		Reason            string `json:"reason"`
		StartedEventId    string `json:"startedEventId"`
		WorkflowExecution struct {
			RunId      string `json:"runId"`
			WorkflowId string `json:"workflowId"`
		} `json:"workflowExecution"`
		WorkflowType struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"workflowType"`
	} `json:"childWorkflowExecutionFailedEventAttributes"`
	ChildWorkflowExecutionStartedEventAttributes struct {
		InitiatedEventId  string `json:"initiatedEventId"`
		WorkflowExecution struct {
			RunId      string `json:"runId"`
			WorkflowId string `json:"workflowId"`
		} `json:"workflowExecution"`
		WorkflowType struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"workflowType"`
	} `json:"childWorkflowExecutionStartedEventAttributes"`
	ChildWorkflowExecutionTerminatedEventAttributes struct {
		InitiatedEventId  string `json:"initiatedEventId"`
		StartedEventId    string `json:"startedEventId"`
		WorkflowExecution struct {
			RunId      string `json:"runId"`
			WorkflowId string `json:"workflowId"`
		} `json:"workflowExecution"`
		WorkflowType struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"workflowType"`
	} `json:"childWorkflowExecutionTerminatedEventAttributes"`
	ChildWorkflowExecutionTimedOutEventAttributes struct {
		InitiatedEventId  string `json:"initiatedEventId"`
		StartedEventId    string `json:"startedEventId"`
		TimeoutType       string `json:"timeoutType"`
		WorkflowExecution struct {
			RunId      string `json:"runId"`
			WorkflowId string `json:"workflowId"`
		} `json:"workflowExecution"`
		WorkflowType struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"workflowType"`
	} `json:"childWorkflowExecutionTimedOutEventAttributes"`
	CompleteWorkflowExecutionFailedEventAttributes struct {
		Cause                        string `json:"cause"`
		DecisionTaskCompletedEventId string `json:"decisionTaskCompletedEventId"`
	} `json:"completeWorkflowExecutionFailedEventAttributes"`
	ContinueAsNewWorkflowExecutionFailedEventAttributes struct {
		Cause                        string `json:"cause"`
		DecisionTaskCompletedEventId string `json:"decisionTaskCompletedEventId"`
	} `json:"continueAsNewWorkflowExecutionFailedEventAttributes"`
	DecisionTaskCompletedEventAttributes struct {
		ExecutionContext string `json:"executionContext"`
		ScheduledEventId string `json:"scheduledEventId"`
		StartedEventId   string `json:"startedEventId"`
	} `json:"decisionTaskCompletedEventAttributes"`
	DecisionTaskScheduledEventAttributes struct {
		StartToCloseTimeout string `json:"startToCloseTimeout"`
		TaskList            struct {
			Name string `json:"name"`
		} `json:"taskList"`
	} `json:"decisionTaskScheduledEventAttributes"`
	DecisionTaskStartedEventAttributes struct {
		Identity         string `json:"identity"`
		ScheduledEventId string `json:"scheduledEventId"`
	} `json:"decisionTaskStartedEventAttributes"`
	DecisionTaskTimedOutEventAttributes struct {
		ScheduledEventId string `json:"scheduledEventId"`
		StartedEventId   string `json:"startedEventId"`
		TimeoutType      string `json:"timeoutType"`
	} `json:"decisionTaskTimedOutEventAttributes"`
	EventId                                                 string `json:"eventId"`
	EventTimestamp                                          string `json:"eventTimestamp"`
	EventType                                               string `json:"eventType"`
	ExternalWorkflowExecutionCancelRequestedEventAttributes struct {
		InitiatedEventId  string `json:"initiatedEventId"`
		WorkflowExecution struct {
			RunId      string `json:"runId"`
			WorkflowId string `json:"workflowId"`
		} `json:"workflowExecution"`
	} `json:"externalWorkflowExecutionCancelRequestedEventAttributes"`
	ExternalWorkflowExecutionSignaledEventAttributes struct {
		InitiatedEventId  string `json:"initiatedEventId"`
		WorkflowExecution struct {
			RunId      string `json:"runId"`
			WorkflowId string `json:"workflowId"`
		} `json:"workflowExecution"`
	} `json:"externalWorkflowExecutionSignaledEventAttributes"`
	FailWorkflowExecutionFailedEventAttributes struct {
		Cause                        string `json:"cause"`
		DecisionTaskCompletedEventId string `json:"decisionTaskCompletedEventId"`
	} `json:"failWorkflowExecutionFailedEventAttributes"`
	MarkerRecordedEventAttributes struct {
		DecisionTaskCompletedEventId string `json:"decisionTaskCompletedEventId"`
		Details                      string `json:"details"`
		MarkerName                   string `json:"markerName"`
	} `json:"markerRecordedEventAttributes"`
	RecordMarkerFailedEventAttributes struct {
		Cause                        string `json:"cause"`
		DecisionTaskCompletedEventId string `json:"decisionTaskCompletedEventId"`
		MarkerName                   string `json:"markerName"`
	} `json:"recordMarkerFailedEventAttributes"`
	RequestCancelActivityTaskFailedEventAttributes struct {
		ActivityId                   string `json:"activityId"`
		Cause                        string `json:"cause"`
		DecisionTaskCompletedEventId string `json:"decisionTaskCompletedEventId"`
	} `json:"requestCancelActivityTaskFailedEventAttributes"`
	RequestCancelExternalWorkflowExecutionFailedEventAttributes struct {
		Cause                        string `json:"cause"`
		Control                      string `json:"control"`
		DecisionTaskCompletedEventId string `json:"decisionTaskCompletedEventId"`
		InitiatedEventId             string `json:"initiatedEventId"`
		RunId                        string `json:"runId"`
		WorkflowId                   string `json:"workflowId"`
	} `json:"requestCancelExternalWorkflowExecutionFailedEventAttributes"`
	RequestCancelExternalWorkflowExecutionInitiatedEventAttributes struct {
		Control                      string `json:"control"`
		DecisionTaskCompletedEventId string `json:"decisionTaskCompletedEventId"`
		RunId                        string `json:"runId"`
		WorkflowId                   string `json:"workflowId"`
	} `json:"requestCancelExternalWorkflowExecutionInitiatedEventAttributes"`
	ScheduleActivityTaskFailedEventAttributes struct {
		ActivityId                   string `json:"activityId"`
		ActivityType struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"activityType"`
		Cause                        string `json:"cause"`
		DecisionTaskCompletedEventId string `json:"decisionTaskCompletedEventId"`
	} `json:"scheduleActivityTaskFailedEventAttributes"`
	SignalExternalWorkflowExecutionFailedEventAttributes struct {
		Cause                        string `json:"cause"`
		Control                      string `json:"control"`
		DecisionTaskCompletedEventId string `json:"decisionTaskCompletedEventId"`
		InitiatedEventId             string `json:"initiatedEventId"`
		RunId                        string `json:"runId"`
		WorkflowId                   string `json:"workflowId"`
	} `json:"signalExternalWorkflowExecutionFailedEventAttributes"`
	SignalExternalWorkflowExecutionInitiatedEventAttributes struct {
		Control                      string `json:"control"`
		DecisionTaskCompletedEventId string `json:"decisionTaskCompletedEventId"`
		Input                        string `json:"input"`
		RunId                        string `json:"runId"`
		SignalName                   string `json:"signalName"`
		WorkflowId                   string `json:"workflowId"`
	} `json:"signalExternalWorkflowExecutionInitiatedEventAttributes"`
	StartChildWorkflowExecutionFailedEventAttributes struct {
		Cause                        string `json:"cause"`
		Control                      string `json:"control"`
		DecisionTaskCompletedEventId string `json:"decisionTaskCompletedEventId"`
		InitiatedEventId             string `json:"initiatedEventId"`
		WorkflowId                   string `json:"workflowId"`
		WorkflowType                 struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"workflowType"`
	} `json:"startChildWorkflowExecutionFailedEventAttributes"`
	StartChildWorkflowExecutionInitiatedEventAttributes struct {
		ChildPolicy                  string   `json:"childPolicy"`
		Control                      string   `json:"control"`
		DecisionTaskCompletedEventId string   `json:"decisionTaskCompletedEventId"`
		ExecutionStartToCloseTimeout string   `json:"executionStartToCloseTimeout"`
		Input                        string   `json:"input"`
		TagList                      []string `json:"tagList"`
		TaskList                     struct {
			Name string `json:"name"`
		} `json:"taskList"`
		TaskStartToCloseTimeout      string `json:"taskStartToCloseTimeout"`
		WorkflowId                   string `json:"workflowId"`
		WorkflowType            struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"workflowType"`
	} `json:"startChildWorkflowExecutionInitiatedEventAttributes"`
	StartTimerFailedEventAttributes struct {
		Cause                        string `json:"cause"`
		DecisionTaskCompletedEventId string `json:"decisionTaskCompletedEventId"`
		TimerId                      string `json:"timerId"`
	} `json:"startTimerFailedEventAttributes"`
	TimerCanceledEventAttributes struct {
		DecisionTaskCompletedEventId string `json:"decisionTaskCompletedEventId"`
		StartedEventId               string `json:"startedEventId"`
		TimerId                      string `json:"timerId"`
	} `json:"timerCanceledEventAttributes"`
	TimerFiredEventAttributes struct {
		StartedEventId string `json:"startedEventId"`
		TimerId        string `json:"timerId"`
	} `json:"timerFiredEventAttributes"`
	TimerStartedEventAttributes struct {
		Control                      string `json:"control"`
		DecisionTaskCompletedEventId string `json:"decisionTaskCompletedEventId"`
		StartToFireTimeout           string `json:"startToFireTimeout"`
		TimerId                      string `json:"timerId"`
	} `json:"timerStartedEventAttributes"`
	WorkflowExecutionCancelRequestedEventAttributes struct {
		Cause                     string `json:"cause"`
		ExternalInitiatedEventId  string `json:"externalInitiatedEventId"`
		ExternalWorkflowExecution struct {
			RunId      string `json:"runId"`
			WorkflowId string `json:"workflowId"`
		} `json:"externalWorkflowExecution"`
	} `json:"workflowExecutionCancelRequestedEventAttributes"`
	WorkflowExecutionCanceledEventAttributes struct {
		DecisionTaskCompletedEventId string `json:"decisionTaskCompletedEventId"`
		Details                      string `json:"details"`
	} `json:"workflowExecutionCanceledEventAttributes"`
	WorkflowExecutionCompletedEventAttributes struct {
		DecisionTaskCompletedEventId string `json:"decisionTaskCompletedEventId"`
		Result                       string `json:"result"`
	} `json:"workflowExecutionCompletedEventAttributes"`
	WorkflowExecutionContinuedAsNewEventAttributes struct {
		ChildPolicy                  string   `json:"childPolicy"`
		DecisionTaskCompletedEventId string   `json:"decisionTaskCompletedEventId"`
		ExecutionStartToCloseTimeout string   `json:"executionStartToCloseTimeout"`
		Input                        string   `json:"input"`
		NewExecutionRunId            string   `json:"newExecutionRunId"`
		TagList                      []string `json:"tagList"`
		TaskList                     struct {
			Name string `json:"name"`
		} `json:"taskList"`
		TaskStartToCloseTimeout      string `json:"taskStartToCloseTimeout"`
		WorkflowType            struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"workflowType"`
	} `json:"workflowExecutionContinuedAsNewEventAttributes"`
	WorkflowExecutionFailedEventAttributes struct {
		DecisionTaskCompletedEventId string `json:"decisionTaskCompletedEventId"`
		Details                      string `json:"details"`
		Reason                       string `json:"reason"`
	} `json:"workflowExecutionFailedEventAttributes"`
	WorkflowExecutionSignaledEventAttributes struct {
		ExternalInitiatedEventId  string `json:"externalInitiatedEventId"`
		ExternalWorkflowExecution struct {
			RunId      string `json:"runId"`
			WorkflowId string `json:"workflowId"`
		} `json:"externalWorkflowExecution"`
		Input                     string `json:"input"`
		SignalName                string `json:"signalName"`
	} `json:"workflowExecutionSignaledEventAttributes"`
	WorkflowExecutionStartedEventAttributes struct {
		ChildPolicy                  string `json:"childPolicy"`
		ContinuedExecutionRunId      string `json:"continuedExecutionRunId"`
		ExecutionStartToCloseTimeout string `json:"executionStartToCloseTimeout"`
		Input                        string `json:"input"`
		ParentInitiatedEventId       string `json:"parentInitiatedEventId"`
		ParentWorkflowExecution      struct {
			RunId      string `json:"runId"`
			WorkflowId string `json:"workflowId"`
		} `json:"parentWorkflowExecution"`
		TagList                      []string `json:"tagList"`
		TaskList struct {
			Name string `json:"name"`
		} `json:"taskList"`
		TaskStartToCloseTimeout      string `json:"taskStartToCloseTimeout"`
		WorkflowType            struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"workflowType"`
	} `json:"workflowExecutionStartedEventAttributes"`
	WorkflowExecutionTerminatedEventAttributes struct {
		Cause       string `json:"cause"`
		ChildPolicy string `json:"childPolicy"`
		Details     string `json:"details"`
		Reason      string `json:"reason"`
	} `json:"workflowExecutionTerminatedEventAttributes"`
	WorkflowExecutionTimedOutEventAttributes struct {
		ChildPolicy string `json:"childPolicy"`
		TimeoutType string `json:"timeoutType"`
	} `json:"workflowExecutionTimedOutEventAttributes"`
}

type PollForDecisionTaskResponse struct {
	Events                 []HistoryEvent    `json:"events"`
	NextPageToken          string            `json:"nextPageToken"`
	PreviousStartedEventId string            `json:"previousStartedEventId"`
	StartedEventId         string            `json:"startedEventId"`
	TaskToken              string            `json:"taskToken"`
	WorkflowExecution      WorkflowExecution `json:"workflowExecution"`
	WorkflowType           WorkflowType      `json:"workflowType"`
}

type Decision struct {
	CancelTimerDecisionAttributes struct {
		TimerId string `json:"timerId"`
	} `json:"cancelTimerDecisionAttributes"`
	CancelWorkflowExecutionDecisionAttributes struct {
		Details string `json:"details"`
	} `json:"cancelWorkflowExecutionDecisionAttributes"`
	CompleteWorkflowExecutionDecisionAttributes struct {
		Result string `json:"result"`
	} `json:"completeWorkflowExecutionDecisionAttributes"`
	ContinueAsNewWorkflowExecutionDecisionAttributes struct {
		ChildPolicy                  string   `json:"childPolicy"`
		ExecutionStartToCloseTimeout string   `json:"executionStartToCloseTimeout"`
		Input                        string   `json:"input"`
		TagList                      []string `json:"tagList"`
		TaskList                     TaskList `json:"taskList"`
		TaskStartToCloseTimeout      string   `json:"taskStartToCloseTimeout"`
		WorkflowTypeVersion          string   `json:"workflowTypeVersion"`
	} `json:"continueAsNewWorkflowExecutionDecisionAttributes"`
	DecisionType                            string `json:"decisionType"`
	FailWorkflowExecutionDecisionAttributes struct {
		Details string `json:"details"`
		Reason  string `json:"reason"`
	} `json:"failWorkflowExecutionDecisionAttributes"`
	RecordMarkerDecisionAttributes struct {
		Details    string `json:"details"`
		MarkerName string `json:"markerName"`
	} `json:"recordMarkerDecisionAttributes"`
	RequestCancelActivityTaskDecisionAttributes struct {
		ActivityId string `json:"activityId"`
	} `json:"requestCancelActivityTaskDecisionAttributes"`
	RequestCancelExternalWorkflowExecutionDecisionAttributes struct {
		Control    string `json:"control"`
		RunId      string `json:"runId"`
		WorkflowId string `json:"workflowId"`
	} `json:"requestCancelExternalWorkflowExecutionDecisionAttributes"`
	ScheduleActivityTaskDecisionAttributes struct {
		ActivityId             string       `json:"activityId"`
		ActivityType           ActivityType `json:"activityType"`
		Control                string       `json:"control"`
		HeartbeatTimeout       string       `json:"heartbeatTimeout"`
		Input                  string       `json:"input"`
		ScheduleToCloseTimeout string       `json:"scheduleToCloseTimeout"`
		ScheduleToStartTimeout string       `json:"scheduleToStartTimeout"`
		StartToCloseTimeout    string       `json:"startToCloseTimeout"`
		TaskList               struct {
			Name string `json:"name"`
		} `json:"taskList"`
	} `json:"scheduleActivityTaskDecisionAttributes"`
	SignalExternalWorkflowExecutionDecisionAttributes struct {
		Control    string `json:"control"`
		Input      string `json:"input"`
		RunId      string `json:"runId"`
		SignalName string `json:"signalName"`
		WorkflowId string `json:"workflowId"`
	} `json:"signalExternalWorkflowExecutionDecisionAttributes"`
	StartChildWorkflowExecutionDecisionAttributes struct {
		ChildPolicy                  string       `json:"childPolicy"`
		Control                      string       `json:"control"`
		ExecutionStartToCloseTimeout string       `json:"executionStartToCloseTimeout"`
		Input                        string       `json:"input"`
		TagList                      []string     `json:"tagList"`
		TaskList                     TaskList     `json:"taskList"`
		TaskStartToCloseTimeout      string       `json:"taskStartToCloseTimeout"`
		WorkflowId                   string       `json:"workflowId"`
		WorkflowType                 WorkflowType `json:"workflowType"`
	} `json:"startChildWorkflowExecutionDecisionAttributes"`
	StartTimerDecisionAttributes struct {
		Control            string `json:"control"`
		StartToFireTimeout string `json:"startToFireTimeout"`
		TimerId            string `json:"timerId"`
	} `json:"startTimerDecisionAttributes"`
}

type RespondDecisionTaskCompletedRequest struct {
	Decisions        []Decision `json:"decisions"`
	ExecutionContext string     `json:"executionContext"`
	TaskToken        string     `json:"taskToken"`
}

type PollForActivityTaskRequest struct {
	Domain   string   `json:"domain"`
	Identity string   `json:"identity"`
	TaskList TaskList `json:"taskList"`
}

type PollForActivityTaskResponse struct {
	ActivityId        string            `json:"activityId"`
	ActivityType      ActivityType      `json:"activityType"`
	Input             string            `json:"input"`
	StartedEventId    string            `json:"startedEventId"`
	TaskToken         string            `json:"taskToken"`
	WorkflowExecution WorkflowExecution `json:"workflowExecution"`
}

type RespondActivityTaskCompletedRequest struct {
	Result    string `json:"result"`
	TaskToken string `json:"taskToken"`
}

type RespondActivityTaskFailedRequest struct {
	Details   string `json:"details"`
	Reason    string `json:"reason"`
	TaskToken string `json:"taskToken"`
}

type RespondActivityTaskCanceledRequest struct {
	Details   string `json:"details"`
	TaskToken string `json:"taskToken"`
}

type SignalWorkflowRequest struct {
	Domain     string `json:"domain"`
	Input      string `json:"input"`
	RunId      string `json:"runId"`
	SignalName string `json:"signalName"`
	WorkflowId string `json:"workflowId"`
}

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
