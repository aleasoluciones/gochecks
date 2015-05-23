package felixcheck

import (
	"fmt"
	"log"

	"encoding/json"

	"github.com/aleasoluciones/simpleamqp"
	"github.com/bigdatadev/goryman"
)

type CheckPublisher interface {
	PublishCheckResult(Event)
}

type LogPublisher struct {
}

func NewLogPublisher() LogPublisher {
	return LogPublisher{}
}

func (p LogPublisher) PublishCheckResult(event Event) {
	log.Println(event)
}

type RabbitMqPublisher struct {
	publisher *simpleamqp.AmqpPublisher
}

func NewRabbitMqPublisher(amqpuri, exchange string) RabbitMqPublisher {
	p := RabbitMqPublisher{simpleamqp.NewAmqpPublisher(amqpuri, exchange)}
	return p
}

func (p RabbitMqPublisher) PublishCheckResult(event Event) {
	topic := fmt.Sprintf("check.%s.%s", event.Host, event.Service)
	serialized, _ := json.Marshal(event)
	p.publisher.Publish(topic, serialized)
}

type RiemannPublisher struct {
	client *goryman.GorymanClient
}

func NewRiemannPublisher(addr string) RiemannPublisher {
	p := RiemannPublisher{goryman.NewGorymanClient(addr)}
	return p
}

func (p RiemannPublisher) PublishCheckResult(event Event) {
	err := p.client.Connect()
	if err != nil {
		log.Printf("[error] publishing check %s", event)
		return
	}
	defer p.client.Close()
	riemannEvent := goryman.Event{Description: event.Description, Host: event.Host, Service: event.Service, State: event.State, Metric: event.Metric, Tags: event.Tags, Attributes: event.Attributes, Ttl: event.Ttl}

	err = p.client.SendEvent(&riemannEvent)
	if err != nil {
		log.Printf("[error] sending check %s", event)
		return
	}
}
