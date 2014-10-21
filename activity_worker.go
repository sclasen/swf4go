package swf

import "log"

type ActivityWorker struct {
	client          *Client
	StateSerializer StateSerializer
	idGenerator     IdGenerator
}

func NewActivityWorker(client *Client, stateSerializer StateSerializer, idGenerator IdGenerator) *ActivityWorker {
	return &ActivityWorker{client: client, StateSerializer: stateSerializer, idGenerator: idGenerator}
}

func (a *ActivityWorker) CompleteActivity(taskToken string, result interface{}) error {
	serialized, err := a.StateSerializer.Serialize(result)
	if err != nil {
		return err
	}

	return a.client.RespondActivityTaskCompleted(RespondActivityTaskCompletedRequest{
		TaskToken: taskToken,
		Result:    serialized,
	})
}

func (d *ActivityWorker) PollTaskList(domain string, identity string, taskList string) *ActivityTaskPoller {
	poller := &ActivityTaskPoller{
		client:   d.client,
		Domain:   domain,
		Identity: identity,
		TaskList: taskList,
		Tasks:    make(chan *PollForActivityTaskResponse),
		stop:     make(chan bool, 1),
	}
	poller.start()
	return poller

}

type ActivityTaskPoller struct {
	client   *Client
	Identity string
	Domain   string
	TaskList string
	Tasks    chan *PollForActivityTaskResponse
	stop     chan bool
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
						p.Tasks <- resp
					}
				}
			}

		}
	}()
}

func (p *ActivityTaskPoller) Stop() {
	p.stop <- true
}
