package databricks

import (
	"context"
	"testing"
)

func TestNewClientBuildsWorkspaceClientFromHostAndToken(t *testing.T) {
	t.Parallel()

	client, err := NewClient(context.Background(), Config{
		Host:  "https://workspace.example.com",
		Token: "token",
	})
	if err != nil {
		t.Fatalf("unexpected client construction error: %v", err)
	}
	if client == nil {
		t.Fatal("expected client")
	}
	if client.Connections() == nil {
		t.Fatal("expected connection client")
	}
}

func TestNewClientRequiresHost(t *testing.T) {
	t.Parallel()

	_, err := NewClient(context.Background(), Config{
		Token: "token",
	})
	if err == nil {
		t.Fatal("expected missing host to fail")
	}
}

func TestNewClientRequiresToken(t *testing.T) {
	t.Parallel()

	_, err := NewClient(context.Background(), Config{
		Host: "https://workspace.example.com",
	})
	if err == nil {
		t.Fatal("expected missing token to fail")
	}
}
