package swf

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
)

func TestTrackPendingActivities(t *testing.T) {
	fsm := &FSM{
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
					ActivityID:   testActivityInfo.ActivityID,
					ActivityType: testActivityInfo.ActivityType,
					TaskList:     &TaskList{Name: "taskList"},
					Input:        serialized,
				},
			}
			return f.Goto("working", testData, []Decision{decision})
		},
	})

	// Deciders should be able to retrieve info about the pending activity
	fsm.AddState(&FSMState{
		Name: "working",
		Decider: typedFuncs.Decider(func(f *FSMContext, lastEvent HistoryEvent, testData *TestData) Outcome {
			testData.States = append(testData.States, "working")
			serialized := f.Serialize(testData)
			var decisions = f.EmptyDecisions()
			if lastEvent.EventType == EventTypeActivityTaskCompleted {
				trackedActivityInfo := f.ActivityInfo(lastEvent)
				if !reflect.DeepEqual(*trackedActivityInfo, testActivityInfo) {
					t.Fatalf("pending activity not being tracked\nExpected:\n%+v\nGot:\n%+v",
						testActivityInfo, trackedActivityInfo,
					)
				}
				timeoutSeconds := "5" //swf uses stringy numbers in many places
				decision := Decision{
					DecisionType: DecisionTypeStartTimer,
					StartTimerDecisionAttributes: &StartTimerDecisionAttributes{
						StartToFireTimeout: timeoutSeconds,
						TimerID:            "timeToComplete",
					},
				}
				return f.Goto("done", testData, []Decision{decision})
			} else if lastEvent.EventType == EventTypeActivityTaskFailed {
				trackedActivityInfo := f.ActivityInfo(lastEvent)
				if !reflect.DeepEqual(*trackedActivityInfo, testActivityInfo) {
					t.Fatalf("pending activity not being tracked\nExpected:\n%+v\nGot:\n%+v",
						testActivityInfo, trackedActivityInfo,
					)
				}
				decision := Decision{
					DecisionType: DecisionTypeScheduleActivityTask,
					ScheduleActivityTaskDecisionAttributes: &ScheduleActivityTaskDecisionAttributes{
						ActivityID:   testActivityInfo.ActivityID,
						ActivityType: testActivityInfo.ActivityType,
						TaskList:     &TaskList{Name: "taskList"},
						Input:        serialized,
					},
				}
				decisions = append(decisions, decision)
			}
			return f.Stay(testData, decisions)
		}),
	})

	// Pending activities are cleared after finished
	fsm.AddState(&FSMState{
		Name: "done",
		Decider: typedFuncs.Decider(func(f *FSMContext, lastEvent HistoryEvent, testData *TestData) Outcome {
			decisions := f.EmptyDecisions()
			if lastEvent.EventType == EventTypeTimerFired {
				testData.States = append(testData.States, "done")
				serialized := f.Serialize(testData)
				trackedActivityInfo := f.ActivityInfo(lastEvent)
				if trackedActivityInfo != nil {
					t.Fatalf("pending activity not being cleared\nGot:\n%+v", trackedActivityInfo)
				}
				decision := Decision{
					DecisionType: DecisionTypeCompleteWorkflowExecution,
					CompleteWorkflowExecutionDecisionAttributes: &CompleteWorkflowExecutionDecisionAttributes{
						Result: serialized,
					},
				}
				decisions = append(decisions, decision)
			}
			return f.Stay(testData, decisions)
		}),
	})

	// Schedule a task
	events := []HistoryEvent{
		HistoryEvent{EventType: "DecisionTaskStarted", EventID: 3},
		HistoryEvent{EventType: "DecisionTaskScheduled", EventID: 2},
		HistoryEvent{
			EventID:   1,
			EventType: "WorkflowExecutionStarted",
			WorkflowExecutionStartedEventAttributes: &WorkflowExecutionStartedEventAttributes{
				Input: StartFSMWorkflowInput(fsm.Serializer, new(TestData)),
			},
		},
	}
	first := &PollForDecisionTaskResponse{
		Events:                 events,
		PreviousStartedEventID: 0,
	}
	decisions, _ := fsm.Tick(first)
	recordMarker := FindDecision(decisions, stateMarkerPredicate)
	if recordMarker == nil {
		t.Fatal("No Record State Marker")
	}
	if !Find(decisions, scheduleActivityPredicate) {
		t.Fatal("No ScheduleActivityTask")
	}

	// Fail the task
	secondEvents := []HistoryEvent{
		{
			EventType: "ActivityTaskFailed",
			EventID:   7,
			ActivityTaskFailedEventAttributes: &ActivityTaskFailedEventAttributes{
				ScheduledEventID: 6,
			},
		},
		{
			EventType: "ActivityTaskScheduled",
			EventID:   6,
			ActivityTaskScheduledEventAttributes: &ActivityTaskScheduledEventAttributes{
				ActivityID:   testActivityInfo.ActivityID,
				ActivityType: testActivityInfo.ActivityType,
			},
		},
		{
			EventType: "MarkerRecorded",
			EventID:   5,
			MarkerRecordedEventAttributes: &MarkerRecordedEventAttributes{
				MarkerName: StateMarker,
				Details:    recordMarker.RecordMarkerDecisionAttributes.Details,
			},
		},
	}
	secondEvents = append(secondEvents, events...)
	if state, _ := fsm.findSerializedState(secondEvents); state.ReplicationData.StateName != "working" {
		t.Fatal("current state is not 'working'", secondEvents)
	}
	second := &PollForDecisionTaskResponse{
		Events:                 secondEvents,
		PreviousStartedEventID: 3,
	}
	secondDecisions, _ := fsm.Tick(second)
	recordMarker = FindDecision(secondDecisions, stateMarkerPredicate)
	if recordMarker == nil {
		t.Fatal("No Record State Marker")
	}
	if !Find(secondDecisions, scheduleActivityPredicate) {
		t.Fatal("No ScheduleActivityTask (retry)")
	}

	// Complete the task
	thirdEvents := []HistoryEvent{
		{
			EventType: "ActivityTaskCompleted",
			EventID:   11,
			ActivityTaskCompletedEventAttributes: &ActivityTaskCompletedEventAttributes{
				ScheduledEventID: 10,
			},
		},
		{
			EventType: "ActivityTaskScheduled",
			EventID:   10,
			ActivityTaskScheduledEventAttributes: &ActivityTaskScheduledEventAttributes{
				ActivityID:   testActivityInfo.ActivityID,
				ActivityType: testActivityInfo.ActivityType,
			},
		},
		{
			EventType: "MarkerRecorded",
			EventID:   9,
			MarkerRecordedEventAttributes: &MarkerRecordedEventAttributes{
				MarkerName: StateMarker,
				Details:    recordMarker.RecordMarkerDecisionAttributes.Details,
			},
		},
	}
	thirdEvents = append(thirdEvents, secondEvents...)
	if state, _ := fsm.findSerializedState(thirdEvents); state.ReplicationData.StateName != "working" {
		t.Fatal("current state is not 'working'", thirdEvents)
	}
	third := &PollForDecisionTaskResponse{
		Events:                 thirdEvents,
		PreviousStartedEventID: 7,
	}
	thirdDecisions, _ := fsm.Tick(third)
	recordMarker = FindDecision(thirdDecisions, stateMarkerPredicate)
	if recordMarker == nil {
		t.Fatal("No Record State Marker")
	}
	if !Find(thirdDecisions, startTimerPredicate) {
		t.Fatal("No StartTimer")
	}

	// Finish the workflow, check if pending activities were cleared
	fourthEvents := []HistoryEvent{
		{
			EventType: "TimerFired",
			EventID:   14,
		},
		{
			EventType: "MarkerRecorded",
			EventID:   13,
			MarkerRecordedEventAttributes: &MarkerRecordedEventAttributes{
				MarkerName: StateMarker,
				Details:    recordMarker.RecordMarkerDecisionAttributes.Details,
			},
		},
	}
	fourthEvents = append(fourthEvents, thirdEvents...)
	if state, _ := fsm.findSerializedState(fourthEvents); state.ReplicationData.StateName != "done" {
		t.Fatal("current state is not 'done'", fourthEvents)
	}
	fourth := &PollForDecisionTaskResponse{
		Events:                 fourthEvents,
		PreviousStartedEventID: 11,
	}
	fourthDecisions, _ := fsm.Tick(fourth)
	recordMarker = FindDecision(fourthDecisions, stateMarkerPredicate)
	if recordMarker == nil {
		t.Fatal("No Record State Marker")
	}
	if !Find(fourthDecisions, completeWorkflowPredicate) {
		t.Fatal("No CompleteWorkflow")
	}
}

