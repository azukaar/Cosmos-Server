package cron

import (
	"sync"
)

type ProcessTracker struct {
	mu     sync.Mutex
	count  int
	zeroCh chan struct{}
}

func NewProcessTracker() *ProcessTracker {
	return &ProcessTracker{
		zeroCh: make(chan struct{}, 1),
	}
}

func (pt *ProcessTracker) StartProcess() {
	pt.mu.Lock()
	pt.count++
	pt.mu.Unlock()
}

func (pt *ProcessTracker) EndProcess() {
	pt.mu.Lock()
	pt.count--
	if pt.count == 0 {
		select {
		case pt.zeroCh <- struct{}{}:
		default:
		}
	}
	pt.mu.Unlock()
}

func (pt *ProcessTracker) WaitForZero() {
	if pt.count == 0 {
		return
	}
	<-pt.zeroCh
}