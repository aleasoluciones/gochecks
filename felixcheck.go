package felixcheck

import (
	"fmt"
	"time"

	"encoding/json"

	"github.com/aleasoluciones/goaleasoluciones/scheduledtask"
)

type CheckResultMessage struct {
	Host    string `json:"host"`
	Service string `json:"service"`
	State   string `json:"state"`
}

type CheckResult struct {
	host    string
	service string
	result  bool
	err     error
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
		result, err := check()
		ce.results <- CheckResult{host, service, result, err}
	}, period, 0)
}

type CheckPublisher interface {
	PublishCheckResult(result CheckResult)
}

func (p RabbitMqPublisher) PublishCheckResult(result CheckResult) {
	var state string
	if result.result == true {
		state = "ok"
	} else {
		state = "critical"
	}
	topic := fmt.Sprintf("check.%s.%s", result.service, result.host)
	serialized, _ := json.Marshal(CheckResultMessage{result.host, result.service, state})
	p.publisher.Publish(topic, serialized)
}
