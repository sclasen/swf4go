package swf

import "log"

// PollActivityTaskList returns a started ActivityTaskPoller.
func (c *Client) PollActivityTaskList(domain string, identity string, taskList string) *ActivityTaskPoller {
	poller := &ActivityTaskPoller{
		client:   c,
		Domain:   domain,
		Identity: identity,
		TaskList: taskList,
		Tasks:    make(chan *PollForActivityTaskResponse),
		stop:     make(chan bool, 1),
	}
	poller.start()
	return poller

}

// ActivityTaskPoller polls a given task list in a domain for activity tasks, and sends tasks on its Tasks channel.
type ActivityTaskPoller struct {
	client   *Client
	Identity string
	Domain   string
	TaskList string
	Tasks    chan *PollForActivityTaskResponse
	stop     chan bool
	stopped  chan bool
}

func (p *ActivityTaskPoller) start() {
	go func() {
		for {
			select {
			case <-p.stop:
				return
			default:
				resp, err := p.client.PollForActivityTask(PollForActivityTaskRequest{
					Domain:   p.Domain,
					Identity: p.Identity,
					TaskList: TaskList{Name: p.TaskList},
				})
				if err != nil {
					log.Printf("%s in %+v", err, p)
				} else {
					if resp.TaskToken != "" {
						log.Printf("component=ActivityTaskPoller at=activity-task-recieved activity=%s", resp.ActivityType.Name)
						p.Tasks <- resp
					} else {
						log.Println("component=ActivityTaskPoller at=activity-task-empty-response")
					}
				}
			}

		}
	}()
}

// Stop signals the poller to stop polling after any in-flight poll requests return.
func (p *ActivityTaskPoller) Stop() {
	p.stop <- true
}
