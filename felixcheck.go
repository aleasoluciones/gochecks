package gochecks

import (
	"time"

	"github.com/aleasoluciones/goaleasoluciones/scheduledtask"
)

type Event struct {
	Host        string
	Service     string
	State       string
	Metric      interface{}
	Description string
	Tags        []string
	Attributes  map[string]string
	TTL         float32
}

type CheckEngine struct {
	checkPublishers []CheckPublisher
	results         chan Event
}

func NewCheckEngine(publishers []CheckPublisher) *CheckEngine {
	checkEngine := CheckEngine{publishers, make(chan Event)}
	go func() {
		for result := range checkEngine.results {
			for _, publisher := range checkEngine.checkPublishers {
				publisher.PublishCheckResult(result)
			}
		}
	}()
	return &checkEngine
}

func (ce *CheckEngine) AddResult(event Event) {
	ce.results <- event
}

func (ce *CheckEngine) AddCheck(check CheckFunction, period time.Duration) {
	scheduledtask.NewScheduledTask(func() {
		ce.results <- check()
	}, period, 0)
}

func (ce *CheckEngine) AddMultiCheck(check MultiCheckFunction, period time.Duration) {
	scheduledtask.NewScheduledTask(func() {
		for _, result := range check() {
			ce.results <- result
		}
	}, period, 0)
}
