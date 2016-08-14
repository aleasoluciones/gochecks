package gochecks

import (
	"time"

	"database/sql"

	_ "github.com/lib/pq"
)

// NewPostgresConnectionCheck returns a check function to detect connection/credentials problems to connect to postgres
//
func NewPostgresConnectionCheck(host, service, postgresuri string) CheckFunction {
	return func() Event {
		var t1 = time.Now()
		db, err := sql.Open("postgres", postgresuri)
		if err != nil {
			return Event{Host: host, Service: service, State: "critical", Description: err.Error()}
		}
		err = db.Ping()
		milliseconds := float32((time.Now().Sub(t1)).Nanoseconds() / 1e6)
		if err != nil {
			return Event{Host: host, Service: service, State: "critical", Description: err.Error(), Metric: milliseconds}
		}
		return Event{Host: host, Service: service, State: "ok", Metric: milliseconds}
	}
}
