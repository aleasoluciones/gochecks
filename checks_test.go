// +build integration

package gochecks_test

import (
	"fmt"
	"log"
	"os"
	"testing"

	"net/http"
	"net/http/httptest"

	"github.com/streadway/amqp"

	. "."

	"github.com/stretchr/testify/assert"
)

func TestHttpCheckerWithHttpServerUp(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	}))
	defer ts.Close()

	check := NewHTTPChecker("host", "service", ts.URL, 200)
	checkResult := check()

	assert.Equal(t, "ok", checkResult.State)
	assert.InDelta(t, checkResult.Metric, 0, 100)
}

func TestHttpCheckerWithServerDown(t *testing.T) {
	t.Parallel()

	check := NewHTTPChecker("host", "service", "https://unknownurl/", 200)
	checkResult := check()

	assert.Equal(t, "critical", checkResult.State)
	assert.InDelta(t, checkResult.Metric, 0, 100)
}

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

	conn, err := amqp.Dial(amqpUrl)
	if err != nil {
		log.Panic("Connection error RammbitMQ ", amqpUrl)
	}
	ch, _ := conn.Channel()
	defer conn.Close()
	defer ch.Close()

	ch.ExchangeDeclare(exchange, "topic", true, false, false, false, nil)
	ch.QueueDelete(queue, false, false, true)
	ch.QueueDeclare(queue, false, false, false, false, nil)
	ch.QueueBind(queue, "#", exchange, false, nil)

	check := NewRabbitMQQueueLenCheck("host", "service", amqpUrl, queue, 2)
	checkResult := check()
	assert.Equal(t, "ok", checkResult.State)
	assert.Equal(t, float32(0), checkResult.Metric)

	publishMessage(ch, exchange, routingKey, "msg1")
	publishMessage(ch, exchange, routingKey, "msg2")

	checkResult = check()
	assert.Equal(t, "ok", checkResult.State)
	assert.Equal(t, float32(2), checkResult.Metric)

	publishMessage(ch, exchange, routingKey, "msg3")
	checkResult = check()

	assert.Equal(t, "critical", checkResult.State)
	assert.Equal(t, float32(3), checkResult.Metric)
}

func TestRabbitMQQueueListLenCheck(t *testing.T) {
	t.Parallel()
	amqpUrl := amqpUrlFromEnv()
	queue1 := "q1"
	queue2 := "q2"
	queues := []string{queue1, queue2}
	exchange := "e1"
	routingKey1 := "r1"
	routingKey2 := "r2"

	conn, err := amqp.Dial(amqpUrl)
	if err != nil {
		log.Panic("Connection error RammbitMQ ", amqpUrl)
	}
	ch, _ := conn.Channel()
	defer conn.Close()
	defer ch.Close()

	ch.ExchangeDeclare(exchange, "topic", true, false, false, false, nil)
	ch.QueueDelete(queue1, false, false, true)
	ch.QueueDelete(queue2, false, false, true)
	ch.QueueDeclare(queue1, false, false, false, false, nil)
	ch.QueueDeclare(queue2, false, false, false, false, nil)
	ch.QueueBind(queue1, "r1", exchange, false, nil)
	ch.QueueBind(queue2, "r2", exchange, false, nil)

	check := NewRabbitMQQueueListLenCheck("host", "service", amqpUrl, queues, 6)
	checkResult := check()
	assert.Equal(t, "ok", checkResult.State)
	assert.Equal(t, float32(0), checkResult.Metric)

	publishMessage(ch, exchange, routingKey1, "msg1")
	publishMessage(ch, exchange, routingKey1, "msg1")
	publishMessage(ch, exchange, routingKey2, "msg2")
	publishMessage(ch, exchange, routingKey2, "msg2")

	checkResult = check()
	assert.Equal(t, "ok", checkResult.State)
	assert.Equal(t, float32(4), checkResult.Metric)

	publishMessage(ch, exchange, routingKey1, "msg3")
	publishMessage(ch, exchange, routingKey2, "msg6")
	publishMessage(ch, exchange, routingKey1, "msg7")
	checkResult = check()

	assert.Equal(t, "critical", checkResult.State)
	assert.Equal(t, float32(7), checkResult.Metric)
}

func TestRabbitMQQueueLenCheckReturnsCriticalWhenCantConnectToRabbitMQ(t *testing.T) {
	t.Parallel()

	check := NewRabbitMQQueueLenCheck("host", "service", amqpUrlFromEnv()+"whatever", "queue", 2)
	checkResult := check()

	assert.Equal(t, "critical", checkResult.State)
}

func TestMysqlConnectionErrorCheck(t *testing.T) {
	t.Parallel()

	check := NewMysqlConnectionCheck("host", "service", "mysql://nohost/nodb")
	checkResult := check()

	assert.Equal(t, "critical", checkResult.State)
}

func TestMysqlConnectionOkCheck(t *testing.T) {
	t.Parallel()
	check := NewMysqlConnectionCheck("host", "service", os.Getenv("MYSQL_URL"))
	checkResult := check()

	assert.Equal(t, "ok", checkResult.State)
	assert.InDelta(t, checkResult.Metric, 0, 100)
}

func TestPostgresConnectionErrorCheck(t *testing.T) {
	t.Parallel()
	postgres_url := "postgres://wronguser:wrongpass@localhost/postgres"
	check := NewPostgresConnectionCheck("host", "service", postgres_url)
	checkResult := check()

	assert.Equal(t, "critical", checkResult.State)
}

func TestPostgresOkCheck(t *testing.T) {
	t.Parallel()

	check := NewPostgresConnectionCheck("host", "service", os.Getenv("POSTGRES_URL"))
	checkResult := check()

	assert.Equal(t, "ok", checkResult.State)
	assert.InDelta(t, checkResult.Metric, 0, 100)
}
