# wsmock

Golang library for websocket testing.

A typical websocket handler based on [gorilla](https://github.com/gorilla/websocket) looks like:

```golang

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

func serveWs(w http.ResponseWriter, r *http.Request) { // HTTP handler
    conn, err := upgrader.Upgrade(w, r, nil) // creates a Gorilla websocket.Conn
    if err != nil {
        log.Println(err)
    }
    runPeer(conn) // WebSocket handler, the target of a wsmock test
}
```

The package wsmock provides utilities to:

- mock `websocket.Conn`, and give the mock an extra `Send` method to send messages to `runPeer` (or whatever implementation interacts with the connection), like you would in JS with `ws.send(...)`
- set assertions defining the server handler expected behaviour through its responses, mirroring what would happen in JS with `ws.onmessage(...)`
- give assertions a timeout (that will only be waited if their outcome can't be decided sooner)
- have several mocked connections interacting (through the server handler/s) in the same test

## Status

wsmock is in an early stage of development, API may change.

Currently, only Gorilla websocket mocks are provided (more could be added) and they only provide the following implementations: `ReadJSON`, `WriteJSON`, `Close` (more to be added).

## Installation

```sh
go get github.com/silently/wsmock
```

## Prerequesite

If we go on with our `runPeer` websocket handler, the main gotcha to make it usable in our tests is being able to pass it a mocked `conn` argument, meaning one that is not of Gorilla `websocket.Conn` type.

That's why if runPeer has this signature:

```golang
func runPeer(conn *websocket.Conn) {}
```

...we need to update it with an interface implemented both by Gorilla `websocket.Conn` and `wsmock.GorillaConn`.

*(This approach is similar to [httptest](https://pkg.go.dev/net/http/httptest#example-ResponseRecorder) relying on `ResponseRecorder`, "an implementation of `http.ResponseWriter`; that records its mutations for later inspection in tests")*

Depending on what methods are used by `runPeer` we could go with as little as:

```golang
type IConn interface {
	ReadJSON(any) error
	WriteJSON(any) error
	Close() error
  // add more methods if needed by runPeer implementation
}

func runPeer(conn *IConn) {}
```

Now `runPeer` can receive Gorilla `websocket.Conn` in real usage and `wsmock.GorillaConn` when testing.

Alternatively, instead of defining your own `IConn`, you can rely on `wsmock.IGorilla` interface: it declares all methods available on Gorilla [websocket.Conn](https://pkg.go.dev/github.com/gorilla/websocket#Conn):

```golang
import (
	ws "github.com/silently/wsmock"
)

func runPeer(conn *ws.IGorilla) {}
```

## Usage and assertions

Once `runPeer` accepts `wsmock.GorillaConn` as an argument, a test may look like:

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

    // script events (Johnny connects too late to receive Micheline's greeting)
    conn1.Send(Message{"join", "Micheline"})
    conn1.Send(Message{"send", "Bonjour"})
    conn2.Send(Message{"join", "Johnny"})
    conn2.Send(Message{"send", "Hello"})

    rec1.AssertReceived(Message{"incoming", "Hello"})
    // the next assertion is "not received" (supposing chat history is not implemented)
    rec2.AssertNotReceived(Message{"incoming", "Bonjour"})
    // run all previously declared assertions with a timeout
    wsmock.Run(t, 100 * time.Millisecond)
  })
}
```

We can see that `wsmock.NewGorillaConnAndRecorder` returns two structs:

- `wsmock.GorillaConn` used as the mocked websocket connection given to `runPeer`
- `wsmock.Recorder` that records server responses and is used to define assertions

The only methods you're supposed to use on `wsmock.GorillaConn` in the tests are:

- `Send` to script sent messages
- `Close` if you want to explicitely close connections "client-side" within a test (alternatively, wsmock will close them when test ends)

Here are the assertions provided by `wsmock.Recorder`:

- `AssertReceived`
- `AssertReceivedSparseSequence`
- `AssertReceivedAdjacentSequence`
- `AssertReceivedExactSequence`
- `AssertNotReceived`
- `AssertClosed`

And generic assertions that needs a finder:

- `AssertOnWrite`
- `AssertOnTimeoutOrClose`

Please note `*Received` assertions rely on the equality operator `==` (spec [here](https://go.dev/ref/spec#Comparison_operators)).

## For wsmock developers

Run wsmock own tests with:
```sh
CGO_ENABLED=0 go test
```

Generate coverage reports:

```sh
CGO_ENABLED=0 go test -v -coverprofile cover.out
go tool cover -html cover.out -o cover.html
open cover.html
```