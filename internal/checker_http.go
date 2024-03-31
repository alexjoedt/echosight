package echosight

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var _ Checker = (*HTTPChecker)(nil)
var _ Validator = (*HTTPChecker)(nil)

const (
	BearerAuth HTTPAuthType = "bearer-auth"
	BasicAuth  HTTPAuthType = "basic-auth"
)

type HTTPAuthType string

// HTTPChecker implements the checker interface
type HTTPChecker struct {
	Host               string            `json:"host"` // should be come from the Host type, since a dector is always binded to a host
	URL                string            `json:"url"`
	Headers            map[string]string `json:"headers"`
	AuthenticationType HTTPAuthType      `json:"authenticationType"`
	Credentials        Credentials       `json:"credentials"`
	ExpectedBody       string            `json:"expectedBody"`
	ExpectedStatus     int               `json:"expectedStatus"`
	ExpectedHeader     http.Header       `json:"expectedHeader"`
	detector           *Detector         `json:"-"`
}

func (h *HTTPChecker) Validate() bool {
	_, err := url.Parse(h.URL)
	return err == nil
}

func (h *HTTPChecker) Detector() *Detector {
	return h.detector
}

func (hc *HTTPChecker) ID() string {
	return hc.detector.ID.String()
}

func (hc *HTTPChecker) Interval() time.Duration {
	return time.Duration(hc.detector.Interval)
}

func (hc *HTTPChecker) Check(ctx context.Context) *Result {
	// this is just a imple http check, just for testing, not finished yet
	// TODO: improve check and add the expected cases

	client := http.Client{}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, hc.URL, nil)
	start := time.Now()
	res, err := client.Do(req)
	if err != nil {
		return &Result{
			State:   StateCritical,
			Message: err.Error(),
			err:     err,
		}
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return &Result{
			State:   StateCritical,
			Message: err.Error(),
			err:     err}
	}

	result := Result{
		State:   StateOK,
		Message: string(data)[:16],
		Metric: &Metric{
			Fields: map[string]any{
				"response_time": time.Since(start).Milliseconds(),
			},
			Time: time.Now(),
		},
	}

	if !strings.Contains(string(data), hc.ExpectedBody) {
		result.State = StateCritical
	}

	return hc.detector.ApplyIDs(&result)
}
