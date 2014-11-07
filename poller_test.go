package swf

import (
	"strconv"
	"syscall"
	"testing"
	"time"
)

func TestPollerManager(t *testing.T) {

	mgr := RegisterPollerShutdownManager()
	mgr.exitOnSignal = false

	for i := 1; i < 10; i++ {
		p := TestPoller{strconv.FormatInt(int64(i), 10), make(chan bool, 1), make(chan bool, 1)}
		mgr.Register(p.name, p.stop, p.stopAck)
		mgr.Deregister(p.name)
	}

	for i := 1; i < 10; i++ {
		p := TestPoller{strconv.FormatInt(int64(i), 10), make(chan bool, 1), make(chan bool, 1)}
		go p.eventLoop()
		mgr.Register(p.name, p.stop, p.stopAck)
	}

	mgr.exitChan <- syscall.SIGQUIT
	time.Sleep(1 * time.Second)
}

type TestPoller struct {
	name    string
	stop    chan bool
	stopAck chan bool
}

func (t *TestPoller) eventLoop() {
	<-t.stop
	t.stopAck <- true
}
