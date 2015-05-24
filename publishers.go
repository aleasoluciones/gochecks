package felixcheck

import (
	"fmt"
	"log"
	"net/url"
	"time"

	"encoding/json"

	"github.com/aleasoluciones/simpleamqp"
	"github.com/bigdatadev/goryman"
	"github.com/influxdb/influxdb/client"
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

type InfluxdbPublisher struct {
	client       *client.Client
	databaseName string
}

func NewInfluxdbPublisher(host string, port int, databaseName, username, password string) InfluxdbPublisher {
	u, err := url.Parse(fmt.Sprintf("http://%s:%d", host, port))
	if err != nil {
		log.Fatal(err)
	}
	conf := client.Config{
		URL:      *u,
		Username: username,
		Password: password,
	}
	con, err := client.NewClient(conf)
	if err != nil {
		log.Fatal(err)
	}
	p := InfluxdbPublisher{client: con, databaseName: databaseName}
	return p
}

func (p InfluxdbPublisher) PublishCheckResult(event Event) {
	influxdbTags := make(map[string]string)
	for _, value := range event.Tags {
		influxdbTags[value] = "0"
	}
	point := client.Point{
		Name: "bifer",
		Tags: influxdbTags,
		Fields: map[string]interface{}{
			"value": 666,
		},
		Time:      time.Now(),
		Precision: "s",
	}
	bps := client.BatchPoints{
		Points:          make([]client.Point, 1),
		Database:        p.databaseName,
		RetentionPolicy: "default",
	}
	bps.Points[0] = point
	_, err := p.client.Write(bps)
	if err != nil {
		log.Printf("[error] sending check %s", event)
		return
	}
}
