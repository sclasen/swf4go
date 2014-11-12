package swf

import (
	"bytes"
	"code.google.com/p/goprotobuf/proto"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
)

// constants used as marker names or signal names
const (
	STATE_MARKER        = "FSM.State"
	ERROR_SIGNAL        = "FSM.Error"
	SYSTEM_ERROR_SIGNAL = "FSM.SystemError"
)

// Decider decides an Outcome based on an event and the current data for an FSM
// the interface{} parameter that is passed to the decider is safe to
// be asserted to be the type of the DataType field in the FSM
// Alternatively you can use the TypedDecider to avoid having to do the assertion.
type Decider func(*FSMContext, HistoryEvent, interface{}) Outcome

// EventDataType should return an empty struct of the correct type based on the event
// the FSM will unmarshal data from the event into this struct
type EventDataType func(HistoryEvent) interface{}

type Outcome interface {
	Data() interface{}
	Decisions() []*Decision
}

type TransitionOutcome struct {
	data      interface{}
	state     string
	decisions []*Decision
}

func (t TransitionOutcome) Data() interface{} { return t.data }

func (t TransitionOutcome) Decisions() []*Decision { return t.decisions }

type StayOutcome struct {
	data      interface{}
	decisions []*Decision
}

func (s StayOutcome) Data() interface{} { return s.data }

func (s StayOutcome) Decisions() []*Decision { return s.decisions }

//this can do things like check that the last decision is a termination?
type TerminationOutcome struct {
	data      interface{}
	decisions []*Decision
}

func (t TerminationOutcome) Data() interface{} { return t.data }

func (t TerminationOutcome) Decisions() []*Decision { return t.decisions }

//This could be used to purposefully put the workflow into an error state.
type ErrorOutcome struct {
	state     string
	data      interface{}
	decisions []*Decision
}

func (e ErrorOutcome) Data() interface{} { return e.data }

func (e ErrorOutcome) Decisions() []*Decision { return e.decisions }

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
	Client *Client
	// DataType of the data struct associated with this FSM.
	// The data is automatically peristed to and loaded from workflow history by the FSM.
	DataType interface{}
	// EventDataType returns a zero value struct to deserialize the payload of a HistoryEvent.
	EventDataType EventDataType
	// Serializer used to serialize/deserialise state from workflow history.
	Serializer StateSerializer
	// Kinesis stream in the same region to replicate state to.
	KinesisStream string
	//PollerShtudownManager is used when the FSM is managing the polling
	PollerShutdownManager *PollerShtudownManager
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
					case ERROR_SIGNAL:
						err := &SerializedDecisionError{}
						f.Deserialize(h.WorkflowExecutionSignaledEventAttributes.Input, err)
						f.log("action=default-handle-error at=handle-decision-error error=%+v", err)
						f.log("YOU SHOULD CREATE AN ERROR STATE FOR YOUR FSM, Workflow %s is Hung", h.WorkflowExecutionSignaledEventAttributes.ExternalWorkflowExecution.WorkflowId)
					case SYSTEM_ERROR_SIGNAL:
						err := &SerializedSystemError{}
						f.Deserialize(h.WorkflowExecutionSignaledEventAttributes.Input, err)
						f.log("action=default-handle-error at=handle-system-error error=%+v", err)
						f.log("YOU SHOULD CREATE AN ERROR STATE FOR YOUR FSM, Workflow %s is Hung", h.WorkflowExecutionSignaledEventAttributes.ExternalWorkflowExecution.WorkflowId)
					default:
						f.log("action=default-handle-error at=process-signal-event event=%+v", h)
					}

				}
			default:
				f.log("action=default-handle-error at=process-event event=%+v", h)
			}
			return fsm.Error(data, []*Decision{})
		},
	}
}

// Init initializaed any optional, unspecified values such as the error state, stop channel, serializer, PollerShtudownManager.
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
		f.log("action=start at=no-serializer defaulting-to=JsonSerializer")
		f.Serializer = &JsonStateSerializer{}
	}

	if f.PollerShutdownManager == nil {
		f.PollerShutdownManager = RegisterPollerShutdownManager()
	}
}

