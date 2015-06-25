package gochecks

import (
	"fmt"
	"strings"

	"encoding/json"
	"io/ioutil"
	"net/http"
)

type JobStatus struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

type JobsMessage struct {
	Jobs []JobStatus `json:"jobs"`
}

func NewJenkinsJobsChecker(host, service, jenkinsBaseUrl string) MultiCheckFunction {

	return func() []Event {

		results := []Event{}

		response, err := http.Get(jenkinsBaseUrl + "api/json?tree=jobs[name,color]")
		if err != nil {
			return []Event{Event{Host: host, Service: service, State: "critical", Description: err.Error()}}
		}
		if response.StatusCode != 200 {
			return []Event{Event{Host: host, Service: service, State: "critical", Description: fmt.Sprintf("Response %d", response.StatusCode)}}
		}
		if response.Body == nil {
			return []Event{Event{Host: host, Service: service, State: "critical", Description: fmt.Sprintf("Empty body")}}
		}

		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return []Event{Event{Host: host, Service: service, State: "critical", Description: fmt.Sprintf("Error geting body")}}
		}
		var jobs JobsMessage
		err = json.Unmarshal(body, &jobs)
		if err == nil {
			for _, job := range jobs.Jobs {
				state := "critical"
				if strings.HasPrefix(job.Color, "blue") {
					state = "ok"
				}
				results = append(results, Event{
					Host:    host,
					Service: service + " " + job.Name,
					State:   state})
			}
			return results
		} else {
			return []Event{Event{Host: host, Service: service, State: "critical", Description: err.Error()}}
		}
		return []Event{}
	}

}
