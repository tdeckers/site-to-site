package main

import (
	"fmt"
	"testing"
)

func TestGetStack(t *testing.T) {
	debug = true

	out, err := getStackByName("s2s")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("Stack status: %v\n", out.StackStatus)
}

// TODO: consider adding create/delete stack?