// Start begins processing DecisionTasks with the FSM. It creates a DecisionTaskPoller and spawns a goroutine that continues polling until Stop() is called and any in-flight polls have completed.
// If you wish to manage polling and calling Tick() yourself, you dont need to start the FSM, just call Init().
func (f *FSM) Start() {
	f.Init()
	f.PollerShutdownManager.Register(f.Name, f.stop, f.stopAck)
	poller := f.Client.DecisionTaskPoller(f.Domain, f.Identity, f.TaskList)
	go poller.PollUntilShutdownBy(f.PollerShutdownManager, f.Name, func(decisionTask *PollForDecisionTaskResponse) {
		decisions := f.Tick(decisionTask)
		err := f.Client.RespondDecisionTaskCompleted(
			RespondDecisionTaskCompletedRequest{
				Decisions: decisions,
				TaskToken: decisionTask.TaskToken,
			})

		if err != nil {
			f.log("action=tick at=decide-request-failed error=%s", err.Error())
		}

		if f.KinesisStream != "" {
			stateToReplicate := f.stateFromDecisions(decisions)
			if stateToReplicate != "" {
				resp, err := f.Client.PutRecord(PutRecordRequest{
					StreamName: f.KinesisStream,
					//partition by workflow
					PartitionKey: decisionTask.WorkflowExecution.WorkflowId,
					Data: []byte(stateToReplicate),
				})
				if err != nil {
					f.log("action=tick at=replicate-state-failed error=%s", err.Error())
				} else {
					f.log("action=tick at=replicated-state shard=%s sequence=%s", resp.ShardId, resp.SequenceNumber)
				}
				//todo error handling retries etc.
			}
		}

	})
}

