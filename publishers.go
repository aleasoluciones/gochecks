package felixcheck

import (
	"github.com/aleasoluciones/simpleamqp"
	"github.com/bigdatadev/goryman"
)

type RabbitMqPublisher struct {
	publisher *simpleamqp.AmqpPublisher
}

func NewRabbitMqPublisher(amqpuri, exchange string) RabbitMqPublisher {
	p := RabbitMqPublisher{simpleamqp.NewAmqpPublisher(amqpuri, exchange)}
	return p
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
		panic(err)
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
		panic(err)
	}
}
