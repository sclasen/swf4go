package swf

import "log"

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

	go func() {
		poller := f.DecisionWorker.PollTaskList(f.Domain, f.Identity, f.TaskList, f.Input)
		for {
			select {
			case decisionTask := <-f.Input:
				decisions := f.Tick(decisionTask)
				f.DecisionWorker.Decide(decisionTask.TaskToken, decisions)
			case <-f.stop:
				poller.Stop()
				break
			}
		}
	}()
}

func (f *FSM) Tick(decisionTask *PollForDecisionTaskResponse) []*Decision {
	state := f.findCurrentState(decisionTask.Events)
	log.Printf("component=FSM action=tick at=find-current-state state=%s", state.Name)
	data := f.EmptyData()
	f.DecisionWorker.StateSerializer.Deserialize(f.findCurrentData(decisionTask.Events), data) //handle err
	log.Printf("component=FSM action=tick at=find-current-data data=%v", data)
	lastEvent := f.findLastEvent(decisionTask.Events)
	log.Printf("component=FSM action=tick at=find-last-event event-type=%s", lastEvent.EventType)
	outcome := state.Decider(lastEvent, data)
	log.Printf("component=FSM action=tick at=decide next-state=%s num-decisions=%d", outcome.NextState, len(outcome.Decisions))
	decisions := f.decisions(outcome)
	return decisions
}

func (f *FSM) findCurrentState(events []HistoryEvent) *FSMState {
	for _, event := range events {
		if event.EventType == "MarkerRecorded" && event.MarkerRecordedEventAttributes.MarkerName == "FSM.State" {
			return f.states[event.MarkerRecordedEventAttributes.Details]
		}
	}
	return f.initialState
}

//assumes events ordered newest to oldest
func (f *FSM) findCurrentData(events []HistoryEvent) string {
	for _, event := range events {
		if event.EventType == "MarkerRecorded" && event.MarkerRecordedEventAttributes.MarkerName == "FSM.Data" {
			return event.MarkerRecordedEventAttributes.Details
		} else if event.EventType == "WorkflowExecutionStarted" {
			return event.WorkflowExecutionStartedEventAttributes.Input
		}
	}
	panic("no FSM.Data and no StartWorkflowEvent")
}

func (f *FSM) findLastEvent(events []HistoryEvent) HistoryEvent {
	for _, event := range events {
		t := event.EventType
		if t != "MarkerRecorded" && t != "DecisionTaskScheduled" && t != "DecisionTaskStarted" {
			return event
		}
	}
	panic("only found MarkerRecorded or DecisionTaskScheduled or DecisionTaskScheduled")
}

func (f *FSM) decisions(outcome *Outcome) []*Decision {
	decisions := make([]*Decision, 0)
	decisions = append(decisions, f.DecisionWorker.RecordStringMarker(STATE_MARKER, outcome.NextState))
	dataMarker, _ := f.DecisionWorker.RecordMarker(DATA_MARKER, outcome.Data)
	decisions = append(decisions, dataMarker)
	for _, decision := range outcome.Decisions {
		decisions = append(decisions, decision)
	}
	for _, decision := range decisions {
		log.Printf(decision.DecisionType)
	}
	return decisions
}

func (f *FSM) Stop() {
	f.stop <- true
}
