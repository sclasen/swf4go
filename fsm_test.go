package swf

import (
	"log"
	"testing"
)

//Todo add tests of error handling mechanism
//assert that the decisions have the mark and the signal external...hmm need workflow id for signal external.

func TestFSM(t *testing.T) {

	fsm := FSM{
		Name:           "test-fsm",
		DecisionWorker: &DecisionWorker{StateSerializer: JsonStateSerializer{}, idGenerator: UUIDGenerator{}},
		DataType:       TestData{},
	}

	fsm.AddInitialState(&FSMState{
		Name: "start",
		Decider: func(f *FSM, lastEvent HistoryEvent, data interface{}) *Outcome {
			testData := data.(*TestData)
			testData.States = append(testData.States, "start")
			decision, _ := f.DecisionWorker.ScheduleActivityTaskDecision("activity", "activityVersion", "taskList", testData)
			return &Outcome{
				NextState: "working",
				Data:      testData,
				Decisions: []*Decision{decision},
			}
		},
	})

	fsm.AddState(&FSMState{
		Name: "working",
		Decider: TypedDecider(func(f *FSM, lastEvent HistoryEvent, testData *TestData) *Outcome {
			testData.States = append(testData.States, "working")
			var decisions = f.EmptyDecisions()
			if lastEvent.EventType == "ActivityTaskCompleted" {
				decision, _ := f.DecisionWorker.CompleteWorkflowExecution(testData)
				decisions = append(decisions, decision)
			} else if lastEvent.EventType == "ActivityTaskFailed" {
				decision, _ := f.DecisionWorker.ScheduleActivityTaskDecision("activity", "activityVersion", "taskList", testData)
				decisions = append(decisions, decision)
			}
			return &Outcome{
				NextState: "working",
				Data:      testData,
				Decisions: decisions,
			}
		}),
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

	decisions := fsm.Tick(first)

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

	if state, _ := fsm.findSerializedState(secondEvents); state.StateName != "working" {
		t.Fatal("current state is not 'working'", secondEvents)
	}

	secondDecisions := fsm.Tick(second)

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

func TestMarshalledDecider(t *testing.T) {
	typedDecider := func(f *FSM, h HistoryEvent, d TestData) *Outcome {
		if d.States[0] != "marshalled" {
			t.Fail()
		}
		return &Outcome{NextState: "ok"}
	}

	wrapped := TypedDecider(typedDecider)

	outcome := wrapped(&FSM{}, HistoryEvent{}, TestData{States: []string{"marshalled"}})

	if outcome.NextState != "ok" {
		t.Fatal("NextState was not 'ok'")
	}
}

func TestPanicRecovery(t *testing.T) {
	s := &FSMState{
		Name: "panic",
		Decider: func(f *FSM, e HistoryEvent, data interface{}) *Outcome {
			panic("can you handle it?")
		},
	}
	f := &FSM{}
	f.AddInitialState(s)
	_, err := f.panicSafeDecide(s, HistoryEvent{}, nil)
	if err == nil {
		t.Fatal("fatallz")
	} else {
		log.Println(err)
	}
}

func TestErrorHandling(t *testing.T) {
	fsm := FSM{
		Name:           "test-fsm",
		DecisionWorker: &DecisionWorker{StateSerializer: JsonStateSerializer{}, idGenerator: UUIDGenerator{}},
		DataType:       TestData{},
		allowPanics:    true,
	}

	fsm.AddInitialState(&FSMState{
		Name: "ok",
		Decider: func(f *FSM, h HistoryEvent, d interface{}) *Outcome {
			if h.EventType == EventTypeWorkflowExecutionSignaled && d.(*TestData).States[0] == "recovered" {
				log.Println("recovered")
				return &Outcome{NextState: "ok", Data: d}
			} else {
				t.Fatalf("ok state did not get recovered %s", h)
				return nil
			}
		},
	})

	fsm.AddErrorState(&FSMState{
		Name: "error",
		Decider: func(f *FSM, h HistoryEvent, d interface{}) *Outcome {
			if h.EventType == EventTypeWorkflowExecutionSignaled && h.WorkflowExecutionSignaledEventAttributes.SignalName == ERROR_SIGNAL {
				log.Println("in error recovery")
				return &Outcome{
					NextState: "ok",
					Data:      &TestData{States: []string{"recovered"}},
				}
			} else {
				t.Fatalf("error handler got unexpected event")
				return nil
			}
		},
	})

	events := []HistoryEvent{
		HistoryEvent{
			EventId:   2,
			EventType: EventTypeWorkflowExecutionSignaled,
			WorkflowExecutionSignaledEventAttributes: &WorkflowExecutionSignaledEventAttributes{
				SignalName: "NOT AN ERROR",
				Input:      "Hi",
			},
		},
		HistoryEvent{
			EventId:   1,
			EventType: EventTypeWorkflowExecutionSignaled,
			WorkflowExecutionSignaledEventAttributes: &WorkflowExecutionSignaledEventAttributes{
				SignalName: ERROR_SIGNAL,
				Input:      "{}",
			},
		},
	}

	resp := &PollForDecisionTaskResponse{
		Events:                 events,
		StartedEventId:         1,
		PreviousStartedEventId: 0,
	}

	decisions := fsm.Tick(resp)
	if len(decisions) != 1 && decisions[0].DecisionType != DecisionTypeRecordMarker {
		t.Fatalf("no state marker!")
	}
	//Try with TypedDecider
	id := fsm.initialState.Decider
	fsm.initialState.Decider = TypedDecider(func(f *FSM, h HistoryEvent, d *TestData) *Outcome { return id(f, h, d) })
	ie := fsm.errorState.Decider
	fsm.errorState.Decider = TypedDecider(func(f *FSM, h HistoryEvent, d *TestData) *Outcome { return ie(f, h, d) })

	decisions2 := fsm.Tick(resp)
	if len(decisions2) != 1 && decisions2[0].DecisionType != DecisionTypeRecordMarker {
		t.Fatalf("no state marker for typed decider!")
	}
}
