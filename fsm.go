package swf

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"

	"code.google.com/p/goprotobuf/proto"
)

// constants used as marker names or signal names
const (
	StateMarker       = "FSM.State"
	ErrorSignal       = "FSM.Error"
	SystemErrorSignal = "FSM.SystemError"
)

// Decider decides an Outcome based on an event and the current data for an
// FSM. You can assert the interface{} parameter that is passed to the Decider
// as the type of the DataType field in the FSM. Alternatively, you can use the
// TypedDecider to avoid having to do the assertion.
type Decider func(*FSMContext, HistoryEvent, interface{}) Outcome

// Outcome represents the minimum data needed to be returned by a Decider.
type Outcome interface {
	// Data returns the data for this Outcome.
	Data() interface{}
	// Decisions returns the list of Decisions for this Outcome.
	Decisions() []Decision
	// State returns the state to transition to. An empty string means no
	// transition.
	State() string
}

// TransitionOutcome is an Outcome in which the FSM will transtion to a new state.
type TransitionOutcome struct {
	data      interface{}
	state     string
	decisions []Decision
}

// Data returns the data for this Outcome.
func (t TransitionOutcome) Data() interface{} { return t.data }

// Decisions returns the list of Decisions for this Outcome.
func (t TransitionOutcome) Decisions() []Decision { return t.decisions }

// State returns the next state for this TransitionOutcome.
func (t TransitionOutcome) State() string { return t.state }

// StayOutcome is an Outcome in which the FSM will remain in the same state.
type StayOutcome struct {
	data      interface{}
	decisions []Decision
}

// Data returns the data for this Outcome.
func (s StayOutcome) Data() interface{} { return s.data }

// Decisions returns the list of Decisions for this Outcome.
func (s StayOutcome) Decisions() []Decision { return s.decisions }

// State returns the next state for the StayOutcome, which is always empty.
func (s StayOutcome) State() string { return "" }

// TerminationOutcome can do things like check that the last decision is a termination.
type TerminationOutcome struct {
	data      interface{}
	decisions []Decision
}

// Data returns the data for this Outcome.
func (t TerminationOutcome) Data() interface{} { return t.data }

// Decisions returns the list of Decisions for this Outcome.
func (t TerminationOutcome) Decisions() []Decision { return t.decisions }

// State returns the next state for the TerminationOutcome, which is always
// "TERMINATED".
func (t TerminationOutcome) State() string { return "TERMINATED" }

// ErrorOutcome can be used to purposefully put the workflow into an error state.
type ErrorOutcome struct {
	state     string
	data      interface{}
	decisions []Decision
}

// Data returns the data for this Outcome.
func (e ErrorOutcome) Data() interface{} { return e.data }

// Decisions returns the list of Decisions for this Outcome.
func (e ErrorOutcome) Decisions() []Decision { return e.decisions }

// State returns the next state for the ErrorOutcome, which is always "error".
func (e ErrorOutcome) State() string { return "error" }

type intermediateOutcome struct {
	stateVersion uint64
	state        string
	data         interface{}
	decisions    []Decision
}

// FSMState defines the behavior of one state of an FSM
type FSMState struct {
	// Name is the name of the state. When returning an Outcome, the NextState should match the Name of an FSMState in your FSM.
	Name string
	// Decider decides an Outcome given the current state, data, and an event.
	Decider Decider
}

// FSM models the decision handling logic a workflow in SWF
type FSM struct {
	//Name of the fsm. Used when emitting logs. Should probably be set to the name of the workflow associated with the fsm.
	Name string
	// Domain of the workflow associated with the FSM.
	Domain string
	// TaskList that the underlying poller will poll for decision tasks.
	TaskList string
	// Identity used in PollForDecisionTaskRequests, can be empty.
	Identity string
	// Client used to make SWF api requests.
	Client WorkflowClient
	// DataType of the data struct associated with this FSM.
	// The data is automatically peristed to and loaded from workflow history by the FSM.
	DataType interface{}
	// Serializer used to serialize/deserialise state from workflow history.
	Serializer StateSerializer
	// Kinesis stream in the same region to replicate state to.
	KinesisStream string
	//PollerShutdownManager is used when the FSM is managing the polling
	PollerShutdownManager *PollerShutdownManager
	states                map[string]*FSMState
	initialState          *FSMState
	errorState            *FSMState
	stop                  chan bool
	stopAck               chan bool
	allowPanics           bool //makes testing easier
}

// AddInitialState adds a state to the FSM and uses it as the initial state when a workflow execution is started.
func (f *FSM) AddInitialState(state *FSMState) {
	f.AddState(state)
	f.initialState = state
}

// AddState adds a state to the FSM.
func (f *FSM) AddState(state *FSMState) {
	if f.states == nil {
		f.states = make(map[string]*FSMState)
	}
	f.states[state.Name] = state
}

//ErrorStates should return an outcome with nil Data and "" as NextState if they wish for the normal state recovery mechanism
//to load current state and data.

