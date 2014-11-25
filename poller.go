package swf

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// NewDecisionTaskPoller returns a DecisionTaskPoller whick can be used to poll the given task list.
func NewDecisionTaskPoller(dwc DecisionWorkerClient, domain string, identity string, taskList string) *DecisionTaskPoller {
	return &DecisionTaskPoller{
		client:   dwc,
		Domain:   domain,
		Identity: identity,
		TaskList: taskList,
	}
}

// DecisionTaskPoller polls a given task list in a domain for decision tasks.
type DecisionTaskPoller struct {
	client   DecisionWorkerClient
	Identity string
	Domain   string
	TaskList string
}

// Poll polls the task list for a task. If there is no task available, nil is
// returned. If an error is encountered, no task is returned.
func (p *DecisionTaskPoller) Poll() (*PollForDecisionTaskResponse, error) {
	resp, err := p.client.PollForDecisionTask(PollForDecisionTaskRequest{
		Domain:       p.Domain,
		Identity:     p.Identity,
		ReverseOrder: true,
		TaskList:     TaskList{Name: p.TaskList},
	})
	if err != nil {
		log.Printf("component=DecisionTaskPoller at=error error=%s", err.Error())
		return nil, err
	}
	if resp.TaskToken != "" {
		log.Printf("component=DecisionTaskPoller at=decision-task-recieved workflow=%s", resp.WorkflowType.Name)
		p.logTaskLatency(resp)
		return resp, nil
	}
	log.Println("component=DecisionTaskPoller at=decision-task-empty-response")
	return nil, nil
}

// PollUntilShutdownBy will poll until signaled to shutdown by the PollerShutdownManager. this func blocks, so run it in a goroutine if necessary.
// The implementation calls Poll() and invokes the callback whenever a valid PollForDecisionTaskResponse is received.
func (p *DecisionTaskPoller) PollUntilShutdownBy(mgr *PollerShutdownManager, pollerName string, onTask func(*PollForDecisionTaskResponse)) {
	stop := make(chan bool, 1)
	stopAck := make(chan bool, 1)
	mgr.Register(pollerName, stop, stopAck)
	for {
		select {
		case <-stop:
			log.Printf("component=DecisionTaskPoller fn=PollUntilShutdownBy at=recieved-stop action=shutting-down poller=%s", pollerName)
			stopAck <- true
			return
		default:
			task, err := p.Poll()
			if err != nil {
				log.Printf("component=DecisionTaskPoller fn=PollUntilShutdownBy at=poll-err  poller=%s error=%q", pollerName, err)
				continue
			}
			if task == nil {
				log.Printf("component=DecisionTaskPoller fn=PollUntilShutdownBy at=poll-no-task  poller=%s", pollerName)
				continue
			}
			onTask(task)
		}
	}
}

func (p *DecisionTaskPoller) logTaskLatency(resp *PollForDecisionTaskResponse) {
	for _, e := range resp.Events {
		if e.EventID == resp.StartedEventID {
			elapsed := time.Since(e.EventTimestamp.Time)
			log.Printf("component=DecisionTaskPoller at=decision-task-latency latency=%s workflow=%s", elapsed, resp.WorkflowType.Name)
		}
	}
}

// NewActivityTaskPoller returns an ActivityTaskPoller.
func NewActivityTaskPoller(awc ActivityWorkerClient, domain string, identity string, taskList string) *ActivityTaskPoller {
	return &ActivityTaskPoller{
		client:   awc,
		Domain:   domain,
		Identity: identity,
		TaskList: taskList,
	}
}

// ActivityTaskPoller polls a given task list in a domain for activity tasks, and sends tasks on its Tasks channel.
type ActivityTaskPoller struct {
	client   ActivityWorkerClient
	Identity string
	Domain   string
	TaskList string
}

// Poll polls the task list for a task. If there is no task, nil is returned.
// If an error is encountered, no task is returned.
func (p *ActivityTaskPoller) Poll() (*PollForActivityTaskResponse, error) {
	resp, err := p.client.PollForActivityTask(PollForActivityTaskRequest{
		Domain:   p.Domain,
		Identity: p.Identity,
		TaskList: TaskList{Name: p.TaskList},
	})
	if err != nil {
		log.Printf("component=ActivityTaskPoller at=error error=%s", err.Error())
		return nil, err
	}
	if resp.TaskToken != "" {
		log.Printf("component=ActivityTaskPoller at=activity-task-recieved activity=%s", resp.ActivityType.Name)
		return resp, nil
	}
	log.Println("component=ActivityTaskPoller at=activity-task-empty-response")
	return nil, nil
}

