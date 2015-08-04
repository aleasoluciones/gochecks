package gochecks

import (
	"fmt"
	"log"

	"encoding/json"

	"github.com/aleasoluciones/simpleamqp"
	"github.com/bigdatadev/goryman"
)

// CheckPublisher define the check result Publisher
type CheckPublisher interface {
	PublishCheckResult(Event)
}

// LogPublisher object to log each check result
type LogPublisher struct {
}

// NewLogPublisher return a new LogPublisher
func NewLogPublisher() LogPublisher {
	return LogPublisher{}
}

func (p LogPublisher) PublishCheckResult(event Event) {
	log.Println(event)
}

type RabbitMqPublisher struct {
	publisher *simpleamqp.AmqpPublisher
}

// NewRabbitMqPublisher return a publisher to send to RabbitMq exchange
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

// NewRiemannPublisher return a publisher to send directly to a riemann instance
func NewRiemannPublisher(addr string) RiemannPublisher {
	p := RiemannPublisher{goryman.NewGorymanClient(addr)}
	return p
}

// PublishCheckResult return a publisher to send directly to a riemann instance
func (p RiemannPublisher) PublishCheckResult(event Event) {
	err := p.client.Connect()
	if err != nil {
		log.Println("[error] publishing check", event)
		return
	}
	defer p.client.Close()
	riemannEvent := goryman.Event{Description: event.Description, Host: event.Host, Service: event.Service, State: event.State, Metric: event.Metric, Tags: event.Tags, Attributes: event.Attributes, Ttl: event.TTL}

	err = p.client.SendEvent(&riemannEvent)
	if err != nil {
		log.Println("[error] sending check", event)
		return
	}
}
