package gochecks

import (
	"fmt"
	"log"
	"strings"

	"encoding/json"

	"github.com/aleasoluciones/simpleamqp"
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
