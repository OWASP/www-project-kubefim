package metrics

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestServerLifecycle(t *testing.T) {
	server, err := Listen("127.0.0.1:0", NewRegistry().Handler())
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- server.Serve() }()

	client := &http.Client{Timeout: time.Second}
	response, err := client.Get("http://" + server.Address().String() + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(response.Body)
	_ = response.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusOK || string(body) != "ok\n" {
		t.Fatalf("health response is %d %q", response.StatusCode, body)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		t.Fatal(err)
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}
