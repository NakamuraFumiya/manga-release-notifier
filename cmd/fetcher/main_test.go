package main

import (
	"context"
	"testing"
)

func TestHello(t *testing.T) {
	if err := hello(context.Background()); err != nil {
		t.Fatalf("hello returned error: %v", err)
	}
}