// AddErrorState adds an error handling state to your FSM. This is a special FSMState that should at a minimum handle
// WorkflowSignaled events where the signal name is FSM.Error and the event data is a SerializedDecisionError, or the
// signal name is FSM.SystemError and the event data is a SerializedSystemError. The error state should take care of transitioning
// the workflow back into a working state, by making decisions, updating data and/or choosing a new state.
func (f *FSM) AddErrorState(state *FSMState) {
	f.AddState(state)
	f.errorState = state
}

// DefaultErrorState is the error state used in an FSM if one has not been set. It simply emits logs admonishing you to add
// a proper error state to your FSM.
func (f *FSM) DefaultErrorState() *FSMState {
	return &FSMState{
		Name: "error",
		Decider: func(fsm *FSMContext, h HistoryEvent, data interface{}) Outcome {
			switch h.EventType {
			case EventTypeWorkflowExecutionSignaled:
				{
					switch h.WorkflowExecutionSignaledEventAttributes.SignalName {
					case ErrorSignal:
						err := &SerializedDecisionError{}
						f.Deserialize(h.WorkflowExecutionSignaledEventAttributes.Input, err)
						f.log("action=default-handle-error at=handle-decision-error error=%+v", err)
						f.log("YOU SHOULD CREATE AN ERROR STATE FOR YOUR FSM, Workflow %s is Hung", h.WorkflowExecutionSignaledEventAttributes.ExternalWorkflowExecution.WorkflowID)
					case SystemErrorSignal:
						err := &SerializedSystemError{}
						f.Deserialize(h.WorkflowExecutionSignaledEventAttributes.Input, err)
						f.log("action=default-handle-error at=handle-system-error error=%+v", err)
						f.log("YOU SHOULD CREATE AN ERROR STATE FOR YOUR FSM, Workflow %s is Hung", h.WorkflowExecutionSignaledEventAttributes.ExternalWorkflowExecution.WorkflowID)
					default:
						f.log("action=default-handle-error at=process-signal-event event=%+v", h)
					}

				}
			default:
				f.log("action=default-handle-error at=process-event event=%+v", h)
			}
			return fsm.Error(data, []Decision{})
		},
	}
}

// Init initializaed any optional, unspecified values such as the error state, stop channel, serializer, PollerShutdownManager.
// it gets called by Start(), so you should only call this if you are manually managing polling for tasks, and calling Tick yourself.
func (f *FSM) Init() {
	if f.initialState == nil {
		panic("No Initial State Defined For FSM")
	}

	if f.errorState == nil {
		f.AddErrorState(f.DefaultErrorState())
	}

	if f.stop == nil {
		f.stop = make(chan bool, 1)
	}

	if f.stopAck == nil {
		f.stopAck = make(chan bool, 1)
	}

	if f.Serializer == nil {
		f.log("action=start at=no-serializer defaulting-to=JSONSerializer")
		f.Serializer = &JSONStateSerializer{}
	}

	if f.PollerShutdownManager == nil {
		f.PollerShutdownManager = NewPollerShutdownManager()
	}
}

// Start begins processing DecisionTasks with the FSM. It creates a DecisionTaskPoller and spawns a goroutine that continues polling until Stop() is called and any in-flight polls have completed.
// If you wish to manage polling and calling Tick() yourself, you dont need to start the FSM, just call Init().
func (f *FSM) Start() {
	f.Init()
	poller := NewDecisionTaskPoller(f.Client, f.Domain, f.Identity, f.TaskList)
	go poller.PollUntilShutdownBy(f.PollerShutdownManager, fmt.Sprintf("%s-poller", f.Name), f.handleDecisionTask)
}

func (f *FSM) handleDecisionTask(decisionTask *PollForDecisionTaskResponse) {
	decisions, state := f.Tick(decisionTask)
	complete := RespondDecisionTaskCompletedRequest{
		Decisions: decisions,
		TaskToken: decisionTask.TaskToken,
	}
	//currently we arent returning state when in in an error state so only set context in non-error state.
	if state != nil {
		complete.ExecutionContext = state.ReplicationData.StateName
	}

	if err := f.Client.RespondDecisionTaskCompleted(complete); err != nil {
		f.log("action=tick at=decide-request-failed error=%q", err.Error())
		return
	}

	if state == nil || f.KinesisStream == "" {
		return // nothing to replicate
	}
	stateToReplicate, err := f.Serializer.Serialize(state)
	if err != nil {
		f.log("action=tick at=serialize-state-failed error=%q", err.Error())
		return
	}
	resp, err := f.Client.PutRecord(PutRecordRequest{
		StreamName: f.KinesisStream,
		//partition by workflow
		PartitionKey: decisionTask.WorkflowExecution.WorkflowID,
		Data:         []byte(stateToReplicate),
	})
	if err != nil {
		f.log("action=tick at=replicate-state-failed error=%q", err.Error())
		return
	}
	f.log("action=tick at=replicated-state shard=%s sequence=%s", resp.ShardID, resp.SequenceNumber)
	//todo error handling retries etc.
}

