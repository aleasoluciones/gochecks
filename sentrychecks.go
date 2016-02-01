package gochecks

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func NewSentryUnresolvedIssuesChecker(host, service, sentryBaseUrl, projectName string) CheckFunction {
	return func() Event {
		response, err := http.Get(sentryBaseUrl + "/api/0/projects/" + projectName + "/issues/?query=is:unresolved&statsPeriod=24h")
		if err != nil {
			return Event{Host: host, Service: service, State: "critical", Description: err.Error()}
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusOK {
			return Event{Host: host, Service: service, State: "critical", Description: fmt.Sprintf("Response %d", response.StatusCode)}
		}
		decoder := json.NewDecoder(response.Body)
		var unresolvedIssues []interface{}
		err = decoder.Decode(&unresolvedIssues)
		if err != nil {
			return Event{Host: host, Service: service, State: "critical", Description: err.Error()}
		}
		state := "ok"
		if len(unresolvedIssues) != 0 {
			state = "critical"
		}
		return Event{Host: host, Service: service, State: state, Metric: len(unresolvedIssues)}
	}
}
