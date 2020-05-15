package main

import (
  "testing"
  "errors"
)

type MockLoopInterface struct{}

func (l MockLoopInterface ) Run() error {
	return errors.New("1")
}

// Test Loop Failure
func TestLoopFailure( t *testing.T) {
	mockLoop := MockLoopInterface{}
	loop(mockLoop)
}