func (f *FSM) stateFromDecisions(decisions []Decision) string {
	for _, d := range decisions {
		if d.DecisionType == DecisionTypeRecordMarker && d.RecordMarkerDecisionAttributes.MarkerName == StateMarker {
			return d.RecordMarkerDecisionAttributes.Details
		}
	}
	return ""
}

// Serialize uses the FSM.Serializer to serialize data to a string.
// If there is an error in serialization this func will panic, so this should usually only be used inside Deciders
// where the panics are recovered and proper errors are recorded in the workflow.
func (f *FSM) Serialize(data interface{}) string {
	serialized, err := f.Serializer.Serialize(data)
	if err != nil {
		panic(err)
	}
	return serialized
}

// Deserialize uses the FSM.Serializer to deserialize data from a string.
// If there is an error in deserialization this func will panic, so this should usually only be used inside Deciders
// where the panics are recovered and proper errors are recorded in the workflow.
func (f *FSM) Deserialize(serialized string, data interface{}) {
	err := f.Serializer.Deserialize(serialized, data)
	if err != nil {
		panic(err)
	}
	return
}

// Tick is called when the DecisionTaskPoller receives a PollForDecisionTaskResponse in its polling loop.
// On errors, a nil *SerializedState is returned, and an error Outcome is included in the Decision list.
// It is exported to facilitate testing.
func (f *FSM) Tick(decisionTask *PollForDecisionTaskResponse) ([]Decision, *SerializedState) {
	lastEvents, errorEvents := f.findLastEvents(decisionTask.PreviousStartedEventID, decisionTask.Events)
	execution := decisionTask.WorkflowExecution
	outcome := new(intermediateOutcome)
	serializedState, err := f.findSerializedState(decisionTask.Events)
	if err != nil {
		f.log("action=tick at=error=find-serialized-state-failed err=%q", err)
		if f.allowPanics {
			panic(err)
		}
		return append(outcome.decisions, f.captureSystemError(execution, "FindSerializedStateError", decisionTask.Events, err)...), nil
	}
	f.log("action=tick at=find-serialized-state state=%s", serializedState.ReplicationData.StateName)
	pendingActivities := &serializedState.PendingActivities
	context := NewFSMContext(f,
		decisionTask.WorkflowType,
		decisionTask.WorkflowExecution,
		pendingActivities,
		"", nil, uint64(0),
	)

	//if there are error events, we dont do normal recovery of state + data, we expect the error state to provide this.
	if len(errorEvents) > 0 {
		//todo how can errors mess up versioning
		outcome.data = reflect.New(reflect.TypeOf(f.DataType)).Interface()
		outcome.state = f.errorState.Name
		context.State = f.errorState.Name
		context.stateData = outcome.data
		for i := len(errorEvents) - 1; i >= 0; i-- {
			e := errorEvents[i]
			anOutcome, err := f.panicSafeDecide(f.errorState, context, e, outcome.data)
			if err != nil {
				f.log("at=error error=error-handling-decision-execution-error err=%q state=%s next-state=%s", err, f.errorState.Name, outcome.state)
				//we wont get here if panics are allowed
				return append(outcome.decisions, f.captureDecisionError(execution, i, errorEvents, outcome.state, outcome.data, err)...), nil //TODO: is nil the thing to do?
			}
			f.mergeOutcomes(outcome, anOutcome)
		}
	}

	//if there was no error processing, we recover the state + data from the marker
	if outcome.data == nil && outcome.state == "" {
		data := reflect.New(reflect.TypeOf(f.DataType)).Interface()
		if err = f.Serializer.Deserialize(serializedState.ReplicationData.StateData, data); err != nil {
			f.log("action=tick at=error=deserialize-state-failed err=&s", err)
			if f.allowPanics {
				panic(err)
			}
			return append(outcome.decisions, f.captureSystemError(execution, "DeserializeStateError", decisionTask.Events, err)...), nil
		}
		f.log("action=tick at=find-current-data data=%v", data)

		outcome.data = data
		outcome.state = serializedState.ReplicationData.StateName
		outcome.stateVersion = serializedState.ReplicationData.StateVersion
		context.stateVersion = serializedState.ReplicationData.StateVersion
	}

	//iterate through events oldest to newest, calling the decider for the current state.
	//if the outcome changes the state use the right FSMState
	for i := len(lastEvents) - 1; i >= 0; i-- {
		e := lastEvents[i]
		f.log("action=tick at=history id=%d type=%s", e.EventID, e.EventType)
		fsmState, ok := f.states[outcome.state]
		if ok {
			context.State = outcome.state
			context.stateData = outcome.data
			anOutcome, err := f.panicSafeDecide(fsmState, context, e, outcome.data)
			if err != nil {
				f.log("at=error error=decision-execution-error err=%q state=%s next-state=%s", err, fsmState.Name, outcome.state)
				if f.allowPanics {
					panic(err)
				}
				return append(outcome.decisions, f.captureDecisionError(execution, i, lastEvents, outcome.state, outcome.data, err)...), nil
			}
			pendingActivities.Track(e)
			curr := outcome.state
			f.mergeOutcomes(outcome, anOutcome)
			f.log("action=tick at=decided-event state=%s next-state=%s decisions=%d", curr, outcome.state, len(anOutcome.Decisions()))
		} else {
			f.log("action=tick at=error error=marked-state-not-in-fsm state=%s", outcome.state)
			return append(outcome.decisions, f.captureSystemError(execution, "MissingFsmStateError", lastEvents[i:], errors.New(outcome.state))...), nil
		}
	}

	f.log("action=tick at=events-processed next-state=%s decisions=%d", outcome.state, len(outcome.decisions))

	for _, d := range outcome.decisions {
		f.log("action=tick at=decide next-state=%s decision=%s", outcome.state, d.DecisionType)
	}

	final, serializedState, err := f.recordStateMarker(outcome, pendingActivities)
	if err != nil {
		f.log("action=tick at=error error=state-serialization-error err=%q error-type=system", err)
		if f.allowPanics {
			panic(err)
		}
		return append(outcome.decisions, f.captureSystemError(execution, "StateSerializationError", []HistoryEvent{}, err)...), nil
	}
	return final, serializedState
}

