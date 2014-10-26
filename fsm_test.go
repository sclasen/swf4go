package swf

import (
	"testing"
)

func TestFSM(t *testing.T) {

	fsm := FSM{
		Name:           "test-fsm",
		DecisionWorker: &DecisionWorker{StateSerializer: JsonStateSerializer{}, idGenerator: UUIDGenerator{}},
		EmptyData:      func() interface{} { return &TestData{} },
		states:         make(map[string]*FSMState),
	}

	fsm.AddInitialState(&FSMState{
		Name: "start",
		Decider: func(f *FSM, lastEvent HistoryEvent, data interface{}) *Outcome {
			testData := data.(*TestData)
			testData.States = append(testData.States, "start")
			decision, _ := fsm.DecisionWorker.ScheduleActivityTaskDecision("activity", "activityVersion", "taskList", testData)
			return &Outcome{
				NextState: "working",
				Data:      testData,
				Decisions: []*Decision{decision},
			}
		},
	})

	fsm.AddState(&FSMState{
		Name: "working",
		Decider: func(f *FSM, lastEvent HistoryEvent, data interface{}) *Outcome {
			testData := data.(*TestData)
			testData.States = append(testData.States, "working")
			var decisions = make([]*Decision, 0)
			if lastEvent.EventType == "ActivityTaskCompleted" {
				decision, _ := fsm.DecisionWorker.CompleteWorkflowExecution(testData)
				decisions = append(decisions, decision)
			} else if lastEvent.EventType == "ActivityTaskFailed" {
				decision, _ := fsm.DecisionWorker.ScheduleActivityTaskDecision("activity", "activityVersion", "taskList", testData)
				decisions = append(decisions, decision)
			}
			return &Outcome{
				NextState: "working",
				Data:      testData,
				Decisions: decisions,
			}
		},
	})

	events := []HistoryEvent{
		HistoryEvent{EventType: "DecisionTaskStarted", EventId: 3},
		HistoryEvent{EventType: "DecisionTaskScheduled", EventId: 2},
		HistoryEvent{
			EventId:   1,
			EventType: "WorkflowExecutionStarted",
			WorkflowExecutionStartedEventAttributes: &WorkflowExecutionStartedEventAttributes{
				Input: "{\"States\":[]}",
			},
		},
	}

	first := &PollForDecisionTaskResponse{
		Events:                 events,
		PreviousStartedEventId: 0,
	}

	decisions, _ := fsm.Tick(first)

	if !Find(decisions, stateMarkerPredicate) {
		t.Fatal("No Record State Marker")
	}

	if !Find(decisions, scheduleActivityPredicate) {
		t.Fatal("No ScheduleActivityTask")
	}

	secondEvents := DecisionsToEvents(decisions)
	secondEvents = append(secondEvents, events...)

	second := &PollForDecisionTaskResponse{
		Events:                 secondEvents,
		PreviousStartedEventId: 3,
	}

	if state, _ := fsm.findSerializedState(secondEvents); state.State != "working" {
		t.Fatal("current state is not 'working'", secondEvents)
	}

	secondDecisions, _ := fsm.Tick(second)

	if !Find(secondDecisions, stateMarkerPredicate) {
		t.Fatal("No Record State Marker")
	}

	if !Find(secondDecisions, completeWorkflowPredicate) {
		t.Fatal("No CompleteWorkflow")
	}

}

func Find(decisions []*Decision, predicate func(*Decision) bool) bool {
	for _, d := range decisions {
		if predicate(d) {
			return true
		}
	}
	return false
}

func stateMarkerPredicate(d *Decision) bool {
	return d.DecisionType == "RecordMarker" && d.RecordMarkerDecisionAttributes.MarkerName == STATE_MARKER
}

func scheduleActivityPredicate(d *Decision) bool {
	return d.DecisionType == "ScheduleActivityTask"
}

func completeWorkflowPredicate(d *Decision) bool {
	return d.DecisionType == "CompleteWorkflowExecution"
}

func DecisionsToEvents(decisions []*Decision) []HistoryEvent {
	events := make([]HistoryEvent, 0)
	for _, d := range decisions {
		if scheduleActivityPredicate(d) {
			event := HistoryEvent{
				EventType: "ActivityTaskCompleted",
				EventId:   7,
			}
			events = append(events, event)
			event = HistoryEvent{
				EventType: "ActivityTaskScheduled",
				EventId:   6,
			}
			events = append(events, event)
		}
		if stateMarkerPredicate(d) {
			event := HistoryEvent{
				EventType: "MarkerRecorded",
				EventId:   5,
				MarkerRecordedEventAttributes: &MarkerRecordedEventAttributes{
					MarkerName: STATE_MARKER,
					Details:    d.RecordMarkerDecisionAttributes.Details,
				},
			}
			events = append(events, event)

		}
	}
	return events
}

type TestData struct {
	States []string
}
