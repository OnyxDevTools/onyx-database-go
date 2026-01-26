package main

import (
	"bytes"
	"os"
	"testing"
)

func TestRunMainUnknownCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	os.Args = []string{"onyx-go", "bogus"}
	exit = func(int) {}
	defer func() { exit = os.Exit }()
	if code := dispatch(os.Args[1:], &stdout, &stderr); code != 2 {
		t.Fatalf("expected unknown command code 2, got %d", code)
	}
}
