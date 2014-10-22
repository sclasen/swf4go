package swf

import (
	"testing"
)

func TestFSM(t *testing.T) {

	fsm := FSM{
		DecisionWorker: &DecisionWorker{StateSerializer: JsonStateSerializer{}, idGenerator: UUIDGenerator{}},
		EmptyData:      func() interface{} { return &TestData{} },
		states:         make(map[string]*FSMState),
	}

	fsm.AddInitialState(&FSMState{
		Name: "start",
		Decider: func(lastEvent HistoryEvent, data interface{}) *Outcome {
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
		Decider: func(lastEvent HistoryEvent, data interface{}) *Outcome {
			testData := data.(*TestData)
			testData.States = append(testData.States, "working")
			var decision *Decision
			if lastEvent.EventType == "ActivityTaskCompleted" {
				decision, _ = fsm.DecisionWorker.CompleteWorkflowExecution(testData)
			} else if lastEvent.EventType == "ActivityTaskFailed" {
				decision, _ = fsm.DecisionWorker.ScheduleActivityTaskDecision("activity", "activityVersion", "taskList", testData)
			}
			return &Outcome{
				NextState: "working",
				Data:      testData,
				Decisions: []*Decision{decision},
			}
		},
	})

	events := []HistoryEvent{
		HistoryEvent{EventType: "DecisionTaskStarted"},
		HistoryEvent{EventType: "DecisionTaskScheduled"},
		HistoryEvent{
			EventType: "WorkflowExecutionStarted",
			WorkflowExecutionStartedEventAttributes: &WorkflowExecutionStartedEventAttributes{
				Input: "{\"States\":[]}",
			},
		},
	}

	first := &PollForDecisionTaskResponse{
		Events: events,
	}

	decisions := fsm.Tick(first)

	if !Find(decisions, stateMarkerPredicate) {
		t.Fatal("No Record State Marker")
	}

	if !Find(decisions, dataMarkerPredicate) {
		t.Fatal("No Record Data Marker")
	}

	if !Find(decisions, scheduleActivityPredicate) {
		t.Fatal("No ScheduleActivityTask")
	}

	secondEvents := DecisionsToEvents(decisions)
	secondEvents = append(secondEvents, events...)

	second := &PollForDecisionTaskResponse{
		Events: secondEvents,
	}

	if fsm.findCurrentState(secondEvents).Name != "working" {
		t.Fatal("current state is not 'working'", secondEvents)
	}

	var curr = &TestData{}
	fsm.DecisionWorker.StateSerializer.Deserialize(fsm.findCurrentData(secondEvents), curr)

	if len(curr.States) != 1 && curr.States[0] != "start" {
		t.Fatal("current data is not right", curr.States)
	}

	secondDecisions := fsm.Tick(second)

	if !Find(secondDecisions, stateMarkerPredicate) {
		t.Fatal("No Record State Marker")
	}

	if !Find(secondDecisions, dataMarkerPredicate) {
		t.Fatal("No Record Data Marker")
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

func dataMarkerPredicate(d *Decision) bool {
	return d.DecisionType == "RecordMarker" && d.RecordMarkerDecisionAttributes.MarkerName == DATA_MARKER
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
			}
			events = append(events, event)
			event = HistoryEvent{
				EventType: "ActivityTaskScheduled",
			}
			events = append(events, event)
		}
		if stateMarkerPredicate(d) {
			event := HistoryEvent{
				EventType: "MarkerRecorded",
				MarkerRecordedEventAttributes: &MarkerRecordedEventAttributes{
					MarkerName: STATE_MARKER,
					Details:    d.RecordMarkerDecisionAttributes.Details,
				},
			}
			events = append(events, event)

		}
		if dataMarkerPredicate(d) {
			event := HistoryEvent{
				EventType: "MarkerRecorded",
				MarkerRecordedEventAttributes: &MarkerRecordedEventAttributes{
					MarkerName: DATA_MARKER,
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