func TestFSMContextActivityTracking(t *testing.T) {
	ctx := NewFSMContext(
		&FSM{
			Name:       "test-fsm",
			DataType:   TestData{},
			Serializer: JSONStateSerializer{},
		},
		WorkflowType{Name: "test-workflow", Version: "1"},
		WorkflowExecution{WorkflowID: "test-workflow-1", RunID: "123123"},
		&ActivityCorrelator{},
		"InitialState", &TestData{}, 0,
	)
	scheduledEventID := rand.Int()
	activityID := fmt.Sprintf("test-activity-%d", scheduledEventID)
	taskScheduled := HistoryEvent{
		EventType: "ActivityTaskScheduled",
		EventID:   scheduledEventID,
		ActivityTaskScheduledEventAttributes: &ActivityTaskScheduledEventAttributes{
			ActivityID: activityID,
			ActivityType: ActivityType{
				Name:    "test-activity",
				Version: "1",
			},
		},
	}
	ctx.Decide(taskScheduled, &TestData{}, func(ctx *FSMContext, h HistoryEvent, data interface{}) Outcome {
		if len(ctx.ActivitiesInfo()) != 0 {
			t.Fatal("There should be no activies being tracked yet")
		}
		if !reflect.DeepEqual(h, taskScheduled) {
			t.Fatal("Got an unexpected event")
		}
		return ctx.Stay(data, ctx.EmptyDecisions())
	})
	if len(ctx.ActivitiesInfo()) < 1 {
		t.Fatal("Pending activity task is not being tracked")
	}

	// the pending activity can now be retrieved
	taskCompleted := HistoryEvent{
		EventType: "ActivityTaskCompleted",
		EventID:   rand.Int(),
		ActivityTaskCompletedEventAttributes: &ActivityTaskCompletedEventAttributes{
			ScheduledEventID: scheduledEventID,
		},
	}
	taskFailed := HistoryEvent{
		EventType: "ActivityTaskFailed",
		EventID:   rand.Int(),
		ActivityTaskFailedEventAttributes: &ActivityTaskFailedEventAttributes{
			ScheduledEventID: scheduledEventID,
		},
	}
	infoOnCompleted := ctx.ActivityInfo(taskCompleted)
	infoOnFailed := ctx.ActivityInfo(taskFailed)
	if infoOnCompleted.ActivityID != activityID ||
		infoOnCompleted.Name != "test-activity" ||
		infoOnCompleted.Version != "1" {
		t.Fatal("Pending activity can not be retrieved when completed")
	}
	if infoOnFailed.ActivityID != activityID ||
		infoOnFailed.Name != "test-activity" ||
		infoOnFailed.Version != "1" {
		t.Fatal("Pending activity can not be retrieved when failed")
	}

	// pending activities are also cleared after terminated
	ctx.Decide(taskCompleted, &TestData{}, func(ctx *FSMContext, h HistoryEvent, data interface{}) Outcome {
		if len(ctx.ActivitiesInfo()) != 1 {
			t.Fatal("There should be one activity being tracked")
		}
		if !reflect.DeepEqual(h, taskCompleted) {
			t.Fatal("Got an unexpected event")
		}
		infoOnCompleted := ctx.ActivityInfo(taskCompleted)
		if infoOnCompleted.ActivityID != activityID ||
			infoOnCompleted.Name != "test-activity" ||
			infoOnCompleted.Version != "1" {
			t.Fatal("Pending activity can not be retrieved when completed")
		}
		return ctx.Stay(data, ctx.EmptyDecisions())
	})
	if len(ctx.ActivitiesInfo()) > 0 {
		t.Fatal("Pending activity task is not being cleared after completed")
	}
}

