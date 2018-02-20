package node

import (
	"time"
)

type AlarmFn func([]interface{})

type Alarm struct {
	fn       AlarmFn
	fnParams []interface{}
	duration time.Duration
	done     chan bool
}

func NewAlarm(fn AlarmFn, fnParams []interface{}, d time.Duration) *Alarm {
	a := &Alarm{fn, fnParams, d, make(chan bool)}

	go a.Run()

	return a
}

func (a *Alarm) Run() {
	ticker := time.NewTicker(a.duration)
	defer ticker.Stop()

	for {
		select {
		case <-a.done:
			return
		case _ = <-ticker.C:
			a.fn(a.fnParams)
		}
	}

}

func (a *Alarm) Stop() {
	a.done <- true
}
