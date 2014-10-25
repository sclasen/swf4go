package swf

import (
	"errors"
	"log"
	"fmt"
)

/*
FSM start -> decider worker poll tasklist
on event ->
get all events in Reverse until finding latest marker event with MarkerName FSM.State.
and lastest marker event with MarkerName FSM.Data?
If not found, use initial state/start workflow input?
*/

const (
	STATE_MARKER = "FSM.State"
	DATA_MARKER  = "FSM.Data"
)

type Decider func(*FSM, HistoryEvent, interface{}) *Outcome
type EmptyData func() interface{}
type EmptyInputOrResult func(HistoryEvent) interface{}

type Outcome struct {
	Data      interface{}
	NextState string
	Decisions []*Decision
}

type FSMState struct {
	Name    string
	Decider Decider
}

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
	state, err := f.findCurrentState(decisionTask.Events)
	if err != nil {
		return nil, err
	}
	f.log("action=tick at=find-current-state state=%s", state)
	data := f.EmptyData()
	serialized, err := f.findCurrentData(decisionTask.Events)
	if err != nil {
		f.log("action=tick at=error=find-data-failed")
		return nil, err
	}
	err = f.DecisionWorker.StateSerializer.Deserialize(serialized, data)
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
	outcome.NextState = state

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
			return nil, errors.New(outcome.NextState+" does not exist")
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
			err := f.DecisionWorker.StateSerializer.Deserialize(serialized, eventData)
			if err != nil {
				log.Printf("")
			}
		}
	}

	return eventData

}

func (f *FSM)log(format string, data... interface{}){
	actualFormat := fmt.Sprintf("component=FSM name=%s %s", f.Name, format)
	log.Printf(actualFormat, data...)
}

func (f *FSM) findCurrentState(events []HistoryEvent) (string, error) {
	for _, event := range events {
		if f.isStateMarker(event) {
			return event.MarkerRecordedEventAttributes.Details, nil
		}
	}
	return f.initialState.Name, nil
}

//assumes events ordered newest to oldest
func (f *FSM) findCurrentData(events []HistoryEvent) (string, error) {
	for _, event := range events {
		if f.isDataMarker(event) {
			return event.MarkerRecordedEventAttributes.Details, nil
		} else if event.EventType == EventTypeWorkflowExecutionStarted {
			return event.WorkflowExecutionStartedEventAttributes.Input, nil
		}
	}
	return "", errors.New("Cant Find Current Data")
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
	decisions := make([]*Decision, 0)
	decisions = append(decisions, f.DecisionWorker.RecordStringMarker(STATE_MARKER, outcome.NextState))
	dataMarker, err := f.DecisionWorker.RecordMarker(DATA_MARKER, outcome.Data)
	if err != nil {
		return nil, err
	}
	decisions = append(decisions, dataMarker)
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

func (f *FSM) isDataMarker(e HistoryEvent) bool {
	return e.EventType == EventTypeMarkerRecorded && e.MarkerRecordedEventAttributes.MarkerName == DATA_MARKER
}

func (f *FSM) isStateOrDataMarker(e HistoryEvent) bool {
	return f.isStateMarker(e) || f.isDataMarker(e)
}
