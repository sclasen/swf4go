package swf

import (
	"errors"
	"fmt"
	"log"
	"reflect"
)

// The marker name used then recording the current state and data of a workflow
const (
	STATE_MARKER = "FSM.State"
)

// Decider decides an Outcome based on an event and the current data for an FSM
type Decider func(*FSM, HistoryEvent, interface{}) *Outcome

// EmptyData specifies the type of data used by the FSM, return
type EmptyData func() interface{}

// EmptyInputOrResult
type EmptyInputOrResult func(HistoryEvent) interface{}

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
	Name               string
	Domain             string
	TaskList           string
	Identity           string
	DecisionWorker     *DecisionWorker
	states             map[string]*FSMState
	initialState       *FSMState
	Input              chan *PollForDecisionTaskResponse
	EmptyData          EmptyData
	EmptyInputOrResult EmptyInputOrResult
	stop               chan bool
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

func (f *FSM) Start() {
	if f.initialState == nil {
		panic("No Initial State Defined For FSM")
	}
	go func() {
		poller := f.DecisionWorker.PollTaskList(f.Domain, f.Identity, f.TaskList, f.Input)
		for {
			select {
			case decisionTask, ok := <-f.Input:
				if ok {
					decisions, err := f.Tick(decisionTask)
					if err != nil {
						f.log("action=tick error=tick-failed state=%s", decisionTask)
					} else {
						err = f.DecisionWorker.Decide(decisionTask.TaskToken, decisions)
						if err != nil {
							f.log("action=tick at=decide-request-failed error=%s", err.Error())
							poller.Stop()
							return
						}
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

func (f *FSM) Tick(decisionTask *PollForDecisionTaskResponse) ([]*Decision, error) {
	serializedState, err := f.findSerializedState(decisionTask.Events)
	if err != nil {
		return nil, err
	}

	f.log("action=tick at=find-current-state state=%s", serializedState.State)
	data := f.EmptyData()
	err = f.Serializer().Deserialize(serializedState.Data, data)
	if err != nil {
		f.log("action=tick at=error=deserialize-state-failed")
		return nil, err
	}

	f.log("action=tick at=find-current-data data=%v", data)
	lastEvents, err := f.findLastEvents(decisionTask.PreviousStartedEventId, decisionTask.Events)

	if err != nil {
		return nil, err
	}

	outcome := new(Outcome)
	outcome.Data = data
	outcome.NextState = serializedState.State

	//iterate through events oldest to newest, calling the decider for the current state.
	//if the outcome changes the state use the right FSMState
	for i := len(lastEvents) - 1; i >= 0; i-- {
		e := lastEvents[i]
		f.log("action=tick at=history id=%d type=%s", e.EventId, e.EventType)
		fsmState, ok := f.states[outcome.NextState]
		if ok {
			anOutcome := fsmState.Decider(f, e, outcome.Data)
			f.log("action=tick at=decided-event state=%s next-state=%s decisions=%d", outcome.NextState, anOutcome.NextState, len(anOutcome.Decisions))
			outcome.Data = anOutcome.Data
			outcome.NextState = anOutcome.NextState
			outcome.Decisions = append(outcome.Decisions, anOutcome.Decisions...)
		} else {
			f.log("action=tick error=marked-state-not-in-fsm state=%s", outcome.NextState)
			return nil, errors.New(outcome.NextState + " does not exist")
		}

	}

	f.log("action=tick at=events-processed next-state=%s decisions=%d", outcome.NextState, len(outcome.Decisions))

	for _, d := range outcome.Decisions {
		f.log("action=tick at=decide next-state=%s decision=%s", outcome.NextState, d.DecisionType)
	}

	return f.decisions(outcome)
}

func (f *FSM) EventData(event HistoryEvent) interface{} {
	eventData := f.EmptyInputOrResult(event)

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
			err := f.Serializer().Deserialize(serialized, eventData)
			if err != nil {
				log.Printf("") //TODO propagate here. too much error handling in deciders, maybe make them panic safe.
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
			err := f.Serializer().Deserialize(event.MarkerRecordedEventAttributes.Details, state)
			return state, err
		} else if event.EventType == EventTypeWorkflowExecutionStarted {
			return &SerializedState{State: f.initialState.Name, Data: event.WorkflowExecutionStartedEventAttributes.Input}, nil
		}
	}
	return &SerializedState{}, errors.New("Cant Find Current Data")
}

func (f *FSM) findLastEvents(prevStarted int, events []HistoryEvent) ([]HistoryEvent, error) {
	lastEvents := make([]HistoryEvent, 0)
	for _, event := range events {
		if event.EventId == prevStarted {
			return lastEvents, nil
		} else {
			t := event.EventType
			if t != EventTypeMarkerRecorded &&
				t != EventTypeDecisionTaskScheduled &&
				t != EventTypeDecisionTaskCompleted &&
				t != EventTypeDecisionTaskStarted &&
				t != EventTypeDecisionTaskTimedOut {
				lastEvents = append(lastEvents, event)
			}
		}
	}

	return lastEvents, nil
}

func (f *FSM) decisions(outcome *Outcome) ([]*Decision, error) {

	serializedData, err := f.Serializer().Serialize(outcome.Data)

	state := SerializedState{
		State: outcome.NextState,
		Data:  serializedData,
	}

	d, err := f.DecisionWorker.RecordMarker(STATE_MARKER, state)
	if err != nil {
		return nil, err
	}
	decisions := f.EmptyDecisions()
	decisions = append(decisions, d)

	for _, decision := range outcome.Decisions {
		decisions = append(decisions, decision)
	}
	return decisions, nil
}

func (f *FSM) Stop() {
	f.stop <- true
}

func (f *FSM) isStateMarker(e HistoryEvent) bool {
	return e.EventType == EventTypeMarkerRecorded && e.MarkerRecordedEventAttributes.MarkerName == STATE_MARKER
}

func (f *FSM) Serializer() StateSerializer {
	return f.DecisionWorker.StateSerializer
}

func (f *FSM) EmptyDecisions() []*Decision {
	return make([]*Decision, 0)
}

type SerializedState struct {
	State string `json:"state"`
	Data  string `json:"data"`
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
	return m.v.Call([]reflect.Value{reflect.ValueOf(f), reflect.ValueOf(h), reflect.ValueOf(data)})[0].Interface().(*Outcome)
}