// Stay is a helper func to easily create a StayOutcome.
func (f *FSMContext) Stay(data interface{}, decisions []Decision) Outcome {
	return StayOutcome{
		data:      data,
		decisions: decisions,
	}
}

// Goto is a helper func to easily create a TransitionOutcome.
func (f *FSMContext) Goto(state string, data interface{}, decisions []Decision) Outcome {
	return TransitionOutcome{
		state:     state,
		data:      data,
		decisions: decisions,
	}
}

// Terminate is a helper func to easily create a TerminationOutcome.
func (f *FSMContext) Terminate(data interface{}, decisions []Decision) Outcome {
	return TerminationOutcome{
		data:      data,
		decisions: decisions,
	}
}

// Goto is a helper func to easily create an ErrorOutcome.
func (f *FSMContext) Error(data interface{}, decisions []Decision) Outcome {
	return ErrorOutcome{
		state:     "error",
		data:      data,
		decisions: decisions,
	}
}

func (f *FSM) mergeOutcomes(final *intermediateOutcome, intermediate Outcome) {
	final.decisions = append(final.decisions, intermediate.Decisions()...)
	final.data = intermediate.Data()
	if _, ok := intermediate.(StayOutcome); !ok {
		final.state = intermediate.State()
	}
}

//if the outcome is good good if its an error, we capture the error state above

func (f *FSM) panicSafeDecide(state *FSMState, context *FSMContext, event HistoryEvent, data interface{}) (anOutcome Outcome, anErr error) {
	defer func() {
		if !f.allowPanics {
			if r := recover(); r != nil {
				f.log("at=error error=decide-panic-recovery %v", r)
				if err, ok := r.(error); ok && err != nil {
					anErr = err
				} else {
					anErr = errors.New("panic in decider, null error, capture error state")
				}
			}
		} else {
			log.Printf("at=panic-safe-decide-allowing-panic fsm-allow-panics=%t", f.allowPanics)
		}
	}()
	anOutcome = context.Decide(event, data, state.Decider)
	return
}

func (f *FSM) captureDecisionError(execution WorkflowExecution, event int, lastEvents []HistoryEvent, stateName string, stateData interface{}, err error) []Decision {
	return f.captureError(ErrorSignal, execution, &SerializedDecisionError{
		ErrorEventID:        lastEvents[event].EventID,
		UnprocessedEventIDs: f.eventIDs(lastEvents[event+1:]),
		StateName:           stateName,
		StateData:           stateData,
		Error:               err,
	})
}

func (f *FSM) captureSystemError(execution WorkflowExecution, errorType string, lastEvents []HistoryEvent, err error) []Decision {
	return f.captureError(SystemErrorSignal, execution, &SerializedSystemError{
		ErrorType:           errorType,
		UnprocessedEventIDs: f.eventIDs(lastEvents),
		Error:               err,
	})
}

func (f *FSM) eventIDs(events []HistoryEvent) []int {
	ids := make([]int, len(events))
	for _, e := range events {
		ids = append(ids, e.EventID)
	}
	return ids
}

