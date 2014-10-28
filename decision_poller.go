package swf

import "log"

func (c *Client) PollDecisionTaskList(domain string, identity string, taskList string, taskChannel chan *PollForDecisionTaskResponse) *DecisionTaskPoller {
	poller := &DecisionTaskPoller{
		client:   c,
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
