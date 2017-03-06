package gochecks

import (
	"fmt"
	"strings"
	"log"

	"encoding/json"

	"github.com/aleasoluciones/simpleamqp"
	"github.com/jsoriano/goryman"
)

const (
	riemannDescriptionMaxLength = 250
)

// CheckPublisher define the check result Publisher
type CheckPublisher interface {
	PublishCheckResult(Event)
}

// LogPublisher object to log each check result
type LogPublisher struct{}

// NewLogPublisher return a new LogPublisher
func NewLogPublisher() LogPublisher {
	return LogPublisher{}
}

// PublishCheckResult log the event
func (p LogPublisher) PublishCheckResult(event Event) {
	log.Println(event)
}

// ChannelPublisher object to publish to a channel each check result
type ChannelPublisher struct {
	Channel chan Event
}

// NewChannelPublisher return a new ChannelPublisher
func NewChannelPublisher(c chan Event) ChannelPublisher {
	return ChannelPublisher{c}
}

// PublishCheckResult send the event to the publisher channel
func (p ChannelPublisher) PublishCheckResult(event Event) {
	p.Channel <- event
}

// RabbitMqPublisher object to publish the event to a rabbitmq exchange
type RabbitMqPublisher struct {
	publisher *simpleamqp.AmqpPublisher
}

// NewRabbitMqPublisher return a publisher to send to RabbitMq exchange
func NewRabbitMqPublisher(amqpuri, exchange string) RabbitMqPublisher {
	p := RabbitMqPublisher{simpleamqp.NewAmqpPublisher(amqpuri, exchange)}
	return p
}

// PublishCheckResult publish the event to a configured rabbitmq exchange
func (p RabbitMqPublisher) PublishCheckResult(event Event) {
	service := strings.Replace(event.Service, " ", ".", -1)
	topic := fmt.Sprintf("check.%s.%s", event.Host, service)
	serialized, _ := json.Marshal(event)
	p.publisher.Publish(topic, serialized)
}

// RiemannPublisher object to publish events to a riemann server
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
	riemannEvent := goryman.Event{Description: p.normalizeDescriptionLength(event.Description),
		Host: event.Host,
		Service: event.Service,
		State: event.State,
		Metric: event.Metric,
		Tags: event.Tags,
		Attributes: event.Attributes,
		Ttl: event.TTL}

	err = p.client.SendEvent(&riemannEvent)
	if err != nil {
		log.Println("[error] sending check", event)
		return
	}
}

func (p RiemannPublisher) normalizeDescriptionLength(description string) string {
	if len(description) < riemannDescriptionMaxLength {
		return description
	}
	return description[0:riemannDescriptionMaxLength]
}
