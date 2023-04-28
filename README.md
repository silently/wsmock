# wsmock

Library for testing websocket connection handlers:

- provides a mocked websocket connection that the server handler (the subject of the test) will read from and write to (in place of a real `Conn`)
- tests are scripted by sending messages (client to server) like you would in JS with `ws.send(...)`
- possibility to have several mocked connections interacting (through the server handler/s) in the same test
- for each mocked conn a `Recorder` struct is provided to define the outcome of the test with assertions (`AssertReceived`, `AssertNotReceived`, `AssertReceivedSparseSequence`...)
- possibility to define custom assertions
- assertions are batched-run with a timeout, they won't wait till the timeout if they succeed before it. But some assertions (like `AssertNotReceived`) have to wait till the timeout (that you define).

## Status

The project is in an early stage of development, API may change.

Currently only Gorilla websocket mocks are provided, more to come.

## Example

```golang
package mypackage

import (
  "testing"
  "time"

  "github.com/silently/wsmock"
)

type Message struct {
	Kind    string `json:"kind"`
	Payload any    `json:"payload"`
}

func TestWs(t *testing.T) {
  t.Run("supervisor has runner", func(t *testing.T) {
    conn, rec := wsmock.NewGorillaMockWithRecorder(t)
    
    serve(conn) // target of the test (defined in mypackage)
    conn.Send(Message{"join", "room:1"})
    rec.AssertReceived(Message{"joined", "room:1"})
    rec.AssertReceived(Message{"users", []string{"Micheline", "Johnny"}})
    conn.Send(Message{"quit", ""})
    rec.AssertClosed()
    rec.RunAssertions(100 * time.Millisecond)
  })
}
```

## Development

Generate coverage reports:

```
go test -v -coverprofile cover.out
go tool cover -html cover.out -o cover.html
open cover.html
```