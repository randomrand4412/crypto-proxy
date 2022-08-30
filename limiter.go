package cryptoproxy

import (
	"time"
)

type Limiter struct {
	maxCount int
	count    int
	ticker   *time.Ticker
	ch       chan struct{}
}

// TODO: add sliding window, or find good lib
func NewLimiter(d time.Duration, count int) *Limiter {
	l := &Limiter{
		maxCount: count,
		count:    count,
		ticker:   time.NewTicker(d),
		ch:       make(chan struct{}),
	}
	go l.run()

	return l
}

func (l *Limiter) Wait() <-chan struct{} {
	return l.ch
}

func (l *Limiter) run() {
	for {
		if l.count <= 0 {
			<-l.ticker.C
			l.count = l.maxCount
		}

		select {
		case l.ch <- struct{}{}:
			l.count--

		case <-l.ticker.C:
			l.count = l.maxCount
		}
	}
}
