package gochecks

import (
	"github.com/aleasoluciones/simpleamqp"
)

func amqp2CheckEngine(checkEngine *CheckEngine, exchange, topic, queue string, queueOptions *simpleamqp.QueueOptions) {

	if !queueOptions {
		queueOptions := simpleamqp.QueueOptions{Durable: false, Delete: true, Exclusive: true}
	}
	amqpEvents := amqpConsumer.ReceiveWithoutTimeout(exchange, []string{topic}, queue, queueOptions)

	for ev := range amqpEvents {
		event := Event
		err = json.Unmarshal([]byte(ev.Body), &event)
		checkEngine.AddResult(event)
	}
}
