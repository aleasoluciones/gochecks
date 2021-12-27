all: clean test build

update_dep:
	go get $(DEP)
	go mod tidy

update_all_deps:
	go get -u
	go mod tidy

format:
	go fmt ./...

test:
	go vet ./...
	go clean -testcache
	go test -tags integration ./... -timeout 60s

build:
	go build -o changeme example/example.go

clean:
	rm -f changeme

start_dependencies:
	docker-compose -f dev/gochecks_devdocker/docker-compose.yml up -d

stop_dependencies:
	docker-compose -f dev/gochecks_devdocker/docker-compose.yml stop

rm_dependencies:
	docker-compose -f dev/gochecks_devdocker/docker-compose.yml down -v


.PHONY: all update_dep update_all_deps format test build clean start_dependencies stop_dependencies rm_dependencies