func (f *FSM) captureError(signal string, execution WorkflowExecution, error interface{}) []Decision {
	decisions := f.EmptyDecisions()
	r, err := f.recordMarker(signal, error)
	if err != nil {
		//really bail
		panic(fmt.Sprintf("giving up, can't even create a RecordMarker decsion: %s", err))
	}
	d := Decision{
		DecisionType: DecisionTypeSignalExternalWorkflowExecution,
		SignalExternalWorkflowExecutionDecisionAttributes: &SignalExternalWorkflowExecutionDecisionAttributes{
			WorkflowID: execution.WorkflowID,
			RunID:      execution.RunID,
			SignalName: signal,
			Input:      r.RecordMarkerDecisionAttributes.Details,
		},
	}
	return append(decisions, d)
}

// EventData works in combination with the FSM.Serializer to provide
// deserialization of data sent in a HistoryEvent. It is sugar around extracting the event payload from the proper
// field of the proper Attributes struct on the HistoryEvent
func (f *FSM) EventData(ctx *FSMContext, event HistoryEvent, eventData interface{}) {

	if eventData != nil {
		var serialized string
		switch event.EventType {
		case EventTypeActivityTaskCompleted:
			serialized = event.ActivityTaskCompletedEventAttributes.Result
		case EventTypeChildWorkflowExecutionFailed:
			serialized = event.ActivityTaskFailedEventAttributes.Details
		case EventTypeWorkflowExecutionCompleted:
			serialized = event.WorkflowExecutionCompletedEventAttributes.Result
		case EventTypeChildWorkflowExecutionCompleted:
			serialized = event.ChildWorkflowExecutionCompletedEventAttributes.Result
		case EventTypeWorkflowExecutionSignaled:
			serialized = event.WorkflowExecutionSignaledEventAttributes.Input
		case EventTypeWorkflowExecutionStarted:
			serialized = event.WorkflowExecutionStartedEventAttributes.Input
		case EventTypeWorkflowExecutionContinuedAsNew:
			serialized = event.WorkflowExecutionContinuedAsNewEventAttributes.Input
		}
		if serialized != "" {
			f.Deserialize(serialized, eventData)
		} else {
			panic(fmt.Sprintf("event payload was empty for %s", event.String()))
		}
	}

}

func (f *FSM) log(format string, data ...interface{}) {
	actualFormat := fmt.Sprintf("component=FSM name=%s %s", f.Name, format)
	log.Printf(actualFormat, data...)
}

func (f *FSM) findSerializedState(events []HistoryEvent) (*SerializedState, error) {
	for _, event := range events {
		if f.isStateMarker(event) {
			state := &SerializedState{}
			err := f.Serializer.Deserialize(event.MarkerRecordedEventAttributes.Details, state)
			return state, err
		} else if event.EventType == EventTypeWorkflowExecutionStarted {
			log.Println(event)
			state := &SerializedState{}
			err := f.Serializer.Deserialize(event.WorkflowExecutionStartedEventAttributes.Input, state)
			if err == nil {
				if state.ReplicationData.StateName == "" {
					state.ReplicationData.StateName = f.initialState.Name
				}
			}
			return state, err
		}
	}
	return nil, errors.New("Cant Find Current Data")
}

func (f *FSM) findLastEvents(prevStarted int, events []HistoryEvent) ([]HistoryEvent, []HistoryEvent) {
	var lastEvents []HistoryEvent
	var errorEvents []HistoryEvent

	for _, event := range events {
		if event.EventID == prevStarted {
			return lastEvents, errorEvents
		}
		switch event.EventType {
		case EventTypeDecisionTaskCompleted, EventTypeDecisionTaskScheduled,
			EventTypeDecisionTaskStarted:
			//no-op, dont even process these?
		case EventTypeMarkerRecorded:
			if !f.isStateMarker(event) {
				lastEvents = append(lastEvents, event)
			}
		case EventTypeWorkflowExecutionSignaled:
			if f.isErrorSignal(event) {
				errorEvents = append(errorEvents, event)
			} else {
				lastEvents = append(lastEvents, event)
			}
		default:
			lastEvents = append(lastEvents, event)
		}

	}

	return lastEvents, errorEvents
}

func (f *FSM) recordStateMarker(outcome *intermediateOutcome, pending *ActivityCorrelator) ([]Decision, *SerializedState, error) {
	serializedData, err := f.Serializer.Serialize(outcome.data)

	state := &SerializedState{
		ReplicationData: ReplicationData{
			StateVersion: outcome.stateVersion + 1, //increment the version here only.
			StateName:    outcome.state,
			StateData:    serializedData,
		},
		PendingActivities: *pending,
	}

	d, err := f.recordMarker(StateMarker, state)
	if err != nil {
		return nil, state, err
	}
	decisions := f.EmptyDecisions()
	decisions = append(decisions, d)
	decisions = append(decisions, outcome.decisions...)
	return decisions, state, nil
}

func (f *FSM) recordMarker(markerName string, details interface{}) (Decision, error) {
	serialized, err := f.Serializer.Serialize(details)
	if err != nil {
		return Decision{}, err
	}

	return f.recordStringMarker(markerName, serialized), nil
}

