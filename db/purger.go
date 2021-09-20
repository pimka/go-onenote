package db

import (
	"context"
	"time"
)

type NotePurger struct {
	nh          NoteHandler
	off         chan struct{}
	done        chan struct{}
	timeout     time.Duration
	maxErrCount int
}

func NewPurger(nh NoteHandler, timeout time.Duration, maxErrCount int) *NotePurger {
	return &NotePurger{
		nh:          nh,
		off:         make(chan struct{}, 1),
		done:        make(chan struct{}, 1),
		timeout:     timeout,
		maxErrCount: maxErrCount,
	}
}

func (p *NotePurger) Purge(ctx context.Context) {
	t := time.NewTicker(p.timeout)
	go func() {
		defer func() {
			t.Stop()
			close(p.done)
		}()
		errCount := 0

		for {
			select {
			case <-t.C:
				err := p.nh.ClearExpired(ctx)
				if err != nil {
					errCount++
				} else {
					errCount = 0
				}
				if errCount >= p.maxErrCount {
					return
				}
			case <-p.off:
				return
			}
		}
	}()
}

func (p *NotePurger) Stop() chan<- struct{} {
	return p.off
}

func (p *NotePurger) Done() <-chan struct{} {
	return p.done
}
