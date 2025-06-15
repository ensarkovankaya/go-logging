package elastic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/elastic/go-elasticsearch/v8"

	"github.com/ensarkovankaya/go-logging/core"
)

type testCase struct {
	DocumentID  string       `json:"documentID"`
	Index       string       `json:"index"`
	Level       core.Level   `json:"level"`
	Message     string       `json:"message"`
	Extra       []core.Field `json:"extra,omitempty"`
	Fields      []core.Field `json:"fields,omitempty"`
	ShouldIndex bool         `json:"shouldIndex,omitempty"`
}

var (
	testAddress   = "http://localhost:9200"
	testIndex     = "test-index"
	testTimestamp = time.Date(2023, 10, 1, 12, 0, 0, 0, time.UTC)
	testNowFunc   = func() time.Time {
		return testTimestamp
	}
)

type mockRoundTrip struct {
	T             *testing.T
	lock          sync.Locker
	IndexRequests []map[string]any
}

func newMockTransport(t *testing.T) *mockRoundTrip {
	return &mockRoundTrip{
		T:    t,
		lock: &sync.Mutex{},
	}
}

func (t *mockRoundTrip) Finish() {
	t.lock.Lock()
	defer t.lock.Unlock()
}

func (t *mockRoundTrip) RoundTrip(req *http.Request) (*http.Response, error) {
	t.lock.Lock()
	defer t.lock.Unlock()

	if req.Body == nil {
		t.T.Error("Expected request body, got nil")
	}

	if req.Method != http.MethodPost {
		t.T.Errorf("Expected POST request, got %s", req.Method)
	}
	expectedURL := testAddress + "/_bulk"
	if req.URL.String() != expectedURL {
		t.T.Errorf("Expected request URL %s, got %s", expectedURL, req.URL.String())
	}
	if req.Body == nil {
		t.T.Error("Expected request body, got nil")
		return nil, fmt.Errorf("request body is nil")
	}

	body, err := req.GetBody()
	if err != nil {
		t.T.Errorf("Failed to get request body: %v", err)
		return nil, err
	}

	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		t.T.Errorf("Failed to read request body: %v", err)
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(bodyBytes)), "\n")
	if len(lines) < 2 {
		t.T.Error("Expected at least two line in request body, got none")
		return nil, fmt.Errorf("invalid request body, expected at least two lines")
	}

	if lines[0] != fmt.Sprintf("{\"index\":{\"_index\":\"%v\"}}", testIndex) {
		t.T.Errorf("Expected first line of request body to be empty, got %s", lines[0])
		return nil, fmt.Errorf("first line of request body is not empty")
	}

	for _, line := range lines[1:] {
		var doc map[string]any
		if err = json.Unmarshal([]byte(line), &doc); err != nil {
			t.T.Errorf("Failed to unmarshal line:\n%s\nError: %v", line, err)
			return nil, fmt.Errorf("failed to unmarshal line: %w", err)
		}
		t.IndexRequests = append(t.IndexRequests, doc)
	}

	return buildIndexResponse(t.T, len(lines)-1), nil
}

func getTestClient(t *testing.T, transport http.RoundTripper, options ...ClientOption) *elasticsearch.Client {
	options = append([]ClientOption{func(cfg *elasticsearch.Config) {
		cfg.Addresses = []string{testAddress}
		cfg.Transport = transport
	}}, options...)
	client, err := NewClient(options...)
	if err != nil {
		t.Fatalf("Failed to initialize Elasticsearch client: %v", err)
	}
	return client
}

func buildIndexResponse(t *testing.T, logCount int) *http.Response {
	items := make([]map[string]any, logCount)
	for i := 0; i < logCount; i++ {
		items[i] = map[string]any{
			"index": map[string]any{
				"_index":   testIndex,
				"_id":      fmt.Sprintf("id-%d", i),
				"_version": 1,
				"result":   "created",
				"_shards": map[string]any{
					"total":      2,
					"successful": 1,
					"failed":     0,
				},
				"_seq_no":       i + 1,
				"_primary_term": 1,
				"status":        http.StatusCreated,
			},
		}
	}
	payload := map[string]any{
		"errors": false,
		"took":   526,
		"items":  items,
	}
	return buildHTTPResponse(t, http.StatusOK, payload)
}

func buildHTTPResponse(t *testing.T, statusCode int, payload map[string]any) *http.Response {
	var body io.ReadCloser
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			t.Fatal("Failed to marshal response payload: " + err.Error())
		}
		body = io.NopCloser(bytes.NewReader(data))
	}
	return &http.Response{
		Proto:      "HTTP/1.1",
		Status:     fmt.Sprintf("%d %s", statusCode, http.StatusText(statusCode)),
		StatusCode: statusCode,
		Body:       body, // In a real test, you would use a proper io.Reader
		Header: http.Header{
			"Content-Type":      []string{"application/json"},
			"X-Elastic-Product": []string{"Elasticsearch"},
		},
	}
}