func (f *FSM) recordStringMarker(markerName string, details string) Decision {
	return Decision{
		DecisionType: DecisionTypeRecordMarker,
		RecordMarkerDecisionAttributes: &RecordMarkerDecisionAttributes{
			MarkerName: markerName,
			Details:    details,
		},
	}
}

// Stop causes the DecisionTask select loop to exit, and to stop the DecisionTaskPoller
func (f *FSM) Stop() {
	f.stop <- true
}

func (f *FSM) isStateMarker(e HistoryEvent) bool {
	return e.EventType == EventTypeMarkerRecorded && e.MarkerRecordedEventAttributes.MarkerName == StateMarker
}

func (f *FSM) isErrorSignal(e HistoryEvent) bool {
	if e.EventType == EventTypeWorkflowExecutionSignaled {
		switch e.WorkflowExecutionSignaledEventAttributes.SignalName {
		case SystemErrorSignal, ErrorSignal:
			return true
		default:
			return false
		}
	} else {
		return false
	}
}

// EmptyDecisions is a helper method to give you an empty decisions array for use in your Deciders.
func (f *FSM) EmptyDecisions() []Decision {
	return make([]Decision, 0)
}

// SerializedState is a wrapper struct that allows serializing the current state and current data for the FSM in
// a MarkerRecorded event in the workflow history. We also maintain an epoch, which counts the number of times a workflow has
// been continued, and the StartedId of the DecisionTask that generated this state.  The epoch + the id provide a total ordering
// of state over the lifetime of different runs of a workflow.
type SerializedState struct {
	ReplicationData   ReplicationData
	PendingActivities ActivityCorrelator
}

// StartFSMWorkflowInput should be used to construct the input for any StartWorkflowExecutionRequests.
// This panics on errors cause really this should never err.
func StartFSMWorkflowInput(serializer StateSerializer, data interface{}) string {
	ss := new(SerializedState)
	stateData, err := serializer.Serialize(data)
	if err != nil {
		panic(err)
	}

	ss.ReplicationData.StateData = stateData
	serialized, err := serializer.Serialize(ss)
	if err != nil {
		panic(err)
	}
	return serialized
}

//ContinueFSMWorkflowInput should be used to construct the input for any ContinueAsNewWorkflowExecution decisions.
func ContinueFSMWorkflowInput(ctx *FSMContext, data interface{}) string {
	ss := new(SerializedState)
	stateData := ctx.Serialize(data)

	ss.ReplicationData.StateData = stateData
	ss.ReplicationData.StateName = ctx.fsm.initialState.Name
	ss.ReplicationData.StateVersion = ctx.stateVersion

	return ctx.Serialize(ss)
}

// ReplicationData is the part of SerializedState that will be replicated onto Kinesis streams.
type ReplicationData struct {
	StateVersion uint64 `json:"stateVersion"`
	StateName    string `json:"stateName"`
	StateData    string `json:"stateData"`
}

// SerializedDecisionError is a wrapper struct that allows serializing the context in which an error in a Decider occurred
// into a WorkflowSignaledEvent in the workflow history.
type SerializedDecisionError struct {
	ErrorEventID        int         `json:"errorEventIds"`
	UnprocessedEventIDs []int       `json:"unprocessedEventIds"`
	StateName           string      `json:"stateName"`
	StateData           interface{} `json:"stateData"`
	Error               interface{} `json:"error"`
}

// SerializedSystemError is a wrapper struct that allows serializing the context in which an error internal to FSM processing has occurred
// into a WorkflowSignaledEvent in the workflow history. These errors are generally in finding the current state and data for a workflow, or
// in serializing and deserializing said state.
type SerializedSystemError struct {g
	ErrorType           string      `json:"errorType"`
	Error               interface{} `json:"error"`
	UnprocessedEventIDs []int       `json:"unprocessedEventIds"`
}

// StateSerializer defines the interface for serializing state to and deserializing state from the workflow history.
type StateSerializer interface {
	Serialize(state interface{}) (string, error)
	Deserialize(serialized string, state interface{}) error
}

// JSONStateSerializer is a StateSerializer that uses go json serialization.
type JSONStateSerializer struct{}

// Serialize serializes the given struct to a json string.
func (j JSONStateSerializer) Serialize(state interface{}) (string, error) {
	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(state); err != nil {
		return "", err
	}
	return b.String(), nil
}

// Deserialize unmarshalls the given (json) string into the given struct
func (j JSONStateSerializer) Deserialize(serialized string, state interface{}) error {
	err := json.NewDecoder(strings.NewReader(serialized)).Decode(state)
	return err
}

// ProtobufStateSerializer is a StateSerializer that uses base64 encoded protobufs.
type ProtobufStateSerializer struct{}

// Serialize serializes the given struct into bytes with protobuf, then base64 encodes it.  The struct passed to Serialize must satisfy proto.Message.
func (p ProtobufStateSerializer) Serialize(state interface{}) (string, error) {
	bin, err := proto.Marshal(state.(proto.Message))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bin), nil
}

