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

type Decider func([]HistoryEvent, interface{}) *Outcome
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

func EventsByType(events []HistoryEvent) map[string][]HistoryEvent {
	byType := make(map[string][]HistoryEvent)
	for _, e := range events {
		es, ok := byType[e.EventType]
		if !ok {
			es = make([]HistoryEvent, 0)
		}
		byType[e.EventType] = append(es, e)
	}
	return byType
}

type FSM struct {
	Name           string
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
						log.Printf("component=FSM name=%s action=tick error=tick-failed state=%s", f.Name, decisionTask)
					} else {
						err = f.DecisionWorker.Decide(decisionTask.TaskToken, decisions)
						if err != nil {
							log.Printf("component=FSM name=%s action=tick at=decide-request-failed error=%s", f.Name, err.Error())
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
	log.Printf("component=FSM name=%s action=tick at=find-current-state state=%s", f.Name, state.Name)
	data := f.EmptyData()
	serialized, err := f.findCurrentData(decisionTask.Events)
	if err != nil {
		log.Printf("component=FSM name=%s action=tick at=error=find-data-failed", f.Name)
		return nil, err
	}
	err = f.DecisionWorker.StateSerializer.Deserialize(serialized, data)
	if err != nil {
		log.Println("component=FSM name=%s action=tick at=error=deserialize-state-failed", f.Name)
		return nil, err
	}
	log.Printf("component=FSM name=%s action=tick at=find-current-data data=%v", f.Name, data)
	lastEvents, err := f.findLastEvents(decisionTask.PreviousStartedEventId, decisionTask.Events)
	if err != nil {
		return nil, err
	}
	for _, e := range lastEvents {
		log.Printf("component=FSM name=%s action=tick at=history id=%d type=%s", f.Name, e.EventId, e.EventType)
	}
	outcome := state.Decider(lastEvents, data)
	log.Printf("component=FSM name=%s action=tick at=decide next-state=%s decisions=%d", f.Name, outcome.NextState, len(outcome.Decisions))
	for _, d := range outcome.Decisions {
		log.Printf("component=FSM name=%s action=tick at=decide next-state=%s decision=%s", f.Name, outcome.NextState, d.DecisionType)
	}

	return f.decisions(outcome)
}

func (f *FSM) findCurrentState(events []HistoryEvent) (*FSMState, error) {
	for _, event := range events {
		if f.isStateMarker(event) {
			markerState := event.MarkerRecordedEventAttributes.Details
			state, ok := f.states[markerState]
			if ok {
				return state, nil
			} else {
				log.Printf("component=FSM name=%s action=tick error=marked-state-not-in-fsm marker-state=%s", f.Name, markerState)
				return nil, errors.New(markerState+" does not exist")
			}
		}
	}
	return f.initialState, nil
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
