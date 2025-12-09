package karmada

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func assertEqual[T comparable](t *testing.T, name string, got, want T) {
	t.Helper()
	if got != want {
		t.Fatalf("%s = %v, want %v", name, got, want)
	}
}

func TestGetMemberClusters_Success(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// verify path
		if r.URL.Path != "/apis/cluster.karmada.io/v1alpha1/clusters" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		// verify Authorization header
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Fatalf("unexpected Authorization header: %q", auth)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
  "items": [
    {
      "metadata": { "name": "member-1" },
      "spec": { "apiEndpoint": "https://member1.example.com" }
    },
    {
      "metadata": { "name": "member-2" },
      "spec": { "apiEndpoint": "https://member2.example.com" }
    }
  ]
}`))
	}))
	defer ts.Close()

	c := &Client{
		api:    ts.URL,
		token:  "test-token",
		client: ts.Client(),
	}

	ctx := context.Background()
	clusters, err := c.GetMemberClusters(ctx)
	if err != nil {
		t.Fatalf("GetMemberClusters returned error: %v", err)
	}

	if len(clusters) != 2 {
		t.Fatalf("len(clusters) = %d, want 2", len(clusters))
	}

	assertEqual(t, "clusters[0].Name", clusters[0].Name, "member-1")
	assertEqual(t, "clusters[0].Endpoint", clusters[0].Endpoint, "https://member1.example.com")
	assertEqual(t, "clusters[1].Name", clusters[1].Name, "member-2")
	assertEqual(t, "clusters[1].Endpoint", clusters[1].Endpoint, "https://member2.example.com")
}

func TestGetMemberClusters_Non200Status(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "some error", http.StatusInternalServerError)
	}))
	defer ts.Close()

	c := &Client{
		api:    ts.URL,
		token:  "test-token",
		client: ts.Client(),
	}

	_, err := c.GetMemberClusters(context.Background())
	if err == nil {
		t.Fatalf("expected error for non-200 response, got nil")
	}
}
