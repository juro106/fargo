package watcher

import (
	"time"

	"./filenotify"

	"github.com/fsnotify/fsnotify"
)

// Batcher batches file watch events in a given interval.
type Batcher struct {
	filenotify.FileWatcher
	interval time.Duration
	done     chan struct{}

	Events chan []fsnotify.Event // Events are returned on this channel
}

// New creates and starts a Batcher with the given time interval.
// It will fall back to a poll based watcher if native isn's supported.
// To always use polling, set poll to true.
func New(intervalBatcher, intervalPoll time.Duration, poll bool) (*Batcher, error) {
	var err error
	var watcher filenotify.FileWatcher

	if poll {
		watcher = filenotify.NewPollingWatcher(intervalPoll)
	} else {
		watcher, err = filenotify.New(intervalPoll)
	}

	if err != nil {
		return nil, err
	}

	batcher := &Batcher{}
	batcher.FileWatcher = watcher
	batcher.interval = intervalBatcher
	batcher.done = make(chan struct{}, 1)
	batcher.Events = make(chan []fsnotify.Event, 1)

	if err == nil {
		go batcher.run()
	}

	return batcher, nil
}

func (b *Batcher) run() {
	tick := time.Tick(b.interval)
	evs := make([]fsnotify.Event, 0)
OuterLoop:
	for {
		select {
		case ev := <-b.FileWatcher.Events():
			evs = append(evs, ev)
		case <-tick:
			if len(evs) == 0 {
				continue
			}
			b.Events <- evs
			evs = make([]fsnotify.Event, 0)
		case <-b.done:
			break OuterLoop
		}
	}
	close(b.done)
}

// Close stops the watching of the files.
func (b *Batcher) Close() {
	b.done <- struct{}{}
	b.FileWatcher.Close()
}
