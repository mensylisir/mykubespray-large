package kubernetes

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Mock data for testing
const mockIndexYAML = `
entries:
  stable:
    - name: example-chart
      version: 1.0.0
      description: An example Helm chart
      urls:
        - https://example.com/charts/example-chart-1.0.0.tgz
      digest: sha256:abcd1234
      dependencies:
        - name: dependency-chart
          version: 1.0.0
`

// Create a mock server
func createMockServer() *httptest.Server {
	handler := http.NewServeMux()
	handler.HandleFunc("/index.yaml", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/index.yaml" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockIndexYAML))
	})
	return httptest.NewServer(handler)
}

// TestGetHelmCharts tests the GetHelmCharts function
func TestGetHelmCharts(t *testing.T) {
	server := createMockServer()
	defer server.Close()

	// Call GetHelmCharts with the mock server URL
	charts, err := GetHelmCharts("http://172.30.1.13:18091/repository/helm/index.yaml")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	t.Log(charts)
}
