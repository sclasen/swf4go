package swf

import (
	"bytes"
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
type Decider func(*FSM, HistoryEvent, interface{}) *Outcome

// EventDataType should return an empty struct of the correct type based on the event
// the FSM will unmarshal data from the event into this struct
type EventDataType func(HistoryEvent) interface{}

// Outcome is created by Deciders
type Outcome struct {
	Data      interface{}
	NextState string
	Decisions []*Decision
}

// FSMState defines the behavior of one state of an FSM
type FSMState struct {
	Name    string
	Decider Decider
}

// FSM models the decision handling logic a workflow in SWF
type FSM struct {
	Name          string
	Domain        string
	TaskList      string
	Identity      string
	Client        *Client
	Input         chan *PollForDecisionTaskResponse
	DataType      interface{}
	EventDataType EventDataType
	Serializer    StateSerializer
	states        map[string]*FSMState
	initialState  *FSMState
	errorState    *FSMState
	stop          chan bool
	allowPanics   bool //makes testing easier
}

func (f *FSM) AddInitialState(state *FSMState) {
	f.AddState(state)
	f.initialState = state
}
func (f *FSM) AddState(state *FSMState) {
	if f.states == nil {
		f.states = make(map[string]*FSMState)
	}
	f.states[state.Name] = state
}

//ErrorStates should return an outcome with nil Data and "" as NextState if they wish for the normal state recovery mechanism
//to load current state and data.

func (f *FSM) AddErrorState(state *FSMState) {
	f.AddState(state)
	f.errorState = state
}

func (f *FSM) DefaultErrorState() *FSMState {
	return &FSMState{
		Name: "error",
		Decider: func(f *FSM, h HistoryEvent, data interface{}) *Outcome {
			switch h.EventType {
			case EventTypeWorkflowExecutionSignaled:
				{
					switch h.WorkflowExecutionSignaledEventAttributes.SignalName {
					case ERROR_SIGNAL:
						err := &SerializedDecisionError{}
						f.Serializer.Deserialize(h.WorkflowExecutionSignaledEventAttributes.Input, err)
						f.log("action=default-handle-error at=handle-decision-error error=%+v", err)
						f.log("YOU SHOULD CREATE AN ERROR STATE FOR YOUR FSM, Workflow %s is Hung", h.WorkflowExecutionSignaledEventAttributes.ExternalWorkflowExecution.WorkflowId)
					case SYSTEM_ERROR_SIGNAL:
						err := &SerializedSystemError{}
						f.Serializer.Deserialize(h.WorkflowExecutionSignaledEventAttributes.Input, err)
						f.log("action=default-handle-error at=handle-system-error error=%+v", err)
						f.log("YOU SHOULD CREATE AN ERROR STATE FOR YOUR FSM, Workflow %s is Hung", h.WorkflowExecutionSignaledEventAttributes.ExternalWorkflowExecution.WorkflowId)
					default:
						f.log("action=default-handle-error at=process-signal-event event=%+v", h)
					}

				}
			default:
				f.log("action=default-handle-error at=process-event event=%+v", h)
			}
			return &Outcome{NextState: "error", Data: data, Decisions: []*Decision{}}
		},
	}
}

func (f *FSM) Start() {
	if f.initialState == nil {
		panic("No Initial State Defined For FSM")
	}

	if f.errorState == nil {
		f.AddErrorState(f.DefaultErrorState())
	}

	if f.stop == nil {
		f.stop = make(chan bool)
	}

	if f.Serializer == nil {
		f.log("action=start at=no-serializer defaulting-to=JsonSerializer")
		f.Serializer = &JsonStateSerializer{}
	}

	poller := f.Client.PollDecisionTaskList(f.Domain, f.Identity, f.TaskList, f.Input)
	go func() {
		for {
			select {
			case decisionTask, ok := <-f.Input:
				if ok {
					decisions := f.Tick(decisionTask)
					err := f.Client.RespondDecisionTaskCompleted(
						RespondDecisionTaskCompletedRequest{
							Decisions: decisions,
							TaskToken: decisionTask.TaskToken,
						})

					if err != nil {
						f.log("action=tick at=decide-request-failed error=%s", err.Error())
						//TODO Retry the Decide?
						poller.Stop()
						return
					}

				} else {
					poller.Stop()
					return
				}
			case <-f.stop:
				poller.Stop()
				return
			}
		}
	}()
}

func (f *FSM) Tick(decisionTask *PollForDecisionTaskResponse) []*Decision {
	lastEvents, errorEvents := f.findLastEvents(decisionTask.PreviousStartedEventId, decisionTask.Events)
	execution := decisionTask.WorkflowExecution
	outcome := new(Outcome)

	//if there are error events, we dont do normal recovery of state + data, we expect the error state to provide this.
	if len(errorEvents) > 0 {
		outcome.Data = reflect.New(reflect.TypeOf(f.DataType)).Interface()
		outcome.NextState = f.errorState.Name
		for i := len(errorEvents) - 1; i >= 0; i-- {
			e := errorEvents[i]
			anOutcome, err := f.panicSafeDecide(f.errorState, e, outcome.Data)
			if err != nil {
				f.log("at=error error=error-handling-decision-execution-error state=%s next-state=%", f.errorState.Name, outcome.NextState)
				//we wont get here if panics are allowed
				return append(outcome.Decisions, f.captureDecisionError(execution, i, errorEvents, outcome.NextState, outcome.Data, err)...)
			}
			outcome.Data = anOutcome.Data
			outcome.NextState = anOutcome.NextState
			outcome.Decisions = append(outcome.Decisions, anOutcome.Decisions...)
		}

	}

	//if there was no error processing, we recover the state + data from the expected marker
	if outcome.Data == nil && outcome.NextState == "" {
		serializedState, err := f.findSerializedState(decisionTask.Events)

		if err != nil {
			f.log("action=tick at=error=find-serialized-state-failed")
			if f.allowPanics {
				panic(err)
			}
			return append(outcome.Decisions, f.captureSystemError(execution, "FindSerializedStateError", decisionTask.Events, err)...)
		}

		f.log("action=tick at=find-current-state state=%s", serializedState.StateName)
		data := reflect.New(reflect.TypeOf(f.DataType)).Interface()
		err = f.Serializer.Deserialize(serializedState.StateData, data)
		if err != nil {
			f.log("action=tick at=error=deserialize-state-failed")
			if f.allowPanics {
				panic(err)
			}
			return append(outcome.Decisions, f.captureSystemError(execution, "DeserializeStateError", decisionTask.Events, err)...)
		}

		f.log("action=tick at=find-current-data data=%v", data)

		outcome.Data = data
		outcome.NextState = serializedState.StateName
	}

	//iterate through events oldest to newest, calling the decider for the current state.
	//if the outcome changes the state use the right FSMState
	for i := len(lastEvents) - 1; i >= 0; i-- {
		e := lastEvents[i]
		f.log("action=tick at=history id=%d type=%s", e.EventId, e.EventType)
		fsmState, ok := f.states[outcome.NextState]
		if ok {
			anOutcome, err := f.panicSafeDecide(fsmState, e, outcome.Data)
			if err != nil {
				f.log("at=error error=decision-execution-error state=%s next-state=%", fsmState.Name, outcome.NextState)
				if f.allowPanics {
					panic(err)
				}
				return append(outcome.Decisions, f.captureDecisionError(execution, i, lastEvents, outcome.NextState, outcome.Data, err)...)
			}

			f.log("action=tick at=decided-event state=%s next-state=%s decisions=%d", outcome.NextState, anOutcome.NextState, len(anOutcome.Decisions))
			outcome.Data = anOutcome.Data
			outcome.NextState = anOutcome.NextState
			outcome.Decisions = append(outcome.Decisions, anOutcome.Decisions...)
		} else {
			f.log("action=tick at=error error=marked-state-not-in-fsm state=%s", outcome.NextState)
			return append(outcome.Decisions, f.captureSystemError(execution, "MissingFsmStateError", lastEvents[i:], errors.New(outcome.NextState))...)
		}
	}

	f.log("action=tick at=events-processed next-state=%s decisions=%d", outcome.NextState, len(outcome.Decisions))

	for _, d := range outcome.Decisions {
		f.log("action=tick at=decide next-state=%s decision=%s", outcome.NextState, d.DecisionType)
	}

	final, err := f.appendState(outcome)
	if err != nil {
		f.log("action=tick at=error error=state-serialization-error error-type=system")
		if f.allowPanics {
			panic(err)
		}
		return append(outcome.Decisions, f.captureSystemError(execution, "StateSerializationError", []HistoryEvent{}, err)...)
	}
	return final
}

//if the outcome is good good if its an error, we capture the error state above

func (f *FSM) panicSafeDecide(state *FSMState, event HistoryEvent, data interface{}) (anOutcome *Outcome, anErr error) {
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
	anOutcome = state.Decider(f, event, data)
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
			err := f.Serializer.Deserialize(serialized, eventData)
			if err != nil {
				f.log("action=EventData at=error error=unable-to-deserialize")
				panic("Unable to Deserialize Event Data")
			}
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
	lastEvents := make([]HistoryEvent, 0)
	errorEvents := make([]HistoryEvent, 0)

	for _, event := range events {
		if event.EventId == prevStarted {
			return lastEvents, errorEvents
		} else {
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
	}

	return lastEvents, errorEvents
}

func (f *FSM) appendState(outcome *Outcome) ([]*Decision, error) {

	serializedData, err := f.Serializer.Serialize(outcome.Data)

	state := SerializedState{
		StateName: outcome.NextState,
		StateData: serializedData,
	}

	d, err := f.recordMarker(STATE_MARKER, state)
	if err != nil {
		return nil, err
	}
	decisions := f.EmptyDecisions()
	decisions = append(decisions, d)
	decisions = append(decisions, outcome.Decisions...)
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

func (f *FSM) EmptyDecisions() []*Decision {
	return make([]*Decision, 0)
}

type SerializedState struct {
	StateName string `json:"stateName"`
	StateData string `json:"stateData"`
}

type SerializedDecisionError struct {
	ErrorEventId        int         `json:"errorEventIds"`
	UnprocessedEventIds []int       `json:"unprocessedEventIds"`
	StateName           string      `json:"stateName"`
	StateData           interface{} `json:"stateData"`
}

type SerializedSystemError struct {
	ErrorType           string      `json:"errorType"`
	Error               interface{} `json:"error"`
	UnprocessedEventIds []int       `json:"unprocessedEventIds"`
}

type DecisionErrorPointer struct {
	Error error
}

type StateSerializer interface {
	Serialize(state interface{}) (string, error)
	Deserialize(serialized string, state interface{}) error
}

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

/*tigertonic-like marhslling of data*/
type MarshalledDecider struct {
	v reflect.Value
}

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
	if "*swf.FSM" != t.In(0).String() {
		panic(fmt.Sprintf(
			"type of first argument was %v, not *swf.FSM",
			t.In(0),
		))
	}
	if "swf.HistoryEvent" != t.In(1).String() {
		panic(fmt.Sprintf(
			"type of second argument was %v, not swf.HistoryEvent",
			t.In(1),
		))
	}

	if "*swf.Outcome" != t.Out(0).String() {
		panic(fmt.Sprintf(
			"type of return value was %v, not *swf.Outcome",
			t.Out(0),
		))
	}

	return MarshalledDecider{reflect.ValueOf(decider)}.Decide
}

func (m MarshalledDecider) Decide(f *FSM, h HistoryEvent, data interface{}) *Outcome {

	// reflection will asplode if we try to use nil
	if data == nil {
		data = reflect.New(reflect.TypeOf(f.DataType)).Interface()
	}
	return m.v.Call([]reflect.Value{reflect.ValueOf(f), reflect.ValueOf(h), reflect.ValueOf(data)})[0].Interface().(*Outcome)
}
