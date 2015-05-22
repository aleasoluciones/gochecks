package felixcheck

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
	Ttl         float32
}

type CheckEngine struct {
	checkPublisher CheckPublisher
	results        chan Event
}

func NewCheckEngine(checkPublisher CheckPublisher) CheckEngine {
	checkEngine := CheckEngine{checkPublisher, make(chan Event)}
	go func() {
		for result := range checkEngine.results {
			checkEngine.checkPublisher.PublishCheckResult(result)
		}
	}()
	return checkEngine
}

func (ce CheckEngine) AddCheck(check CheckFunction, period time.Duration) {
	scheduledtask.NewScheduledTask(func() {
		ce.results <- check()
	}, period, 0)
}

func (ce CheckEngine) AddMultiCheck(check MultiCheckFunction, period time.Duration) {
	scheduledtask.NewScheduledTask(func() {
		for _, result := range check() {
			ce.results <- result
		}
	}, period, 0)
}
