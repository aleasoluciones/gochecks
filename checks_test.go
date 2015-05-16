package felixcheck_test

// +build integration

import (
	"os"

	"testing"

	"github.com/streadway/amqp"

	. "github.com/aleasoluciones/felixcheck"
	"github.com/stretchr/testify/assert"
)

func amqpUrlFromEnv() string {
	url := os.Getenv("AMQP_URL")
	if url == "" {
		url = "amqp://"
	}
	return url
}

func publishMessage(ch *amqp.Channel, exchange, routingKey, text string) {
	ch.Publish(
		exchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     "application/json",
			ContentEncoding: "",
			Body:            []byte(text),
			DeliveryMode:    amqp.Transient,
			Priority:        0,
		})

}

func TestRabbitMQQueueLenCheck(t *testing.T) {
	t.Parallel()
	amqpUrl := amqpUrlFromEnv()
	queue := "q"
	exchange := "e"
	routingKey := "r"

	conn, _ := amqp.Dial(amqpUrl)
	ch, _ := conn.Channel()
	defer conn.Close()
	defer ch.Close()

	ch.ExchangeDeclare(exchange, "topic", true, false, false, false, nil)
	ch.QueueDelete(queue, false, false, true)
	ch.QueueDeclare(queue, false, false, false, false, nil)
	ch.QueueBind(queue, "#", exchange, false, nil)

	check := NewRabbitMQQueueLenCheck("host", "service", amqpUrl, queue, 2)
	checkResult := check()
	assert.Equal(t, checkResult.State, "ok")
	assert.Equal(t, checkResult.Metric, float32(0))

	publishMessage(ch, exchange, routingKey, "msg1")
	publishMessage(ch, exchange, routingKey, "msg2")

	checkResult = check()
	assert.Equal(t, checkResult.State, "ok")
	assert.Equal(t, checkResult.Metric, float32(2))

	publishMessage(ch, exchange, routingKey, "msg3")
	checkResult = check()
	assert.Equal(t, checkResult.State, "critical")
	assert.Equal(t, checkResult.Metric, float32(3))

}

// func TestRabbitMQQueueLenCheckReturnsCriticalWhenCantConnectToRabbitMQ(t *testing.T) {
// 	t.Parallel()

// 	check := NewRabbitMQQueueLenCheck("host", "service", amqpUrlFromEnv()+"whatever", "queue", 2)
// 	checkResult := check()

// 	assert.Equal(t, checkResult.State, "critical")
// }