// Deserialize base64 decodes the given string then unmarshalls the bytes into the struct using protobuf. The struct passed to Deserialize must satisfy proto.Message.
func (p ProtobufStateSerializer) Deserialize(serialized string, state interface{}) error {
	bin, err := base64.StdEncoding.DecodeString(serialized)
	if err != nil {
		return err
	}
	err = proto.Unmarshal(bin, state.(proto.Message))

	return err
}

/*tigertonic-like marhsalling of data from interface{} to specific type*/

// MarshalledDecider is used to convert a standard decider with data of type interface{} to a typed decider
// which has data of user specified type FSM.DataType
type MarshalledDecider struct {
	v reflect.Value
}

// TypedDecider wraps a user specified typed decider in a standard decider.
func TypedDecider(decider interface{}) Decider {
	t := reflect.TypeOf(decider)
	if reflect.Func != t.Kind() {
		panic(fmt.Sprintf("kind was %v, not Func", t.Kind()))
	}
	if 3 != t.NumIn() {
		panic(fmt.Sprintf(
			"input arity was %v, not 3",
			t.NumIn(),
		))
	}
	if "*swf.FSMContext" != t.In(0).String() {
		panic(fmt.Sprintf(
			"type of first argument was %v, not *swf.FSMContext",
			t.In(0),
		))
	}
	if "swf.HistoryEvent" != t.In(1).String() {
		panic(fmt.Sprintf(
			"type of second argument was %v, not swf.HistoryEvent",
			t.In(1),
		))
	}

	if "swf.Outcome" != t.Out(0).String() {
		panic(fmt.Sprintf(
			"type of return value was %v, not swf.Outcome",
			t.Out(0),
		))
	}

	return MarshalledDecider{reflect.ValueOf(decider)}.Decide
}

// Decide uses reflection to call the user specified, typed decider.
func (m MarshalledDecider) Decide(f *FSMContext, h HistoryEvent, data interface{}) Outcome {

	// reflection will asplode if we try to use nil
	if data == nil {
		data = reflect.New(reflect.TypeOf(f.fsm.DataType)).Interface()
	}
	return m.v.Call([]reflect.Value{reflect.ValueOf(f), reflect.ValueOf(h), reflect.ValueOf(data)})[0].Interface().(Outcome)
}

// FSMContext is populated by the FSM machinery and passed to Deciders.
type FSMContext struct {
	fsm *FSM
	WorkflowType
	WorkflowExecution
	pendingActivities *ActivityCorrelator
	State             string
	stateData         interface{}
	stateVersion      uint64
}

// NewFSMContext constructs an FSMContext.
func NewFSMContext(
	fsm *FSM,
	wfType WorkflowType, wfExec WorkflowExecution,
	pending *ActivityCorrelator,
	state string, stateData interface{}, stateVersion uint64,
) *FSMContext {
	return &FSMContext{
		fsm:               fsm,
		WorkflowType:      wfType,
		WorkflowExecution: wfExec,
		pendingActivities: pending,
		State:             state,
		stateData:         stateData,
		stateVersion:      stateVersion,
	}
}

// Decide executes a decider making sure that Activity tasks are being tracked.
func (f *FSMContext) Decide(h HistoryEvent, data interface{}, decider Decider) Outcome {
	outcome := decider(f, h, data)
	f.pendingActivities.Track(h)
	return outcome
}

// EventData will extract a payload from the given HistoryEvent and unmarshall it into the given struct.
func (f *FSMContext) EventData(h HistoryEvent, data interface{}) {
	f.fsm.EventData(f, h, data)
}

// ActivityInfo will find information for ActivityTasks being tracked. It can only be used when handling events related to ActivityTasks.
// ActivityTasks are automatically tracked after a EventTypeActivityTaskScheduled event.
// When there is no pending activity related to the event, nil is returned.
func (f *FSMContext) ActivityInfo(h HistoryEvent) *ActivityInfo {
	return f.pendingActivities.ActivityType(h)
}

// ActivitiesInfo will return a map of activityId -> ActivityInfo for all in-flight activities in the workflow.
func (f *FSMContext) ActivitiesInfo() map[string]*ActivityInfo {
	return f.pendingActivities.Activities
}

// Serialize will use the current fsm's Serializer to serialize the given struct. It will panic on errors, which is ok in the context of a Decider.
// If you want to handle errors, use Serializer().Serialize(...) instead.
func (f *FSMContext) Serialize(data interface{}) string {
	return f.fsm.Serialize(data)
}

// Serializer returns the current fsm's Serializer.
func (f *FSMContext) Serializer() StateSerializer {
	return f.fsm.Serializer
}

// Deserialize will use the current fsm' Serializer to deserialize the given string into the given struct. It will panic on errors, which is ok in the context of a Decider.
// If you want to handle errors, use Serializer().Deserialize(...) instead.
func (f *FSMContext) Deserialize(serialized string, data interface{}) {
	f.fsm.Deserialize(serialized, data)
}

