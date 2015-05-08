package felixcheck

import (
	"fmt"
	"log"

	"encoding/json"

	"github.com/aleasoluciones/simpleamqp"
	"github.com/bigdatadev/goryman"
)

type CheckPublisher interface {
	PublishCheckResult(goryman.Event)
}

type LogPublisher struct {
}

func NewLogPublisher() LogPublisher {
	return LogPublisher{}
}

func (p LogPublisher) PublishCheckResult(event goryman.Event) {
	log.Println(event)
}

type RabbitMqPublisher struct {
	publisher *simpleamqp.AmqpPublisher
}

func NewRabbitMqPublisher(amqpuri, exchange string) RabbitMqPublisher {
	p := RabbitMqPublisher{simpleamqp.NewAmqpPublisher(amqpuri, exchange)}
	return p
}

func (p RabbitMqPublisher) PublishCheckResult(event goryman.Event) {
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

func (p RiemannPublisher) PublishCheckResult(event goryman.Event) {
	err := p.client.Connect()
	if err != nil {
		log.Printf("[error] publishing check %s", event)
		return
	}
	defer p.client.Close()

	err = p.client.SendEvent(&event)
	if err != nil {
		log.Printf("[error] sending check %s", event)
		return
	}
}
