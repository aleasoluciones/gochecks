package felixcheck

import (
	"fmt"
	"log"

	"encoding/json"

	"github.com/aleasoluciones/simpleamqp"
	"github.com/bigdatadev/goryman"
)

type CheckPublisher interface {
	PublishCheckResult(result CheckResult)
}

type RabbitMqPublisher struct {
	publisher *simpleamqp.AmqpPublisher
}

func NewRabbitMqPublisher(amqpuri, exchange string) RabbitMqPublisher {
	p := RabbitMqPublisher{simpleamqp.NewAmqpPublisher(amqpuri, exchange)}
	return p
}

func (p RabbitMqPublisher) PublishCheckResult(result CheckResult) {
	var state string
	if result.result == true {
		state = "ok"
	} else {
		state = "critical"
	}
	topic := fmt.Sprintf("check.%s.%s", result.service, result.host)
	serialized, _ := json.Marshal(CheckResultMessage{result.host, result.service, state, result.metric})
	p.publisher.Publish(topic, serialized)
}

type RiemannPublisher struct {
	client *goryman.GorymanClient
}

func NewRiemannPublisher(addr string) RiemannPublisher {
	p := RiemannPublisher{goryman.NewGorymanClient(addr)}
	return p
}

func (p RiemannPublisher) PublishCheckResult(result CheckResult) {
	err := p.client.Connect()
	if err != nil {
		log.Printf("[error] publishing check %s %s", result, err)
		return
	}
	defer p.client.Close()

	var state string
	if result.result == true {
		state = "ok"
	} else {
		state = "critical"
	}
	err = p.client.SendEvent(&goryman.Event{
		Host:    result.host,
		Service: result.host,
		State:   state,
		Metric:  result.metric,
	})
	if err != nil {
		log.Printf("[error] sending check %s %s", result, err)
		return
	}
}
