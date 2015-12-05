package gochecks

import (
	"encoding/json"
	"github.com/aleasoluciones/simpleamqp"
)

func amqp2CheckEngine(checkEngine *CheckEngine, brokerURI, exchange, topic, queue string, queueOptions *simpleamqp.QueueOptions) {

	if queueOptions == nil {
		queueOptions = &simpleamqp.QueueOptions{Durable: false, Delete: true, Exclusive: true}
	}
	amqpConsumer := simpleamqp.NewAmqpConsumer(brokerURI)
	amqpEvents := amqpConsumer.ReceiveWithoutTimeout(exchange, []string{topic}, queue, *queueOptions)

	for ev := range amqpEvents {
		event := Event{}
		json.Unmarshal([]byte(ev.Body), &event)
		checkEngine.AddResult(event)
	}
}
