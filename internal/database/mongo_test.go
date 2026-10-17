package database

import (
	"context"
	"strings"
	"testing"
)

func TestOpenRejectsInvalidConfigurationWithoutSecrets(t *testing.T) {
	client, err := Open(context.Background(), "invalid://private-password")
	if client != nil || err == nil {
		t.Fatal("expected configuration error")
	}
	if strings.Contains(err.Error(), "private-password") {
		t.Fatal("error exposes secret")
	}
}

func TestOpenHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	client, err := Open(ctx, "mongodb://127.0.0.1:1")
	if client != nil || err == nil {
		t.Fatal("expected canceled connection to fail")
	}
}
