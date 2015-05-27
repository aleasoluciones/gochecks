package main

import (
	"fmt"
	"log"
	"net/url"
	"time"

	_ "github.com/aleasoluciones/felixcheck"
	"github.com/influxdb/influxdb/client"
)

type InfluxdbQuerier struct {
	conn     *client.Client
	query    string
	database string
}

func NewInfluxdbClient(host string, port int, username, password string) *client.Client {
	u, err := url.Parse(fmt.Sprintf("http://%s:%d", host, port))
	if err != nil {
		log.Fatal(err)
	}
	conf := client.Config{
		URL:      *u,
		Username: username,
		Password: password,
	}
	client, err := client.NewClient(conf)
	if err != nil {
		log.Fatal(err)
	}
	return client
}

func NewInfluxdbQuerier(host string, port int, database, username, password, query string) InfluxdbQuerier {

	return InfluxdbQuerier{conn: NewInfluxdbClient(host, port, username, password), query: query, database: database}

}

func (i InfluxdbQuerier) MetricFunc() float32 {
	res, err := i.queryDB()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("EGI >>>>>>", len(res))
	value := res[0].Series[0].Values[0][1]
	fmt.Println("EGI >>>>>>>>>>>>", value)
	result, _ := value.(float32)
	return result

}

func (i InfluxdbQuerier) queryDB() (res []client.Result, err error) {
	fmt.Println("EGI query >>>>", i.database)
	q := client.Query{
		Command:  i.query,
		Database: i.database,
	}
	response, err := i.conn.Query(q)
	if err != nil {
		fmt.Println("EGI error", err)
	}
	if response.Error() != nil {
		fmt.Println("EGI >>>>>>", "no hay errores....")
	}

	return

}

func main() {
	query := "select * from configurations"
	influxdbquerier := NewInfluxdbQuerier("oldmanpeabody-pepsifree-1.c.influxdb.com", 8086, "configurations", "root", "ecb0c71aa60c2830", query)

	for {
		value := influxdbquerier.MetricFunc()
		fmt.Println("EGI >>>>>>", value)
		time.Sleep(2 * time.Second)
	}
}
