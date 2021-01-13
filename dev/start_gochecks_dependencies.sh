#!/bin/bash

docker run --name mysql_gochecks -p 3306:3306 -e MYSQL_ROOT_PASSWORD=rootpass -d mysql:5.5
docker run --name postgres_gochecks -p 5432:5432 -e POSTGRES_PASSWORD=mysecretpassword -d postgres:9.5
docker run --name rabbitmq_gochecks -p 5672:5672 -d rabbitmq