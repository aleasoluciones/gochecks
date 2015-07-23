package gochecks

import (
	"fmt"
	"regexp"
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

func NewJenkinsJobsChecker(host, service, jenkinsBaseURL string, jobRegExp string) CheckFunction {

	return func() Event {

		brokenJobs := []string{}

		response, err := http.Get(jenkinsBaseURL + "api/json?tree=jobs[name,color]")
		if err != nil {
			return Event{Host: host, Service: service, State: "critical", Description: err.Error()}
		}
		if response.StatusCode != 200 {
			return Event{Host: host, Service: service, State: "critical", Description: fmt.Sprintf("Response %d", response.StatusCode)}
		}
		if response.Body == nil {
			return Event{Host: host, Service: service, State: "critical", Description: fmt.Sprintf("Empty body")}
		}

		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return Event{Host: host, Service: service, State: "critical", Description: fmt.Sprintf("Error geting body")}
		}

		state := "ok"
		jobsOk := 0
		var jobs JobsMessage
		err = json.Unmarshal(body, &jobs)
		if err == nil {
			for _, job := range jobs.Jobs {
				matched, _ := regexp.MatchString(jobRegExp, job.Name)
				if matched {
					if !strings.HasPrefix(job.Color, "blue") {
						state = "critical"
						brokenJobs = append(brokenJobs, job.Name)
					} else {
						jobsOk = jobsOk + 1
					}
				}
			}
			return Event{Host: host, Service: service, State: state, Description: strings.Join(brokenJobs, ","), Metric: jobsOk}
		}
		return Event{Host: host, Service: service, State: "critical", Description: err.Error()}
	}

}
