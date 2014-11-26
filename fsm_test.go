package swf

import (
	"log"
	"strconv"
	"testing"
	"time"

	"os"

	"code.google.com/p/go-uuid/uuid"
	"code.google.com/p/goprotobuf/proto"
)

//Todo add tests of error handling mechanism
//assert that the decisions have the mark and the signal external...hmm need workflow id for signal external.

func TestFSM(t *testing.T) {

	fsm := FSM{
		Name:       "test-fsm",
		DataType:   TestData{},
		Serializer: JSONStateSerializer{},
	}

	fsm.AddInitialState(&FSMState{
		Name: "start",
		Decider: func(f *FSMContext, lastEvent HistoryEvent, data interface{}) Outcome {
			testData := data.(*TestData)
			testData.States = append(testData.States, "start")
			serialized := f.Serialize(testData)
			decision := Decision{
				DecisionType: DecisionTypeScheduleActivityTask,
				ScheduleActivityTaskDecisionAttributes: &ScheduleActivityTaskDecisionAttributes{
					ActivityID:   uuid.New(),
					ActivityType: ActivityType{Name: "activity", Version: "activityVersion"},
					TaskList:     &TaskList{Name: "taskList"},
					Input:        serialized,
				},
			}

			return f.Goto("working", testData, []Decision{decision})

		},
	})

	fsm.AddState(&FSMState{
		Name: "working",
		Decider: TypedDecider(func(f *FSMContext, lastEvent HistoryEvent, testData *TestData) Outcome {
			testData.States = append(testData.States, "working")
			serialized := f.Serialize(testData)
			var decisions = f.EmptyDecisions()
			if lastEvent.EventType == EventTypeActivityTaskCompleted {
				decision := Decision{
					DecisionType: DecisionTypeCompleteWorkflowExecution,
					CompleteWorkflowExecutionDecisionAttributes: &CompleteWorkflowExecutionDecisionAttributes{
						Result: serialized,
					},
				}
				decisions = append(decisions, decision)
			} else if lastEvent.EventType == EventTypeActivityTaskFailed {
				decision := Decision{
					DecisionType: DecisionTypeScheduleActivityTask,
					ScheduleActivityTaskDecisionAttributes: &ScheduleActivityTaskDecisionAttributes{
						ActivityID:   uuid.New(),
						ActivityType: ActivityType{Name: "activity", Version: "activityVersion"},
						TaskList:     &TaskList{Name: "taskList"},
						Input:        serialized,
					},
				}
				decisions = append(decisions, decision)
			}

			return f.Stay(testData, decisions)
		}),
	})

	events := []HistoryEvent{
		HistoryEvent{EventType: "DecisionTaskStarted", EventID: 3},
		HistoryEvent{EventType: "DecisionTaskScheduled", EventID: 2},
		HistoryEvent{
			EventID:   1,
			EventType: "WorkflowExecutionStarted",
			WorkflowExecutionStartedEventAttributes: &WorkflowExecutionStartedEventAttributes{
				Input: "{\"States\":[]}",
			},
		},
	}

	first := &PollForDecisionTaskResponse{
		Events:                 events,
		PreviousStartedEventID: 0,
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
		PreviousStartedEventID: 3,
	}

	if state, _ := fsm.findSerializedState(secondEvents); state.StateName != "working" {
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

func Find(decisions []Decision, predicate func(Decision) bool) bool {
	for _, d := range decisions {
		if predicate(d) {
			return true
		}
	}
	return false
}

func stateMarkerPredicate(d Decision) bool {
	return d.DecisionType == "RecordMarker" && d.RecordMarkerDecisionAttributes.MarkerName == StateMarker
}

func scheduleActivityPredicate(d Decision) bool {
	return d.DecisionType == "ScheduleActivityTask"
}

func completeWorkflowPredicate(d Decision) bool {
	return d.DecisionType == "CompleteWorkflowExecution"
}

func DecisionsToEvents(decisions []Decision) []HistoryEvent {
	var events []HistoryEvent
	for _, d := range decisions {
		if scheduleActivityPredicate(d) {
			event := HistoryEvent{
				EventType: "ActivityTaskCompleted",
				EventID:   7,
			}
			events = append(events, event)
			event = HistoryEvent{
				EventType: "ActivityTaskScheduled",
				EventID:   6,
			}
			events = append(events, event)
		}
		if stateMarkerPredicate(d) {
			event := HistoryEvent{
				EventType: "MarkerRecorded",
				EventID:   5,
				MarkerRecordedEventAttributes: &MarkerRecordedEventAttributes{
					MarkerName: StateMarker,
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
	typedDecider := func(f *FSMContext, h HistoryEvent, d TestData) Outcome {
		if d.States[0] != "marshalled" {
			t.Fail()
		}
		return f.Goto("ok", d, nil)
	}

	wrapped := TypedDecider(typedDecider)

	outcome := wrapped(&FSMContext{}, HistoryEvent{}, TestData{States: []string{"marshalled"}})

	if outcome.(TransitionOutcome).state != "ok" {
		t.Fatal("NextState was not 'ok'")
	}
}

func TestPanicRecovery(t *testing.T) {
	s := &FSMState{
		Name: "panic",
		Decider: func(f *FSMContext, e HistoryEvent, data interface{}) Outcome {
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
		Serializer:  JSONStateSerializer{},
		allowPanics: true,
	}

	fsm.AddInitialState(&FSMState{
		Name: "ok",
		Decider: func(f *FSMContext, h HistoryEvent, d interface{}) Outcome {
			if h.EventType == EventTypeWorkflowExecutionSignaled && d.(*TestData).States[0] == "recovered" {
				log.Println("recovered")
				return f.Stay(d, nil)
			}
			t.Fatalf("ok state did not get recovered %s", h)
			return nil

		},
	})

	fsm.AddErrorState(&FSMState{
		Name: "error",
		Decider: func(f *FSMContext, h HistoryEvent, d interface{}) Outcome {
			if h.EventType == EventTypeWorkflowExecutionSignaled && h.WorkflowExecutionSignaledEventAttributes.SignalName == ErrorSignal {
				log.Println("in error recovery")
				return f.Goto("ok", &TestData{States: []string{"recovered"}}, nil)
			}
			t.Fatalf("error handler got unexpected event")
			return nil

		},
	})

	events := []HistoryEvent{
		HistoryEvent{
			EventID:   2,
			EventType: EventTypeWorkflowExecutionSignaled,
			WorkflowExecutionSignaledEventAttributes: &WorkflowExecutionSignaledEventAttributes{
				SignalName: "NOT AN ERROR",
				Input:      "Hi",
			},
		},
		HistoryEvent{
			EventID:   1,
			EventType: EventTypeWorkflowExecutionSignaled,
			WorkflowExecutionSignaledEventAttributes: &WorkflowExecutionSignaledEventAttributes{
				SignalName: ErrorSignal,
				Input:      "{}",
			},
		},
	}

	resp := &PollForDecisionTaskResponse{
		Events:                 events,
		StartedEventID:         1,
		PreviousStartedEventID: 0,
	}

	decisions, _ := fsm.Tick(resp)
	if len(decisions) != 1 && decisions[0].DecisionType != DecisionTypeRecordMarker {
		t.Fatalf("no state marker!")
	}
	//Try with TypedDecider
	id := fsm.initialState.Decider
	fsm.initialState.Decider = TypedDecider(func(f *FSMContext, h HistoryEvent, d *TestData) Outcome { return id(f, h, d) })
	ie := fsm.errorState.Decider
	fsm.errorState.Decider = TypedDecider(func(f *FSMContext, h HistoryEvent, d *TestData) Outcome { return ie(f, h, d) })

	decisions2, _ := fsm.Tick(resp)
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
	Key             *string `protobuf:"bytes,1,req,name=key" json:"key,omitempty"`
	Str             *string `protobuf:"bytes,2,opt,name=str" json:"str,omitempty"`
	XXXunrecognized []byte  `json:"-"`
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
	//the FSM we will create will oscillate between 2 states,
	//waitForSignal -> will wait till the workflow is started or signalled, and update the StateData based on the Hello message received, set a timer, and transition to waitForTimer
	//waitForTimer -> will wait till the timer set by waitForSignal fires, and will signal the workflow with a Hello message, and transition to waitFotSignal
	waitForSignal := func(f *FSMContext, h HistoryEvent, d *StateData) Outcome {
		decisions := f.EmptyDecisions()
		switch h.EventType {
		case EventTypeWorkflowExecutionStarted, EventTypeWorkflowExecutionSignaled:
			if h.EventType == EventTypeWorkflowExecutionSignaled && h.WorkflowExecutionSignaledEventAttributes.SignalName == "hello" {
				hello := &Hello{}
				f.EventData(h, &Hello{})
				d.Count++
				d.Message = hello.Message
			}
			timeoutSeconds := "5" //swf uses stringy numbers in many places
			timerDecision := Decision{
				DecisionType: DecisionTypeStartTimer,
				StartTimerDecisionAttributes: &StartTimerDecisionAttributes{
					StartToFireTimeout: timeoutSeconds,
					TimerID:            "timeToSignal",
				},
			}
			decisions = append(decisions, timerDecision)
			return f.Goto("waitForTimer", d, decisions)
		}
		//if the event was unexpected just stay here
		return f.Stay(d, decisions)

	}

	waitForTimer := func(f *FSMContext, h HistoryEvent, d *StateData) Outcome {
		decisions := f.EmptyDecisions()
		switch h.EventType {
		case EventTypeTimerFired:
			//every time the timer fires, signal the workflow with a Hello
			message := strconv.FormatInt(time.Now().Unix(), 10)
			signalInput := &Hello{message}
			signalDecision := Decision{
				DecisionType: DecisionTypeSignalExternalWorkflowExecution,
				SignalExternalWorkflowExecutionDecisionAttributes: &SignalExternalWorkflowExecutionDecisionAttributes{
					SignalName: "hello",
					Input:      f.Serialize(signalInput),
					RunID:      f.RunID,
					WorkflowID: f.WorkflowID,
				},
			}
			decisions = append(decisions, signalDecision)

			return f.Goto("waitForSignal", d, decisions)
		}
		//if the event was unexpected just stay here
		return f.Stay(d, decisions)
	}

	//create the FSMState by passing the decider function through TypedDecider(),
	//which lets you use d *StateData rather than d interface{} in your decider.
	waitForSignalState := &FSMState{Name: "waitForSignal", Decider: TypedDecider(waitForSignal)}
	waitForTimerState := &FSMState{Name: "waitForTimer", Decider: TypedDecider(waitForTimer)}
	//wire it up in an fsm
	fsm := &FSM{
		Name:       "example-fsm",
		Client:     client,
		DataType:   StateData{},
		Domain:     "exaple-swf-domain",
		TaskList:   "example-decision-task-list-to-poll",
		Serializer: &JSONStateSerializer{},
	}
	//add states to FSM
	fsm.AddInitialState(waitForSignalState)
	fsm.AddState(waitForTimerState)

	//start it up!
	fsm.Start()
}

func TestChildRelator(t *testing.T) {
	start := StartWorkflowRequest{
		WorkflowID: "123",
		WorkflowType: WorkflowType{
			Name:    "test",
			Version: "123",
		},
	}

	relator := new(ChildRelator)

	relator.Relate("child.1", start.WorkflowID, start.WorkflowType)

	serialized, err := JSONStateSerializer{}.Serialize(relator)

	if err != nil {
		t.Fatal(err)
	}

	freshRelator := new(ChildRelator)

	JSONStateSerializer{}.Deserialize(serialized, freshRelator)

	if freshRelator.WorkflowID("child.1") != "123" {
		t.Fatal("id not 123")
	}

	c := freshRelator.WorkflowType("child.1")
	if c.Name != "test" {
		t.Fatal("name not test")
	}
	if c.Version != "123" {
		t.Fatal("version not test")
	}

	if os.Getenv("AWS_ACCESS_KEY_ID") == "" || os.Getenv("AWS_SECRET_ACCESS_KEY") == "" {
		log.Printf("WARNING: NO AWS CREDS SPECIFIED, SKIPPING CLIENTS TEST")
		return
	}

	client := NewClient(MustGetenv("AWS_ACCESS_KEY_ID"), MustGetenv("AWS_SECRET_ACCESS_KEY"), USEast1)
	client.Debug = true

	resp, err := freshRelator.WorkflowExecutionInfo(client, "swf4go", "123")
	if err != nil {
		t.Fatal(err)
	}

	if resp != nil {
		//nil is expected since there shouldnt be an execution with id 123
		t.Fatal(resp)
	}

}

func TestContinuedWorkflows(t *testing.T) {
	fsm := FSM{
		Name:       "test-fsm",
		DataType:   TestData{},
		Serializer: JSONStateSerializer{},
	}

	fsm.AddInitialState(&FSMState{
		Name: "ok",
		Decider: func(f *FSMContext, h HistoryEvent, d interface{}) Outcome {
			if h.EventType == EventTypeWorkflowExecutionSignaled && d.(*TestData).States[0] == "recovered" {
				log.Println("recovered")
				return f.Stay(d, nil)
			}
			t.Fatalf("ok state did not get recovered %s", h)
			return nil

		},
	})

	stateData := fsm.Serialize(TestData{States: []string{"continuing"}})
	state := SerializedState{
		WorkflowEpoch:  3,
		StartedEventId: 77,
		StateName:      "ok",
		StateData:      stateData,
	}
	serializedState := fsm.Serialize(state)
	resp := &PollForDecisionTaskResponse{
		Events: []HistoryEvent{HistoryEvent{
			EventType: EventTypeWorkflowExecutionContinuedAsNew,
			WorkflowExecutionContinuedAsNewEventAttributes: &WorkflowExecutionContinuedAsNewEventAttributes{
				Input: serializedState,
			},
		}},
		StartedEventID: 5,
	}
	_, updatedSerializedState := fsm.Tick(resp)
	updatedState := new(SerializedState)
	fsm.Deserialize(updatedSerializedState, updatedState)

	if updatedState.StartedEventId != 5 {
		t.Fatal("startedEventId != 5")
	}

	if updatedState.WorkflowEpoch != 4 {
		t.Fatal("workflowEpoch != 4")
	}
}
