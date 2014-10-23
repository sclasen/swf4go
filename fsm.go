package swf

import (
	"errors"
	"log"
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

type Decider func(HistoryEvent, interface{}) *Outcome
type EmptyData func() interface{}

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
	Domain         string
	TaskList       string
	Identity       string
	DecisionWorker *DecisionWorker
	states         map[string]*FSMState
	initialState   *FSMState
	Input          chan *PollForDecisionTaskResponse
	EmptyData      EmptyData
	stop           chan bool
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
						log.Printf("component=FSM action=tick error=tick-failed state=%s", decisionTask)
					} else {
						err = f.DecisionWorker.Decide(decisionTask.TaskToken, decisions)
						if err != nil {
							log.Printf("component=FSM action=tick at=decide-request-failed error=%s", err.Error())
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
	log.Printf("component=FSM action=tick at=find-current-state state=%s", state.Name)
	data := f.EmptyData()
	serialized, err := f.findCurrentData(decisionTask.Events)
	if err != nil {
		log.Println("component=FSM action=tick at=error=find-data-failed")
		return nil, err
	}
	err = f.DecisionWorker.StateSerializer.Deserialize(serialized, data)
	if err != nil {
		log.Println("component=FSM action=tick at=error=deserialize-state-failed")
		return nil, err
	}
	log.Printf("component=FSM action=tick at=find-current-data data=%v", data)
	lastEvent, err := f.findLastEvent(decisionTask.Events)
	if err != nil {
		return nil, err
	}
	log.Printf("component=FSM action=tick at=find-last-event event-type=%s", lastEvent.EventType)
	outcome := state.Decider(lastEvent, data)
	for _, d := range outcome.Decisions {
		log.Printf("component=FSM action=tick at=decide next-state=%s decision=%s", outcome.NextState, d.DecisionType)
	}

	return f.decisions(outcome)
}

func (f *FSM) findCurrentState(events []HistoryEvent) (*FSMState, error) {
	for _, event := range events {
		if event.EventType == EventTypeMarkerRecorded && event.MarkerRecordedEventAttributes.MarkerName == STATE_MARKER {
			markerState := event.MarkerRecordedEventAttributes.Details
			state, ok := f.states[markerState]
			if ok {
				return state, nil
			} else {
				log.Printf("component=FSM action=tick error=marked-state-not-in-fsm marker-state=%s", markerState)
				return nil, errors.New(markerState + " does not exist")
			}
		}
	}
	return f.initialState, nil
}

//assumes events ordered newest to oldest
func (f *FSM) findCurrentData(events []HistoryEvent) (string, error) {
	for _, event := range events {
		if event.EventType == EventTypeMarkerRecorded && event.MarkerRecordedEventAttributes.MarkerName == DATA_MARKER {
			return event.MarkerRecordedEventAttributes.Details, nil
		} else if event.EventType == EventTypeWorkflowExecutionStarted {
			return event.WorkflowExecutionStartedEventAttributes.Input, nil
		}
	}
	return "", errors.New("Cant Find Current Data")
}

func (f *FSM) findLastEvent(events []HistoryEvent) (HistoryEvent, error) {
	for _, event := range events {
		t := event.EventType
		if t != EventTypeMarkerRecorded &&
			t != EventTypeDecisionTaskScheduled &&
			t != EventTypeDecisionTaskCompleted &&
			t != EventTypeDecisionTaskStarted &&
			t != EventTypeDecisionTaskTimedOut {
			return event, nil
		}
	}
	log.Printf("component=FSM action=tick error=unable-to-find-last-event events=%d", len(events))
	return HistoryEvent{}, errors.New("Unable To Find last event")
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
