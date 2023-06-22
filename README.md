# wsmock

Golang library for websocket testing.

A typical websocket handler based on [gorilla](https://github.com/gorilla/websocket) looks like:

```golang
func serveWs(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println(err)
    }
    runPeer(conn) // runPeer is the target of a wsmock test
}
```

Package wsmock provides utilities to:

- mock `conn` with an API to send messages to the server handler, like you would in JS with `ws.send(...)`
- set assertions defining the expected server handler responses, mirroring what should be received in JS with `ws.onmessage(...)`)
- give assertions a timeout (that will only be waited if assertions don't succeed or fail sooner)
- have several mocked connections interacting (through the server handler/s) in the same test

## Assertions

A word on httptest and recorders
https://pkg.go.dev/net/http/httptest#ResponseRecorder

through a `Recorder` struct (`AssertReceived`, `AssertNotReceived`, `AssertReceivedSparseSequence`... or custom assertions)
- assertions are capped with a timeout, but they won't necessarilly wait till it's reached:
  - for instance `AssertReceived` succeeds as soon as the right message comes in
  - at the opposite `AssertNotReceived` needs the timeout to be reached to check the message has not been received.



## Status

wsmock is in an early stage of development, API may change.

Currently only Gorilla websocket mocks are provided, more could be added.

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
  // runPeer is the target of this test, supposedly implemented elsewhere in mypackage
  t.Run("two peers can connect and exchange hellos", func(t *testing.T) {
    conn1, rec1 := wsmock.NewGorillaMockAndRecorder(t)
    conn2, rec2 := wsmock.NewGorillaMockAndRecorder(t)
    runPeer(conn1) 
    runPeer(conn2)

    // sequence: Johnny connects too late to receive Micheline's greeting
    conn1.Send(Message{"join", "Micheline"})
    conn1.Send(Message{"send", "Bonjour"})
    conn2.Send(Message{"join", "Johnny"})
    conn2.Send(Message{"send", "Hello"})

    rec1.AssertReceived(Message{"incoming", "Hello"})
    rec2.AssertNotReceived(Message{"incoming", "Bonjour"})
    // ... or AssertReceived if chat history is implemented
    wsmock.RunAssertions(t, 100 * time.Millisecond)
  })
}
```

## Installation

```sh
go get github.com/silently/wsmock
```

## How to mock

When using Gorilla's websocket, you may have a code similar to: 

```golang
func serveWs(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println(err)
    }
    runPeer(conn)
}
```

The main gotcha here is to be able to pass runPeer our mock connection.

If `runPeer` signature is:

```golang
func runPeer(conn *gorilla.Conn) {}
```

We will have to change it by relying on an interface that satisfies gorilla.Conn this way:

```golang
type wsConn interface {
	ReadJSON(any) error
	WriteJSON(any) error
	Close() error
  // add more methods if you use them
}

func runPeer(conn *wsConn) {}
```

Now it's possible to target runPeer in our tests with a mocked connection:

```golang
conn, rec := wsmock.NewGorillaMockAndRecorder(t)
runPeer(conn) 
```

TODO: if no runPeer function
TODO: several RunAssertions, how they work

## Development

Generate coverage reports:

```sh
go test -v -coverprofile cover.out
go tool cover -html cover.out -o cover.html
open cover.html
```