package main

//we will test the run function

import "testing"

func TestRun(t *testing.T) {
	_, err := run()
	if err != nil {
		t.Error("Failed run")
	}
}
