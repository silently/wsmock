package integration_test

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

var durationUnit = 10 * time.Millisecond

func getTestOutput(t *testing.T) string {
	fv := reflect.ValueOf(t).Elem().FieldByName("output")
	return strings.TrimSpace(string(fv.Bytes()))
}

type Message struct {
	Kind    string `json:"kind"`
	Payload string `json:"payload"`
}
