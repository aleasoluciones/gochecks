package felixcheck

import (
	"time"

	"github.com/aleasoluciones/goaleasoluciones/scheduledtask"
	"github.com/bigdatadev/goryman"
)

type CheckEngine struct {
	checkPublisher CheckPublisher
	results        chan goryman.Event
}

func NewCheckEngine(checkPublisher CheckPublisher) CheckEngine {
	checkEngine := CheckEngine{checkPublisher, make(chan goryman.Event)}
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
