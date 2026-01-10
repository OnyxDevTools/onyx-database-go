package main

import (
	"bytes"
	"os"
	"testing"
)

func TestRunGenInitInvalidArgs(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runGenInit([]string{"--unknown"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("expected usage exit, got %d", code)
	}
}

func TestRunGenMissingArgs(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runGen([]string{"--schema"}, &stdout, &stderr)
	if code == 0 {
		t.Fatalf("expected error on missing schema value")
	}
	// ensure usage printed
	if stdout.Len() == 0 && stderr.Len() == 0 {
		t.Fatalf("expected usage output")
	}
}

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
