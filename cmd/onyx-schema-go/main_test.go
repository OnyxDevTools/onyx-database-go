package main

import (
	"os"
	"testing"
)

func TestMainExitCodes(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want int
	}{
		{name: "help", args: []string{"--help"}, want: 0},
		{name: "unknown", args: []string{"nope"}, want: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			origExit := exit
			origArgs := os.Args
			t.Cleanup(func() {
				exit = origExit
				os.Args = origArgs
			})

			var got int
			exit = func(code int) {
				got = code
			}
			os.Args = append([]string{"onyx-schema-go"}, tt.args...)

			main()
			if got != tt.want {
				t.Fatalf("expected exit %d, got %d", tt.want, got)
			}
		})
	}
}
