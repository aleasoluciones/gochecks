package felixcheck

import (
	"sync"
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
	checkPublishers []CheckPublisher
	results         chan Event
	publishersMutex sync.Mutex
}

func NewCheckEngine(checkPublisher CheckPublisher) CheckEngine {
	checkEngine := CheckEngine{[]CheckPublisher{checkPublisher}, make(chan Event), sync.Mutex{}}
	go func() {
		for result := range checkEngine.results {
			checkEngine.publishersMutex.Lock()
			for _, publisher := range checkEngine.checkPublishers {
				go publisher.PublishCheckResult(result)
			}
			checkEngine.publishersMutex.Unlock()
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