// PollUntilShutdownBy will poll until signaled to shutdown by the PollerShutdownManager. this func blocks, so run it in a goroutine if necessary.
// The implementation calls Poll() and invokes the callback whenever a valid PollForActivityTaskResponse is received.
func (p *ActivityTaskPoller) PollUntilShutdownBy(mgr *PollerShutdownManager, pollerName string, onTask func(*PollForActivityTaskResponse)) {
	stop := make(chan bool, 1)
	stopAck := make(chan bool, 1)
	mgr.Register(pollerName, stop, stopAck)
	for {
		select {
		case <-stop:
			log.Printf("component=ActivityTaskPoller fn=PollUntilShutdownBy at=recieved-stop action=shutting-down poller=%s", pollerName)
			stopAck <- true
			return
		default:
			task, err := p.Poll()
			if err != nil {
				log.Printf("component=ActivityTaskPoller fn=PollUntilShutdownBy at=poll-err  poller=%s error=%q", pollerName, err)
				continue
			}
			if task == nil {
				log.Printf("component=ActivityTaskPoller fn=PollUntilShutdownBy at=poll-no-task  poller=%s", pollerName)
				continue
			}
			onTask(task)
		}
	}
}

// PollerShutdownManager facilitates cleanly shutting down pollers in response to os.Signals before allowing the application to exit. When it receives an os.Signal it will
// send to each of the stopChan that have been registered, then recieve from each of the ackChan that have been registered.  Once this dance is done, it will call system.Exit(0).
type PollerShutdownManager struct {
	exitChan                chan os.Signal
	registeredPollers       map[string]*registeredPoller
	registerChannelsInput   chan *registeredPoller
	deregisterChannelsInput chan string
	exitOnSignal            bool
}

type registeredPoller struct {
	name           string
	stopChannel    chan bool
	stopAckChannel chan bool
}

// RegisterPollerShutdownManager creates a PollerShutdownManager and registers a os.Signal handler. This should be called once per process.
func RegisterPollerShutdownManager() *PollerShutdownManager {

	mgr := &PollerShutdownManager{
		exitChan:                make(chan os.Signal),
		registeredPollers:       make(map[string]*registeredPoller),
		registerChannelsInput:   make(chan *registeredPoller),
		deregisterChannelsInput: make(chan string),
		exitOnSignal:            true,
	}

	signal.Notify(mgr.exitChan, syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	go mgr.eventLoop()

	return mgr

}

func (p *PollerShutdownManager) eventLoop() {
	for {
		select {
		case sig, ok := <-p.exitChan:
			if ok {
				log.Printf("component=PollerShutdownManager at=signal signal=%s", sig)
				for _, r := range p.registeredPollers {
					log.Printf("component=PollerShutdownManager at=sending-stop name=%s", r.name)
					r.stopChannel <- true
				}
				for _, r := range p.registeredPollers {
					log.Printf("component=PollerShutdownManager at=awaiting-stop-ack name=%s", r.name)
					<-r.stopAckChannel
					log.Printf("component=PollerShutdownManager at=stop-ack name=%s", r.name)
				}
				if p.exitOnSignal {
					log.Println("component=PollerShutdownManager at=calling-os.Exit(0)")
					os.Exit(0)
				}
			} else {
				log.Println("component=PollerShutdownManager at=signal-error")
			}
		case register, ok := <-p.registerChannelsInput:
			if ok {
				log.Printf("component=PollerShutdownManager at=register name=%s", register.name)
				p.registeredPollers[register.name] = register
			} else {
				log.Printf("component=PollerShutdownManager at=register-error name=%s", register.name)
			}
		case deregister, ok := <-p.deregisterChannelsInput:
			if ok {
				log.Printf("component=PollerShutdownManager at=deregister name=%s", deregister)
				delete(p.registeredPollers, deregister)
			} else {
				log.Printf("component=PollerShutdownManager at=deregister-error name=%s", deregister)
			}
		}
	}
}

// Register registers a named pair of channels to the shutdown manager. Buffered channels please!
func (p *PollerShutdownManager) Register(name string, stopChan chan bool, ackChan chan bool) {
	p.registerChannelsInput <- &registeredPoller{name, stopChan, ackChan}
}

// Deregister removes a registered pair of channels from the shutdown manager.
func (p *PollerShutdownManager) Deregister(name string) {
	p.deregisterChannelsInput <- name
}
