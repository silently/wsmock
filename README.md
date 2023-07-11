# wsmock

Golang library to help with WebSocket testing, writing tests like:

```golang
conn1, rec1 := wsmock.NewGorillaMockAndRecorder(t)  // get mocked Conn and server writes recorder...
conn2, rec2 := wsmock.NewGorillaMockAndRecorder(t)  // ...similarly to httptest NewRequest and NewRecorder
runWs(conn1)                                        // WebSocket handler -> target of the test
runWs(conn2)
conn1.Send("rock")                                  // client-side scripting
conn2.Send("paper")
rec1.AssertNotReceived("won")                       // declare assertions
rec2.AssertReceived("won")
wsmock.Run(t, 100*time.Millisecond)                 // run assertions on a *testing.T, with a timeout
```

Where `runWs` is a WebSocket handler based on [Gorilla WebSocket](https://github.com/gorilla/websocket), typically called like:

```golang

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

func serveWs(w http.ResponseWriter, r *http.Request) {  // HTTP handler
    conn, err := upgrader.Upgrade(w, r, nil)            // creates a Gorilla websocket.Conn
    if err != nil {
        log.Println(err)
    }
    runWs(conn)                                         // WebSocket handler -> target of the test
}
```

## Status

wsmock is in an early stage of development, API may change.

Currently, only Gorilla WebSocket mocks are provided (more WebSocket implementation mocks could be added) with a focus on reading from and writing to the Conn:

- that's why we provide mock implementations for the methods: `Close`, `ReadJSON`, `ReadMessage`, `NextReader`, `NextWriter`, `WriteJSON`, `WriteMessage`
- but other methods (like  `CloseHandler`, `EnableWriteCompression`...) from Gorilla `websocket.Conn` are blank/noop

*(wsmock itself has a good test coverage, but does not reach 100% because of these blank/noop implementations: they will only be tested when a proper implementation is considered)*

## Installation

```sh
go get github.com/silently/wsmock
```

## Prerequesite

Going on with our `runWs` WebSocket handler, the main gotcha to test it is: being able to give it a mocked `conn` argument, meaning one that is not of Gorilla `websocket.Conn` type.

If `runWs` has this signature:

```golang
func runWs(conn *websocket.Conn) {}
```

...we need to update it with an interface implemented both by Gorilla `websocket.Conn` and `wsmock.GorillaConn`.

*(This approach is similar to [httptest](https://pkg.go.dev/net/http/httptest#example-ResponseRecorder) that relies on `ResponseRecorder`, "an implementation of `http.ResponseWriter`; that records its mutations for later inspection in tests")*

Depending on what methods are used by `runWs` we could go with as little as:

```golang
type IConn interface {
	ReadJSON(any) error
	WriteJSON(any) error
	Close() error
  // add more methods if needed by runWs implementation
}

func runWs(conn *IConn) {}
```

Now `runWs` can receive Gorilla `websocket.Conn` in real usage and `wsmock.GorillaConn` when testing.

Alternatively and instead of defining your own `IConn`, you can rely on `wsmock.IGorilla` interface: it declares all methods available on Gorilla [websocket.Conn](https://pkg.go.dev/github.com/gorilla/websocket#Conn):

```golang
import (
	ws "github.com/silently/wsmock"
)

func runWs(conn *ws.IGorilla) {}
```

## Usage and assertions

Once `runWs` accepts `wsmock.GorillaConn` as an argument, a test looks like:

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
  // runWs is the target of this test, supposedly implemented elsewhere in mypackage
  t.Run("two peers can connect and exchange hellos", func(t *testing.T) {
    conn1, rec1 := wsmock.NewGorillaMockAndRecorder(t)
    conn2, rec2 := wsmock.NewGorillaMockAndRecorder(t)
    runWs(conn1) 
    runWs(conn2)

    // script events (Johnny connects too late to receive Micheline's greeting)
    conn1.Send(Message{"join", "Micheline"})
    conn1.Send(Message{"send", "Bonjour"})
    conn2.Send(Message{"join", "Johnny"})
    conn2.Send(Message{"send", "Hello"})

    rec1.AssertReceived(Message{"incoming", "Hello"})
    // the next assertion is "not received" (supposing chat history is not implemented)
    rec2.AssertNotReceived(Message{"incoming", "Bonjour"})
    // run all assertions in this test, with a timeout
    wsmock.Run(t, 100 * time.Millisecond)
    // or run per recorder: rec1.Run(100 * time.Millisecond)
  })
}
```

Assertions are run either:

- per recorder, for instance `rec1.Run(100 * time.Millisecond)` followed by `rec2.Run(100 * time.Millisecond)`
- per test: `wsmock.Run(t, 100 * time.Millisecond)` (all recorders created with `t` in ` wsmock.NewGorillaMockAndRecorder(t)` will be ran)

`wsmock.NewGorillaConnAndRecorder` returns two structs:

- `wsmock.GorillaConn`, the mocked WebSocket connection given to `runWs`
- `wsmock.Recorder`, server writes recorder used to define assertions

The only methods you're supposed to use on `wsmock.GorillaConn` in the tests are:

- `Send(message any)` to script sent messages
- `Close()` if you want to explicitely close connections "client-side" (alternatively, wsmock will close them when test ends)

Assertions provided by `wsmock.Recorder` are (check the [API documentation here](TODO add godoc)) :

```golang
func (r *Recorder) AssertReceived(target any)
func (r *Recorder) AssertFirstReceived(target any)
func (r *Recorder) AssertLastReceivedOnTimeout(target any)
func (r *Recorder) AssertNotReceived(target any)
func (r *Recorder) AssertReceivedContains(substr string)
func (r *Recorder) AssertClosed()
func (r *Recorder) AssertReceivedSparseSequence(targets []any)   // 2,4 is a sparse sequence within 1,2,3,4
func (r *Recorder) AssertReceivedAdjacentSequence(targets []any) // 3,4 is an adjacent sequence within 1,2,3,4
func (r *Recorder) AssertReceivedExactSequence(targets []any)    // only 1,2,3,4 is an exact match of 1,2,3,4
```

*(assertions rely on the equality operator `==`, see [spec](https://go.dev/ref/spec#Comparison_operators))*

## Custom Assertions

You can define custom assertions with `func (r *Recorder) AssertWith(asserter Asserter)` where `Asserter` type is:

```golang
type Asserter func(end bool, latestWrite any, allWrites []any) (done, passed bool, errorMessage string)
```

With the following behaviour:
- when a write occurs (from the WebSocket server handler, like `runWs` previously), `Asserter` is called with `(false, latestWrite, allWritesIncludingLatest)` and you have to decide if the assertion outcome is known (`done` return value). If `done` is true, also return the test outcome (`passed`) and possibly an error message
- when assertions timeout is reached, `Asserter` is called one last time with `(true, latestWrite, allWritesIncludingLatest)`. The return value `done` is considered true whatever is returned, and `passed` and `errorMessage` give the test outcome 
 
For instance here is `AssertReceived` implementation, please note it can return sooner (if test passes) or later (if timeout is reached):

```golang
func (r *Recorder) AssertReceived(target any) {
	r.t.Helper()

	r.AssertWith(func(end bool, latestWrite any, _ []any) (done, passed bool, errorMessage string) {
		if end { // timeout has been reached
			done = true
			passed = false // if hasn't passed before, must be failing
			errorMessage = fmt.Sprintf("[wsmock] message not received: %v", target)
		} else if latestWrite == target {
			done = true
			passed = true
		}
		return
	})
}
```

## Implementation specifics

A typical flow of messages in a test goes like (considering a `runWs` server handler):
- `conn.Send("input")` -> recorder serverReadCh channel -> read by `runWs` (typically with `ReadJSON` or `ReadMessage`)
- then `runWs` processes the input message
- then `runWs` possibly writes a message (typically with `WriteJSON` or `WriteMessage`) -> recorder serverWriteCh channel -> forwarded by the recorder to each assertion defined on it

Here are some gotchas:
- `conn.Send(message any)` ensures messages are processed in arrival's order on the same `conn`, but depending on your WebSocket server handler implementation, there is no guarantee that messages sent on **several** conns will be processed in the same order the were sent
- all messages sent to the WebSocket server handler (`conn.Send(message any)`) or written by it (`WriteJSON` for instance) go through 256 buffered channels on wsmock `Recorder` type
- the `Recorder` stores all the messages written by the server handler: indeed some assertions needs to know the complete history of messages to decide their outcome
- **but** the message history is cleared after each run (`wsmock.Run(t, timeout)` or `rec.Run(timeout)`)

## For wsmock developers

Run wsmock own tests with:
```sh
CGO_ENABLED=0 go test .
```

Generate coverage reports:

```sh
CGO_ENABLED=0 go test -v -coverprofile cover.out
go tool cover -html cover.out -o cover.html
open cover.html
```