func TestCountActivityAttemtps(t *testing.T) {
	c := new(ActivityCorrelator)
	start := func(eventId int) HistoryEvent {
		return HistoryEvent{
			EventID:   eventId,
			EventType: EventTypeActivityTaskScheduled,
			ActivityTaskScheduledEventAttributes: &ActivityTaskScheduledEventAttributes{
				ActivityID: "the-id",
			},
		}
	}
	fail := HistoryEvent{
		EventID:   2,
		EventType: EventTypeActivityTaskFailed,
		ActivityTaskFailedEventAttributes: &ActivityTaskFailedEventAttributes{
			ScheduledEventID: 1,
		},
	}

	timeout := HistoryEvent{
		EventID:   4,
		EventType: EventTypeActivityTaskTimedOut,
		ActivityTaskTimedOutEventAttributes: &ActivityTaskTimedOutEventAttributes{
			ScheduledEventID: 3,
		},
	}

	c.Track(start(1))
	c.Track(fail)
	c.Track(start(3))
	c.Track(timeout)

	if c.AttemptsForID("the-id") != 2 {
		t.Fatal(c)
	}

	cancel := HistoryEvent{
		EventID:   6,
		EventType: EventTypeActivityTaskCanceled,
		ActivityTaskCanceledEventAttributes: &ActivityTaskCanceledEventAttributes{
			ScheduledEventID: 5,
		},
	}

	c.Track(start(5))
	c.Track(cancel)

	if c.AttemptsForID("the-id") != 0 {
		t.Fatal(c)
	}

	c.Track(start(7))

	if c.AttemptsForID("the-id") != 0 {
		t.Fatal(c)
	}
}
