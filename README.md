# wsmock

Library for testing websocket connection handlers:

- provides a mocked websocket connection that the server handler (the subject of the test) will read from and write to (in place of a real `Conn`)
- tests are scripted by sending messages (client to server) like you would in JS with `ws.send(...)`
- possibility to have several mocked connections interacting (through the server handler/s) in the same test
- for each mocked conn a `Recorder` struct is provided to define the outcome of the test with assertions (`AssertReceived`, `AssertNotReceived`, `AssertReceivedSparseSequence`...)
- possibility to define custom assertions
- assertions are capped with a timeout, but they won't necessarilly wait till it's reached:
  - for instance `AssertReceived` succeeds as soon as the right message comes in
  - at the opposite `AssertNotReceived` needs the timeout to be reached to check the message has not been received.

All the assertions have the client-side point of view, they mirror what you would expect in JS with `ws.onmessage(...)`.

## Status

wsmock is in an early stage of development, API may change.

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