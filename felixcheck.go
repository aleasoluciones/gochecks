package felixcheck

import (
	"time"

	"github.com/aleasoluciones/goaleasoluciones/scheduledtask"
)

type CheckResultMessage struct {
	Host    string  `json:"host"`
	Service string  `json:"service"`
	State   string  `json:"state"`
	Metric  float32 `json:"metric"`
}

type CheckResult struct {
	host    string
	service string
	result  bool
	err     error
	metric  float32
}

type CheckEngine struct {
	checkPublisher CheckPublisher
	results        chan CheckResult
}

func NewCheckEngine(checkPublisher CheckPublisher) CheckEngine {
	checkEngine := CheckEngine{checkPublisher, make(chan CheckResult)}
	go func() {
		for result := range checkEngine.results {
			checkEngine.checkPublisher.PublishCheckResult(result)
		}
	}()
	return checkEngine
}

func (ce CheckEngine) AddCheck(host, service string, period time.Duration, check CheckFunction) {
	scheduledtask.NewScheduledTask(func() {
		result, err, metric := check()
		ce.results <- CheckResult{host, service, result, err, metric}
	}, period, 0)
}
