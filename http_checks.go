package gochecks

import (
	"fmt"
	"time"

	"io/ioutil"
	"net/http"
)

// ValidateHTTPResponseFunction function type that should validate a http response and return the state (ok, critical, warning) and error description for a check. (Used with NewGenericHTTPChecker)
type ValidateHTTPResponseFunction func(resp *http.Response) (state, description string)

// BodyGreaterThan return a function that given a http response return true if the body is greater than a given number of bytes
func BodyGreaterThan(minLength int) ValidateHTTPResponseFunction {
	return func(httpResp *http.Response) (state, description string) {
		if httpResp.StatusCode != 200 {
			return "critical", fmt.Sprintf("Response %d", httpResp.StatusCode)
		}
		if httpResp.Body == nil {
			return "critical", fmt.Sprintf("Empty body")
		}
		body, err := ioutil.ReadAll(httpResp.Body)
		if err != nil {
			return "critical", fmt.Sprintf("Error geting body")
		}
		if len(body) < minLength {
			return "critical", fmt.Sprintf("Obtained %d bytes, expected more than %d", len(body), minLength)
		}
		return "ok", ""
	}
}

// NewGenericHTTPChecker returns a check function that can check the returned http response of a http get with a given validation function
func NewGenericHTTPChecker(host, service, url string, validationFunc ValidateHTTPResponseFunction) CheckFunction {
	return func() Event {
		var t1 = time.Now()

		response, err := http.Get(url)
		milliseconds := float32((time.Now().Sub(t1)).Nanoseconds() / 1e6)
		result := Event{Host: host, Service: service, State: "critical", Metric: milliseconds}
		if err != nil {
			result.Description = err.Error()
		} else {
			if response.Body != nil {
				defer response.Body.Close()
			}
			result.State, result.Description = validationFunc(response)
		}
		return result
	}
}

// NewHTTPChecker returns a check function that get a given url and validate if the return code is the expected one
func NewHTTPChecker(host, service, url string, expectedStatusCode int) CheckFunction {
	return NewGenericHTTPChecker(host, service, url,
		func(httpResp *http.Response) (string, string) {
			if httpResp.StatusCode == expectedStatusCode {
				return "ok", ""
			}
			return "critical", fmt.Sprintf("Response %d", httpResp.StatusCode)
		})
}
