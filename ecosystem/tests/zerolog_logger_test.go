package grpc_job

import (
	"github.com/OddEer0/ayaka/ecosystem"
	"github.com/rs/zerolog"
	"testing"
)

func TestAppLoggerFromZerolog(t *testing.T) {
	output := &testOut{}
	logger := zerolog.New(output)

	log := ecosystem.NewAppLoggerWithZerolog(&logger)

	testAppLogger(t, log, output, "message")
}
