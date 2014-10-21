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

func (d *DecisionWorker) DecideToScheduleActivity(taskToken string, activityName string, activityVersion string, activityTaskList string, input interface{}) error {
	serialized, err := d.StateSerializer.Serialize(input)
	if err != nil {
		return err
	}

	err = d.client.RespondDecisionTaskCompleted(
		RespondDecisionTaskCompletedRequest{
			Decisions: []Decision{
				Decision{DecisionType: "ScheduleActivityTask",
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
					}},
			},
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
