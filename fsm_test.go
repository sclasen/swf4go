package swf

import (
	"code.google.com/p/go-uuid/uuid"
	"github.com/sclasen/swf4go/Godeps/_workspace/src/code.google.com/p/goprotobuf/proto"
	"log"
	"strconv"
	"testing"
	"time"
)

//Todo add tests of error handling mechanism
//assert that the decisions have the mark and the signal external...hmm need workflow id for signal external.

func TestFSM(t *testing.T) {

	fsm := FSM{
		Name:       "test-fsm",
		DataType:   TestData{},
		Serializer: JsonStateSerializer{},
	}

	fsm.AddInitialState(&FSMState{
		Name: "start",
		Decider: func(f *FSMContext, lastEvent HistoryEvent, data interface{}) *Outcome {
			testData := data.(*TestData)
			testData.States = append(testData.States, "start")
			serialized := f.Serialize(testData)
			decision := &Decision{
				DecisionType: DecisionTypeScheduleActivityTask,
				ScheduleActivityTaskDecisionAttributes: &ScheduleActivityTaskDecisionAttributes{
					ActivityId:   uuid.New(),
					ActivityType: ActivityType{Name: "activity", Version: "activityVersion"},
					TaskList:     &TaskList{Name: "taskList"},
					Input:        serialized,
				},
			}
			return &Outcome{
				NextState: "working",
				Data:      testData,
				Decisions: []*Decision{decision},
			}
		},
	})

	fsm.AddState(&FSMState{
		Name: "working",
		Decider: TypedDecider(func(f *FSMContext, lastEvent HistoryEvent, testData *TestData) *Outcome {
			testData.States = append(testData.States, "working")
			serialized := f.Serialize(testData)
			var decisions = f.EmptyDecisions()
			if lastEvent.EventType == EventTypeActivityTaskCompleted {
				decision := &Decision{
					DecisionType: DecisionTypeCompleteWorkflowExecution,
					CompleteWorkflowExecutionDecisionAttributes: &CompleteWorkflowExecutionDecisionAttributes{
						Result: serialized,
					},
				}
				decisions = append(decisions, decision)
			} else if lastEvent.EventType == EventTypeActivityTaskFailed {
				decision := &Decision{
					DecisionType: DecisionTypeScheduleActivityTask,
					ScheduleActivityTaskDecisionAttributes: &ScheduleActivityTaskDecisionAttributes{
						ActivityId:   uuid.New(),
						ActivityType: ActivityType{Name: "activity", Version: "activityVersion"},
						TaskList:     &TaskList{Name: "taskList"},
						Input:        serialized,
					},
				}
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
	var events []HistoryEvent
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
	typedDecider := func(f *FSMContext, h HistoryEvent, d TestData) *Outcome {
		if d.States[0] != "marshalled" {
			t.Fail()
		}
		return &Outcome{NextState: "ok"}
	}

	wrapped := TypedDecider(typedDecider)

	outcome := wrapped(&FSMContext{}, HistoryEvent{}, TestData{States: []string{"marshalled"}})

	if outcome.NextState != "ok" {
		t.Fatal("NextState was not 'ok'")
	}
}

func TestPanicRecovery(t *testing.T) {
	s := &FSMState{
		Name: "panic",
		Decider: func(f *FSMContext, e HistoryEvent, data interface{}) *Outcome {
			panic("can you handle it?")
		},
	}
	f := &FSM{}
	f.AddInitialState(s)
	_, err := f.panicSafeDecide(s, new(FSMContext), HistoryEvent{}, nil)
	if err == nil {
		t.Fatal("fatallz")
	} else {
		log.Println(err)
	}
}

func TestErrorHandling(t *testing.T) {
	fsm := FSM{
		Name:        "test-fsm",
		DataType:    TestData{},
		Serializer:  JsonStateSerializer{},
		allowPanics: true,
	}

	fsm.AddInitialState(&FSMState{
		Name: "ok",
		Decider: func(f *FSMContext, h HistoryEvent, d interface{}) *Outcome {
			if h.EventType == EventTypeWorkflowExecutionSignaled && d.(*TestData).States[0] == "recovered" {
				log.Println("recovered")
				return &Outcome{NextState: "ok", Data: d}
			}
			t.Fatalf("ok state did not get recovered %s", h)
			return nil

		},
	})

	fsm.AddErrorState(&FSMState{
		Name: "error",
		Decider: func(f *FSMContext, h HistoryEvent, d interface{}) *Outcome {
			if h.EventType == EventTypeWorkflowExecutionSignaled && h.WorkflowExecutionSignaledEventAttributes.SignalName == ERROR_SIGNAL {
				log.Println("in error recovery")
				return &Outcome{
					NextState: "ok",
					Data:      &TestData{States: []string{"recovered"}},
				}
			}
			t.Fatalf("error handler got unexpected event")
			return nil

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
	fsm.initialState.Decider = TypedDecider(func(f *FSMContext, h HistoryEvent, d *TestData) *Outcome { return id(f, h, d) })
	ie := fsm.errorState.Decider
	fsm.errorState.Decider = TypedDecider(func(f *FSMContext, h HistoryEvent, d *TestData) *Outcome { return ie(f, h, d) })

	decisions2 := fsm.Tick(resp)
	if len(decisions2) != 1 && decisions2[0].DecisionType != DecisionTypeRecordMarker {
		t.Fatalf("no state marker for typed decider!")
	}
}

func TestProtobufSerialization(t *testing.T) {
	ser := &ProtobufStateSerializer{}
	key := "FOO"
	val := "BAR"
	init := &ConfigVar{Key: &key, Str: &val}
	serialized, err := ser.Serialize(init)
	if err != nil {
		t.Fatal(err)
	}

	log.Println(serialized)

	deserialized := new(ConfigVar)
	err = ser.Deserialize(serialized, deserialized)
	if err != nil {
		t.Fatal(err)
	}

	if init.GetKey() != deserialized.GetKey() {
		t.Fatal(init, deserialized)
	}

	if init.GetStr() != deserialized.GetStr() {
		t.Fatal(init, deserialized)
	}
}

//This is c&p from som generated code

type ConfigVar struct {
	Key              *string `protobuf:"bytes,1,req,name=key" json:"key,omitempty"`
	Str              *string `protobuf:"bytes,2,opt,name=str" json:"str,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *ConfigVar) Reset()         { *m = ConfigVar{} }
func (m *ConfigVar) String() string { return proto.CompactTextString(m) }
func (*ConfigVar) ProtoMessage()    {}

func (m *ConfigVar) GetKey() string {
	if m != nil && m.Key != nil {
		return *m.Key
	}
	return ""
}

func (m *ConfigVar) GetStr() string {
	if m != nil && m.Str != nil {
		return *m.Str
	}
	return ""
}

func ExampleFSM() {
	// create with swf.NewClient
	var client *Client
	// data type that will be managed by the FSM
	type StateData struct {
		Message string `json:"message,omitempty"`
		Count   int    `json:"count,omitempty"`
	}
	//event type that will be signalled to the FSM with signal name "hello"
	type Hello struct {
		Message string `json:"message,omitempty"`
	}
	//EventData func to give us the type(s) of data expected in the payload of event(s)
	eventData := func(h HistoryEvent) interface{} {
		if h.EventType == EventTypeWorkflowExecutionSignaled && h.WorkflowExecutionSignaledEventAttributes.SignalName == "hello" {
			return new(Hello)
		}
		return nil
	}
	//the FSM we will create will oscillate between 2 states,
    //waitForSignal -> will wait till the workflow is started or signalled, and update the StateData based on the Hello message received, set a timer, and transition to waitForTimer
	//waitForTimer -> will wait till the timer set by waitForSignal fires, and will signal the workflow with a Hello message, and transition to waitFotSignal
	waitForSignal := func(f *FSMContext, h HistoryEvent, d *StateData) *Outcome {
		decisions := f.EmptyDecisions()
		switch h.EventType {
		case EventTypeWorkflowExecutionStarted, EventTypeWorkflowExecutionSignaled:
			if h.EventType == EventTypeWorkflowExecutionSignaled && h.WorkflowExecutionSignaledEventAttributes.SignalName == "hello" {
				hello := f.EventData(h).(*Hello)
				d.Count += 1
				d.Message = hello.Message
			}
			timeoutSeconds := "5" //swf uses stringy numbers in many places
			timerDecision := &Decision{
				DecisionType: DecisionTypeStartTimer,
				StartTimerDecisionAttributes: &StartTimerDecisionAttributes{
					StartToFireTimeout: timeoutSeconds,
					TimerId:            "timeToSignal",
				},
			}
			decisions = append(decisions, timerDecision)
			return &Outcome{NextState: "waitForTimer", Data: d, Decisions: decisions}
		}
		//if the event was unexpected just stay here
		return &Outcome{NextState: "waitForSignal", Data: d, Decisions: decisions}

	}

	waitForTimer := func(f *FSMContext, h HistoryEvent, d *StateData) *Outcome {
		decisions := f.EmptyDecisions()
		switch h.EventType {
		case EventTypeTimerFired:
			//every time the timer fires, signal the workflow with a Hello
			message := strconv.FormatInt(time.Now().Unix(), 10)
			signalInput := &Hello{message}
			signalDecision := &Decision{
				DecisionType: DecisionTypeSignalExternalWorkflowExecution,
				SignalExternalWorkflowExecutionDecisionAttributes: &SignalExternalWorkflowExecutionDecisionAttributes{
					SignalName: "hello",
					Input:      f.Serialize(signalInput),
					RunId:      f.RunId,
					WorkflowId: f.WorkflowId,
				},
			}
			decisions = append(decisions, signalDecision)

			return &Outcome{NextState: "waitForSignal", Data: d, Decisions: decisions}
		}
		//if the event was unexpected just stay here
		return &Outcome{NextState: "waitForTimer", Data: d, Decisions: decisions}
	}


	//create the FSMState by passing the decider function through TypedDecider(),
	//which lets you use d *StateData rather than d interface{} in your decider.
	waitForSignalState := &FSMState{Name: "waitForSignal", Decider: TypedDecider(waitForSignal)}
	waitForTimerState := &FSMState{Name: "waitForTimer", Decider: TypedDecider(waitForTimer)}
	//wire it up in an fsm
	fsm := &FSM{
		Name:          "example-fsm",
		Client:        client,
		DataType:      StateData{},
		EventDataType: eventData,
		Domain:        "exaple-swf-domain",
		TaskList:      "example-decision-task-list-to-poll",
		Serializer:    &JsonStateSerializer{},
	}
    //add states to FSM
	fsm.AddInitialState(waitForSignalState)
	fsm.AddState(waitForTimerState)

	//start it up!
	fsm.Start()
}
