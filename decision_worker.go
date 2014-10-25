package swf

import "log"

type DecisionWorker struct {
	client          DecisionWorkerClient
	infoClient      WorkflowInfoClient
	StateSerializer StateSerializer
	idGenerator     IdGenerator
}

func NewDecisionWorker(client *Client, stateSerializer StateSerializer, idGenerator IdGenerator) *DecisionWorker {
	return &DecisionWorker{client: client, infoClient: client, StateSerializer: stateSerializer, idGenerator: idGenerator}
}

func (d *DecisionWorker) ListOpenWorkflowExecutionsByTag(domain string, tag string) ([]WorkflowExecutionInfo, error) {
	resp, err := d.infoClient.ListOpenWorkflowExecutions(ListOpenWorkflowExecutionsRequest{
		Domain:          domain,
		StartTimeFilter: TimeFilter{OldestDate: 0},
		TagFilter:       &TagFilter{Tag: tag},
	})
	if err != nil {
		return nil, err
	}
	return resp.ExecutionInfos, nil
}

func (d *DecisionWorker) ScheduleActivityTaskDecision(activityName string, activityVersion string, activityTaskList string, input interface{}) (*Decision, error) {
	serialized, err := d.StateSerializer.Serialize(input)
	if err != nil {
		return nil, err
	}

	return &Decision{
		DecisionType: DecisionTypeScheduleActivityTask,
		ScheduleActivityTaskDecisionAttributes: &ScheduleActivityTaskDecisionAttributes{
			ActivityId: d.idGenerator.ActivityID(),
			ActivityType: ActivityType{
				Name: activityName, Version: activityVersion,
			},
			Input: serialized,
			TaskList: TaskList{
				Name: activityTaskList,
			},
		},
	}, nil
}

func (d *DecisionWorker) StartChildWorkflowExecutionDecision(workflowType string, workflowVersion string, childPolicy string, taskList string, tags []string, input interface{}) (*Decision, error) {
	return d.StartChildWorkflowExecutionWithIdDecision(workflowType, workflowVersion, childPolicy, taskList, tags, d.idGenerator.ActivityID(), input)
}

func (d *DecisionWorker) StartChildWorkflowExecutionWithIdDecision(workflowType string, workflowVersion string, childPolicy string, taskList string, tags []string, childId string, input interface{}) (*Decision, error) {
	serialized, err := d.StateSerializer.Serialize(input)
	if err != nil {
		return nil, err
	}

	return &Decision{
		DecisionType: DecisionTypeStartChildWorkflowExecution,
		StartChildWorkflowExecutionDecisionAttributes: &StartChildWorkflowExecutionDecisionAttributes{
			ChildPolicy: childPolicy,
			Input:       serialized,
			TagList:     tags,
			TaskList: TaskList{
				Name: taskList,
			},
			ExecutionStartToCloseTimeout: "1000",
			TaskStartToCloseTimeout:      "1000",
			WorkflowType: WorkflowType{
				Name:    workflowType,
				Version: workflowVersion,
			},
			WorkflowId: childId,
		},
	}, nil
}

func (d *DecisionWorker) RecordMarker(markerName string, details interface{}) (*Decision, error) {
	serialized, err := d.StateSerializer.Serialize(details)
	if err != nil {
		return nil, err
	}

	return d.RecordStringMarker(markerName, serialized), nil
}

func (d *DecisionWorker) RecordStringMarker(markerName string, details string) *Decision {
	return &Decision{
		DecisionType: DecisionTypeRecordMarker,
		RecordMarkerDecisionAttributes: &RecordMarkerDecisionAttributes{
			MarkerName: markerName,
			Details:    details,
		},
	}
}

func (d *DecisionWorker) CompleteWorkflowExecution(result interface{}) (*Decision, error) {
	serialized, err := d.StateSerializer.Serialize(result)
	if err != nil {
		return nil, err
	}

	return &Decision{
		DecisionType: DecisionTypeCompleteWorkflowExecution,
		CompleteWorkflowExecutionDecisionAttributes: &CompleteWorkflowExecutionDecisionAttributes{
			Result: serialized,
		},
	}, nil
}

func (d *DecisionWorker) Decide(taskToken string, decisions []*Decision) error {
	err := d.client.RespondDecisionTaskCompleted(
		RespondDecisionTaskCompletedRequest{
			Decisions: decisions,
			TaskToken: taskToken,
		})

	return err
}

func (d *DecisionWorker) PollTaskList(domain string, identity string, taskList string, taskChannel chan *PollForDecisionTaskResponse) *DecisionTaskPoller {
	poller := &DecisionTaskPoller{
		client:   d.client,
		Domain:   domain,
		Identity: identity,
		TaskList: taskList,
		Tasks:    taskChannel,
		stop:     make(chan bool, 1),
	}
	poller.start()
	return poller

}

type DecisionTaskPoller struct {
	client   DecisionWorkerClient
	Identity string
	Domain   string
	TaskList string
	Tasks    chan *PollForDecisionTaskResponse
	stop     chan bool
	Interest chan *bool //tell the poller to go?
}

func (p *DecisionTaskPoller) start() {
	go func() {
		for {
			select {
			case <-p.stop:
				return
			default:
				resp, err := p.client.PollForDecisionTask(PollForDecisionTaskRequest{
					Domain:       p.Domain,
					Identity:     p.Identity,
					ReverseOrder: true,
					TaskList:     TaskList{Name: p.TaskList},
				})
				if err != nil {
					panic(p)
				} else {
					if resp.TaskToken != "" {
						log.Printf("component=DecisionTaskPoller at=decision-task-recieved workflow=%s", resp.WorkflowType.Name)
						p.Tasks <- resp
					} else {
						log.Println("component=DecisionTaskPoller at=decision-task-empty-response")
					}
				}
			}

		}
	}()
}

func (p *DecisionTaskPoller) Stop() {
	p.stop <- true
}
