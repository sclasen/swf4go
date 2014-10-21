package swf

import "log"

type DecisionWorker struct {
	client          *Client
	StateSerializer StateSerializer
	idGenerator     IdGenerator
}

func NewDecisionWorker(client *Client, stateSerializer StateSerializer, idGenerator IdGenerator) *DecisionWorker {
	return &DecisionWorker{client: client, StateSerializer: stateSerializer, idGenerator: idGenerator}
}

func (d *DecisionWorker) ScheduleActivityTaskDecision(activityName string, activityVersion string, activityTaskList string, input interface{}) (*Decision, error) {
	serialized, err := d.StateSerializer.Serialize(input)
	if err != nil {
		return nil, err
	}

	return &Decision{
		DecisionType: "ScheduleActivityTask",
		ScheduleActivityTaskDecisionAttributes: &ScheduleActivityTaskDecisionAttributes{
			ActivityId: d.idGenerator.ActivityID(),
			ActivityType: ActivityType{
				Name: activityName, Version: activityVersion,
			},
			Input: serialized,
			ScheduleToStartTimeout: "NONE",
			ScheduleToCloseTimeout: "NONE",
			StartToCloseTimeout:    "NONE",
			HeartbeatTimeout:       "NONE",
			TaskList: TaskList{
				Name: activityTaskList,
			},
		},
	}, nil
}

func (d *DecisionWorker) StartChildWorkflowExecutionDecision(workflowType string, workflowVersion string, childPolicy string, taskList string, tags []string, input string) (*Decision, error) {
	serialized, err := d.StateSerializer.Serialize(input)
	if err != nil {
		return nil, err
	}

	return &Decision{
		DecisionType: "StartChildWorkflowExecution",
		StartChildWorkflowExecutionDecisionAttributes: &StartChildWorkflowExecutionDecisionAttributes{
			ChildPolicy: childPolicy,
			Input:       serialized,
			TagList:     tags,
			TaskList: TaskList{
				Name: taskList,
			},
			WorkflowType: WorkflowType{
				Name:    workflowType,
				Version: workflowVersion,
			},
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

func (d *DecisionWorker) PollTaskList(domain string, identity string, taskList string) *DecisionTaskPoller {
	poller := &DecisionTaskPoller{
		client:   d.client,
		Domain:   domain,
		Identity: identity,
		TaskList: taskList,
		Tasks:    make(chan *PollForDecisionTaskResponse),
		stop:     make(chan bool, 1),
	}
	poller.start()
	return poller

}

type DecisionTaskPoller struct {
	client   *Client
	Identity string
	Domain   string
	TaskList string
	Tasks    chan *PollForDecisionTaskResponse
	stop     chan bool
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
					log.Printf("%s in %+v", err, p)
				} else {
					if resp.TaskToken != "" {
						p.Tasks <- resp
					}
				}
			}

		}
	}()
}

func (p *DecisionTaskPoller) Stop() {
	p.stop <- true
}
