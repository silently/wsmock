package recorder_test

import (
	"reflect"
	"strings"
	"testing"
)

func getTestOutput(t *testing.T) string {
	fv := reflect.ValueOf(t).Elem().FieldByName("output")
	return strings.TrimSpace(string(fv.Bytes()))
}

type Message struct {
	Kind    string `json:"kind"`
	Payload string `json:"payload"`
}