// EmptyDecisions is a helper to give you an empty Decision slice.
func (f *FSMContext) EmptyDecisions() []Decision {
	return f.fsm.EmptyDecisions()
}

// ContinuationDecision will build a ContinueAsNewWorkflow decision that has the expected SerializedState marshalled to json as its input.
// This decision should be used when it is appropriate to Continue your workflow.
// You are unable to ContinueAsNew a workflow that has running activites, so you should assure there are none running before using this.
// As such there is no need to copy over the ActivityCorrelator.
func (f *FSMContext) ContinuationDecision(continuedState string) Decision {
	return Decision{
		DecisionType: DecisionTypeContinueAsNewWorkflowExecution,
		ContinueAsNewWorkflowExecutionDecisionAttributes: &ContinueAsNewWorkflowExecutionDecisionAttributes{
			Input: f.Serialize(SerializedState{
				ReplicationData: ReplicationData{
					StateName:    continuedState,
					StateData:    f.Serialize(f.stateData),
					StateVersion: f.stateVersion,
				},
				PendingActivities: ActivityCorrelator{},
			},
			),
		},
	}
}

// ActivityCorrelator is a serialization-friendly struct that can be used as a field in your main StateData struct in an FSM.
// You can use it to track the type of a given activity, so you know how to react when an event that signals the
// end or an activity hits your Decider.  This is missing from the SWF api.
type ActivityCorrelator struct {
	Activities map[string]*ActivityInfo
}

// ActivityInfo holds the ActivityID and ActivityType for an activity
type ActivityInfo struct {
	ActivityID string
	ActivityType
}

// Track will add or remove entries based on the EventType.
// A new entry is added when there is a new ActivityTask, or an entry is removed when the ActivityTask is terminating.
func (a *ActivityCorrelator) Track(h HistoryEvent) {
	a.RemoveCorrelation(h)
	a.Correlate(h)
}

// Correlate establishes a mapping of eventId to ActivityType. The HistoryEvent is expected to be of type EventTypeActivityTaskScheduled.
func (a *ActivityCorrelator) Correlate(h HistoryEvent) {
	if a.Activities == nil {
		a.Activities = make(map[string]*ActivityInfo)
	}
	if h.EventType == EventTypeActivityTaskScheduled {
		a.Activities[strconv.Itoa(h.EventID)] = &ActivityInfo{
			ActivityID:   h.ActivityTaskScheduledEventAttributes.ActivityID,
			ActivityType: h.ActivityTaskScheduledEventAttributes.ActivityType,
		}
	}
}

// RemoveCorrelation gcs a mapping of eventId to ActivityType. The HistoryEvent is expected to be of type EventTypeActivityTaskCompleted,EventTypeActivityTaskFailed,EventTypeActivityTaskTimedOut.
func (a *ActivityCorrelator) RemoveCorrelation(h HistoryEvent) {
	if a.Activities == nil {
		a.Activities = make(map[string]*ActivityInfo)
	}
	switch h.EventType {
	case EventTypeActivityTaskCompleted:
		delete(a.Activities, strconv.Itoa(h.ActivityTaskCompletedEventAttributes.ScheduledEventID))
	case EventTypeActivityTaskFailed:
		delete(a.Activities, strconv.Itoa(h.ActivityTaskFailedEventAttributes.ScheduledEventID))
	case EventTypeActivityTaskTimedOut:
		delete(a.Activities, strconv.Itoa(h.ActivityTaskTimedOutEventAttributes.ScheduledEventID))
	case EventTypeActivityTaskCanceled:
		delete(a.Activities, strconv.Itoa(h.ActivityTaskCanceledEventAttributes.ScheduledEventID))
	}
}

// ActivityType returns the ActivityType that is correlates with a given event. The HistoryEvent is expected to be of type EventTypeActivityTaskCompleted,EventTypeActivityTaskFailed,EventTypeActivityTaskTimedOut.
func (a *ActivityCorrelator) ActivityType(h HistoryEvent) *ActivityInfo {
	if a.Activities == nil {
		a.Activities = make(map[string]*ActivityInfo)
	}
	switch h.EventType {
	case EventTypeActivityTaskCompleted:
		return a.Activities[strconv.Itoa(h.ActivityTaskCompletedEventAttributes.ScheduledEventID)]
	case EventTypeActivityTaskFailed:
		return a.Activities[strconv.Itoa(h.ActivityTaskFailedEventAttributes.ScheduledEventID)]
	case EventTypeActivityTaskTimedOut:
		return a.Activities[strconv.Itoa(h.ActivityTaskTimedOutEventAttributes.ScheduledEventID)]
	case EventTypeActivityTaskCanceled:
		return a.Activities[strconv.Itoa(h.ActivityTaskCanceledEventAttributes.ScheduledEventID)]
	}
	return nil
}
