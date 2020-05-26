package main

import (
	"errors"
	"testing"

	log "github.com/sirupsen/logrus"
)

func init() { log.SetLevel(log.ErrorLevel) }

type MockLoopInterface struct{}

func (l MockLoopInterface) Run() error {
	return errors.New("1")
}

// Test Loop Failure
func TestLoopFailure(t *testing.T) {
	mockLoop := MockLoopInterface{}
	loop(mockLoop)
}