func (f *FSM) stateFromDecisions(decisions []*Decision) string {
	for _, d := range decisions {
		if d.DecisionType == DecisionTypeRecordMarker && d.RecordMarkerDecisionAttributes.MarkerName == STATE_MARKER {
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
// It is exported to facilitate testing.
func (f *FSM) Tick(decisionTask *PollForDecisionTaskResponse) []*Decision {
	lastEvents, errorEvents := f.findLastEvents(decisionTask.PreviousStartedEventId, decisionTask.Events)
	execution := decisionTask.WorkflowExecution
	outcome := new(TransitionOutcome)
	context := &FSMContext{f, decisionTask.WorkflowType, decisionTask.WorkflowExecution, ""}
	//if there are error events, we dont do normal recovery of state + data, we expect the error state to provide this.
	if len(errorEvents) > 0 {
		outcome.data = reflect.New(reflect.TypeOf(f.DataType)).Interface()
		outcome.state = f.errorState.Name
		context.State = f.errorState.Name
		for i := len(errorEvents) - 1; i >= 0; i-- {
			e := errorEvents[i]
			anOutcome, err := f.panicSafeDecide(f.errorState, context, e, outcome.data)
			if err != nil {
				f.log("at=error error=error-handling-decision-execution-error state=%s next-state=%", f.errorState.Name, outcome.state)
				//we wont get here if panics are allowed
				return append(outcome.decisions, f.captureDecisionError(execution, i, errorEvents, outcome.state, outcome.data, err)...)
			}
			f.mergeOutcomes(outcome, anOutcome)
		}

	}

	//if there was no error processing, we recover the state + data from the expected marker
	if outcome.data == nil && outcome.state == "" {
		serializedState, err := f.findSerializedState(decisionTask.Events)

		if err != nil {
			f.log("action=tick at=error=find-serialized-state-failed")
			if f.allowPanics {
				panic(err)
			}
			return append(outcome.decisions, f.captureSystemError(execution, "FindSerializedStateError", decisionTask.Events, err)...)
		}

		f.log("action=tick at=find-current-state state=%s", serializedState.StateName)
		data := reflect.New(reflect.TypeOf(f.DataType)).Interface()
		err = f.Serializer.Deserialize(serializedState.StateData, data)
		if err != nil {
			f.log("action=tick at=error=deserialize-state-failed")
			if f.allowPanics {
				panic(err)
			}
			return append(outcome.decisions, f.captureSystemError(execution, "DeserializeStateError", decisionTask.Events, err)...)
		}

		f.log("action=tick at=find-current-data data=%v", data)

		outcome.data = data
		outcome.state = serializedState.StateName
	}

	//iterate through events oldest to newest, calling the decider for the current state.
	//if the outcome changes the state use the right FSMState
	for i := len(lastEvents) - 1; i >= 0; i-- {
		e := lastEvents[i]
		f.log("action=tick at=history id=%d type=%s", e.EventId, e.EventType)
		fsmState, ok := f.states[outcome.state]
		if ok {
			context.State = outcome.state
			anOutcome, err := f.panicSafeDecide(fsmState, context, e, outcome.data)
			if err != nil {
				f.log("at=error error=decision-execution-error state=%s next-state=%", fsmState.Name, outcome.state)
				if f.allowPanics {
					panic(err)
				}
				return append(outcome.decisions, f.captureDecisionError(execution, i, lastEvents, outcome.state, outcome.data, err)...)
			}

			curr := outcome.state
			f.mergeOutcomes(outcome, anOutcome)
			f.log("action=tick at=decided-event state=%s next-state=%s decisions=%d", curr, outcome.state, len(anOutcome.Decisions()))

		} else {
			f.log("action=tick at=error error=marked-state-not-in-fsm state=%s", outcome.state)
			return append(outcome.decisions, f.captureSystemError(execution, "MissingFsmStateError", lastEvents[i:], errors.New(outcome.state))...)
		}
	}

	f.log("action=tick at=events-processed next-state=%s decisions=%d", outcome.state, len(outcome.decisions))

	for _, d := range outcome.decisions {
		f.log("action=tick at=decide next-state=%s decision=%s", outcome.state, d.DecisionType)
	}

	final, err := f.appendState(outcome)
	if err != nil {
		f.log("action=tick at=error error=state-serialization-error error-type=system")
		if f.allowPanics {
			panic(err)
		}
		return append(outcome.decisions, f.captureSystemError(execution, "StateSerializationError", []HistoryEvent{}, err)...)
	}
	return final
}

func (f *FSMContext) Stay(data interface{}, decisions []*Decision) Outcome {
	return StayOutcome{
		data:      data,
		decisions: decisions,
	}
}

func (f *FSMContext) Goto(state string, data interface{}, decisions []*Decision) Outcome {
	return TransitionOutcome{
		state:     state,
		data:      data,
		decisions: decisions,
	}
}

func (f *FSMContext) Terminate(data interface{}, decisions []*Decision) Outcome {
	return TerminationOutcome{
		data:      data,
		decisions: decisions,
	}
}

func (f *FSMContext) Error(data interface{}, decisions []*Decision) Outcome {
	return ErrorOutcome{
		state:     "error",
		data:      data,
		decisions: decisions,
	}
}

func (f *FSM) mergeOutcomes(final *TransitionOutcome, intermediate Outcome) {
	switch i := intermediate.(type) {
	case TransitionOutcome:
		final.state = i.state
		final.decisions = append(final.decisions, i.decisions...)
		final.data = i.data
	case StayOutcome:
		final.decisions = append(final.decisions, i.decisions...)
		final.data = i.data
	case TerminationOutcome:
		final.state = "TERMINATED"
		final.decisions = append(final.decisions, i.decisions...)
		final.data = i.data
	case ErrorOutcome:
		final.state = "error"
		final.decisions = append(final.decisions, i.decisions...)
		final.data = i.data
	default:
		panic("funky")

	}
}

//if the outcome is good good if its an error, we capture the error state above

func (f *FSM) panicSafeDecide(state *FSMState, context *FSMContext, event HistoryEvent, data interface{}) (anOutcome Outcome, anErr error) {
	defer func() {
		if !f.allowPanics {
			if r := recover(); r != nil {
				f.log("at=error error=decide-panic-recovery %v", r)
				anErr = errors.New("panic in decider, capture error state")
			}
		} else {
			log.Printf("at=panic-safe-decide-allowing-panic fsm-allow-panics=%t", f.allowPanics)
		}
	}()
	anOutcome = state.Decider(context, event, data)
	return
}

func (f *FSM) captureDecisionError(execution WorkflowExecution, event int, lastEvents []HistoryEvent, stateName string, stateData interface{}, err error) []*Decision {
	return f.captureError(ERROR_SIGNAL, execution, &SerializedDecisionError{
		ErrorEventId:        lastEvents[event].EventId,
		UnprocessedEventIds: f.eventIds(lastEvents[event+1:]),
		StateName:           stateName,
		StateData:           stateData,
	})
}

func (f *FSM) captureSystemError(execution WorkflowExecution, errorType string, lastEvents []HistoryEvent, err error) []*Decision {
	return f.captureError(SYSTEM_ERROR_SIGNAL, execution, &SerializedSystemError{
		ErrorType:           errorType,
		UnprocessedEventIds: f.eventIds(lastEvents),
		Error:               err,
	})
}

func (f *FSM) eventIds(events []HistoryEvent) []int {
	ids := make([]int, len(events))
	for _, e := range events {
		ids = append(ids, e.EventId)
	}
	return ids
}

func (f *FSM) captureError(signal string, execution WorkflowExecution, error interface{}) []*Decision {
	decisions := f.EmptyDecisions()
	r, err := f.recordMarker(signal, error)
	if err != nil {
		//really bail
		panic("giving up, cant even create a RecordMarker decsion")
	}
	d := &Decision{
		DecisionType: DecisionTypeSignalExternalWorkflowExecution,
		SignalExternalWorkflowExecutionDecisionAttributes: &SignalExternalWorkflowExecutionDecisionAttributes{
			WorkflowId: execution.WorkflowId,
			RunId:      execution.RunId,
			SignalName: signal,
			Input:      r.RecordMarkerDecisionAttributes.Details,
		},
	}
	return append(decisions, d)
}

// EventData works in combination with the FSM.Serializer and the FSM.EventDataType function to provide
// deserialization of data sent in a HistoryEvent. For example if you know that any WorkflowExecutionSignaled
// event that your FSM receives will contain a serialized Foo struct, and your EventDataType func returns a zero value Foo,
// for WorkflowExecutionSignaled events, you can use EventData plus a type assertion to get back the deserialized Foo.
//   if event.EventType == swf.EventTypeWorkflowExecutionSignaled {
//       f.EventData(event).(*Foo)
//   }
func (f *FSM) EventData(event HistoryEvent) interface{} {
	eventData := f.EventDataType(event)

	if eventData != nil {
		var serialized string
		switch event.EventType {
		case EventTypeActivityTaskCompleted:
			serialized = event.ActivityTaskCompletedEventAttributes.Result
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
		}
	}

	return eventData

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
			return &SerializedState{StateName: f.initialState.Name, StateData: event.WorkflowExecutionStartedEventAttributes.Input}, nil
		}
	}
	return nil, errors.New("Cant Find Current Data")
}

func (f *FSM) findLastEvents(prevStarted int, events []HistoryEvent) ([]HistoryEvent, []HistoryEvent) {
	var lastEvents []HistoryEvent
	var errorEvents []HistoryEvent

	for _, event := range events {
		if event.EventId == prevStarted {
			return lastEvents, errorEvents
		}
		switch event.EventType {
		case EventTypeDecisionTaskCompleted, EventTypeDecisionTaskScheduled,
			EventTypeDecisionTaskStarted, EventTypeDecisionTaskTimedOut:
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

func (f *FSM) appendState(outcome *TransitionOutcome) ([]*Decision, error) {

	serializedData, err := f.Serializer.Serialize(outcome.data)

	state := SerializedState{
		StateName: outcome.state,
		StateData: serializedData,
	}

	d, err := f.recordMarker(STATE_MARKER, state)
	if err != nil {
		return nil, err
	}
	decisions := f.EmptyDecisions()
	decisions = append(decisions, d)
	decisions = append(decisions, outcome.decisions...)
	return decisions, nil
}

func (f *FSM) recordMarker(markerName string, details interface{}) (*Decision, error) {
	serialized, err := f.Serializer.Serialize(details)
	if err != nil {
		return nil, err
	}

	return f.recordStringMarker(markerName, serialized), nil
}

func (f *FSM) recordStringMarker(markerName string, details string) *Decision {
	return &Decision{
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
	return e.EventType == EventTypeMarkerRecorded && e.MarkerRecordedEventAttributes.MarkerName == STATE_MARKER
}

func (f *FSM) isErrorSignal(e HistoryEvent) bool {
	if e.EventType == EventTypeWorkflowExecutionSignaled {
		switch e.WorkflowExecutionSignaledEventAttributes.SignalName {
		case SYSTEM_ERROR_SIGNAL, ERROR_SIGNAL:
			return true
		default:
			return false
		}
	} else {
		return false
	}
}

// EmptyDecisions is a helper method to give you an empty decisions array for use in your Deciders.
func (f *FSM) EmptyDecisions() []*Decision {
	return make([]*Decision, 0)
}

// SerializedState is a wrapper struct that allows serializing the current state and current data for the FSM in
// a MarkerRecorded event in the workflow history.
type SerializedState struct {
	StateName string `json:"stateName"`
	StateData string `json:"stateData"`
}

// SerializedDecisionError is a wrapper struct that allows serializing the context in which an error in a Decider occurred
// into a WorkflowSignaledEvent in the workflow history.
type SerializedDecisionError struct {
	ErrorEventId        int         `json:"errorEventIds"`
	UnprocessedEventIds []int       `json:"unprocessedEventIds"`
	StateName           string      `json:"stateName"`
	StateData           interface{} `json:"stateData"`
}

// SerializedSystemError is a wrapper struct that allows serializing the context in which an error internal to FSM processing has occurred
// into a WorkflowSignaledEvent in the workflow history. These errors are generally in finding the current state and data for a workflow, or
// in serializing and deserializing said state.
type SerializedSystemError struct {
	ErrorType           string      `json:"errorType"`
	Error               interface{} `json:"error"`
	UnprocessedEventIds []int       `json:"unprocessedEventIds"`
}

// StateSerializer defines the interface for serializing state to and deserializing state from the workflow history.
type StateSerializer interface {
	Serialize(state interface{}) (string, error)
	Deserialize(serialized string, state interface{}) error
}

// JsonStateSerializer is a StateSerializer that uses go json serialization.
type JsonStateSerializer struct{}

func (j JsonStateSerializer) Serialize(state interface{}) (string, error) {
	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(state); err != nil {
		return "", err
	}
	return b.String(), nil
}

func (j JsonStateSerializer) Deserialize(serialized string, state interface{}) error {
	err := json.NewDecoder(strings.NewReader(serialized)).Decode(state)
	return err
}

type ProtobufStateSerializer struct{}

func (p ProtobufStateSerializer) Serialize(state interface{}) (string, error) {
	bin, err := proto.Marshal(state.(proto.Message))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bin), nil
}

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

type FSMContext struct {
	fsm *FSM
	WorkflowType
	WorkflowExecution
	State string
}

func (f *FSMContext) EventData(h HistoryEvent) interface{} {
	return f.fsm.EventData(h)
}

func (f *FSMContext) Serialize(data interface{}) string {
	return f.fsm.Serialize(data)
}

func (f *FSMContext) Serializer() StateSerializer {
	return f.fsm.Serializer
}

func (f *FSMContext) Deserialize(serialized string, data interface{}) {
	f.fsm.Deserialize(serialized, data)
}

func (f *FSMContext) EmptyDecisions() []*Decision {
	return f.fsm.EmptyDecisions()
}
