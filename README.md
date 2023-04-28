# wsmock

Library for testing websocket connection handlers:

- the server handler (subject of the test) is provided a mocked connection to read from and write to (in place of a real `Conn`)
- tests are scripted by sending messages (client to server) like you would in JS with `ws.send(...)`
- possibility to have several mocked connections interacting (through the server handler/s) in the same test
- for each mocked conn a `Recorder` struct is used to set assertions (`AssertReceived`, `AssertNotReceived`, `AssertReceivedSparseSequence`... or custom assertions)
- assertions are capped with a timeout, but they won't necessarilly wait till it's reached:
  - for instance `AssertReceived` succeeds as soon as the right message comes in
  - at the opposite `AssertNotReceived` needs the timeout to be reached to check the message has not been received.

All the assertions are from the client point of view, they mirror what you would expect to receive in JS with `ws.onmessage(...)`.

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
    conn1, rec1 := wsmock.NewGorillaMockAndRecorder(t)
    conn2, rec2 := wsmock.NewGorillaMockAndRecorder(t)
    // runPeer function is the target of this test (supposedly implemented elsewhere in mypackage)
    runPeer(conn1) 
    runPeer(conn2)

    // sequence: Johnny connects too late to receive Micheline's greeting
    conn1.Send(Message{"join", "Micheline"})
    conn1.Send(Message{"send", "Bonjour"})
    conn2.Send(Message{"join", "Johnny"})
    conn2.Send(Message{"send", "Hello"})

    rec1.AssertReceived(Message{"incoming", "Hello"})
    rec2.AssertNotReceived(Message{"incoming", "Bonjour"}) // or implement chat history in serve!
    wsmock.RunAssertions(t, 100 * time.Millisecond)
  })
}
```

## Installation

```
go get github.com/silently/wsmock
```

## Usage

When using Gorilla's websocket, you may have a code similar to: 

```
func serveWs(w http.ResponseWriter, r *http.Request) {
    ws, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println(err)
    }
    runPeer(ws)
}
```

The main gotcha here is to be able to pass runPeer our mock connection.

If `runPeer` signature is:

```
func runPeer(conn *gorilla.Conn) {}
```

We will have to change it by relying on an interface that satisfies gorilla.Conn this way:

```
type wsConn interface {
	ReadJSON(any) error
	WriteJSON(any) error
	Close() error
  // add more methods if you use them
}

func runPeer(conn *wsConn) {}
```

Now it's possible to target runPeer in our tests with a mocked connection:

```
conn, rec := wsmock.NewGorillaMockAndRecorder(t)
runPeer(conn) 
```

TODO: comparison with httptest

## Development

Generate coverage reports:

```
go test -v -coverprofile cover.out
go tool cover -html cover.out -o cover.html
open cover.html
```