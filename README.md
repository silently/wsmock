# wsmock

Golang library to help with WebSocket testing by providing Gorilla mocks, assertion scripting, assertion parallelism and timeouts. wsmock is itself thoroughly tested and logs meaningful errors.

With wsmock, tests look like:

```golang
func TestRockPaperScissors(t *testing.T) {
  // initialize with mocked Conns and server-sent messages recorder
  // (similarly to httptest NewRequest and NewRecorder)
  conn1, rec1 := wsmock.NewGorillaMockAndRecorder(t)
  conn2, rec2 := wsmock.NewGorillaMockAndRecorder(t)

  // wsHandler is the test target
  go wsHandler(conn1)
  go wsHandler(conn2)

  // script interaction
  conn1.Send("paper")
  conn2.Send("paper")
  // and then
  conn1.Send("paper")
  conn2.Send("rock")

  // assertions are defined as ordered conditions
  rec1.NewAssertion().
    OneToBe("draw").  // one message received is expected to be "draw"
    NextToBe("win")   // then the next message is expected to be "win"
  rec2.NewAssertion().
    OneToBe("draw").
    NextToBe("loss")

  // run assertions for rec1 and rec2, with a timeout
  wsmock.RunAssertions(t, 100*time.Millisecond) 
}      
```

...where `wsHandler` is a WebSocket handler based on [Gorilla WebSocket](https://github.com/gorilla/websocket), typically called in a HTTP handler:

```golang
import (
  "log"
  "net/http"

  "github.com/gorilla/websocket"
)

func serveWs(w http.ResponseWriter, r *http.Request) { // HTTP handler
  conn, err := upgrader.Upgrade(w, r, nil)             // creates a Gorilla websocket.Conn
  if err != nil {
    log.Println(err)
  }
  wsHandler(conn)                                      // WebSocket handler -> target of the test
}
```

## Status

wsmock is in an early stage of development (API may evolve) but has itself a good test coverage.

Currently, only Gorilla WebSocket mocks are provided (more WebSocket implementation mocks could be considered) with a focus on reading from and writing to the Conn:

- that's why we provide mock implementations for the methods: `Close`, `ReadJSON`, `ReadMessage`, `NextReader`, `NextWriter`, `WriteJSON`, `WriteMessage`
- but other methods (like  `CloseHandler`, `EnableWriteCompression`...) from Gorilla `websocket.Conn` are blank/noop

*(wsmock test coverage does not reach 100% because of these blank/noop implementations: they will only be tested when a proper implementation is considered)*

## Installation

```sh
go get github.com/silently/wsmock
```

## Prerequesite

**In short**: replace Gorilla's `*websocket.Conn` type with `wsmock.IGorilla` (or your own compliant interface) to be able to use a mocked conn in tests.

Indeed, the main gotcha is to be able to call WebSocket handlers:

- with Gorilla's `*websocket.Conn` in the main app code
- with `*wsmock.GorillaConn` in tests

For instance, if `wsHandler` current signature is:

```golang
func wsHandler(conn *websocket.Conn) {}
```

...we need to update it with an interface implemented both by Gorilla `websocket.Conn` and `wsmock.GorillaConn`.

*(This approach is similar to [httptest](https://pkg.go.dev/net/http/httptest#example-ResponseRecorder) that relies on `ResponseRecorder`, "an implementation of `http.ResponseWriter`; that records its mutations for later inspection in tests")*

Depending on what methods are used within `wsHandler` we could go with as little as:

```golang
type IConn interface {
  ReadJSON(any) error
  WriteJSON(any) error
  Close() error
  // add more methods if needed by wsHandler implementation
}

func wsHandler(conn *IConn) {}
```

Now `wsHandler` can receive both Gorilla `*websocket.Conn` in real usage and `*wsmock.GorillaConn` when testing.

Alternatively and instead of defining your own `IConn`, you can rely on `wsmock.IGorilla` interface: it declares all methods available on Gorilla [websocket.Conn](https://pkg.go.dev/github.com/gorilla/websocket#Conn):

```golang
import (
  ws "github.com/silently/wsmock"
)

func wsHandler(conn *ws.IGorilla) {}
```

## Example

Let's review a `wsmock` example:

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

func TestChat(t *testing.T) {
  t.Run("two peers can connect and exchange hellos", func(t *testing.T) {
    // 2 users interact via WebSockets
    johnnyConn, johnnyRec := wsmock.NewGorillaMockAndRecorder(t)
    michelineConn, michelineRec := wsmock.NewGorillaMockAndRecorder(t)

    // chatHandler is the target of this test, supposedly implemented elsewhere in mypackage
    go chatHandler(johnnyConn) 
    go chatHandler(michelineConn)

    // round #1 of Sends and Asserts
    johnnyConn.Send(Message{"join", "Johnny"})
    johnnyConn.Send(Message{"send", "Hello"})
    // Micheline connects too late to receive Johnny's greeting
    michelineConn.Send(Message{"join", "Micheline"})
    michelineConn.Send(Message{"send", "Salut"})

    johnnyRec.NewAssertion().OneToBe(Message{"incoming", "Salut"})
    // the next assertion is "not received" (supposing chat history is not implemented)
    michelineRec.NewAssertion().OneNotToBe(Message{"incoming", "Hello"})
    // run all assertions in this test, with a timeout
    wsmock.RunAssertions(t, 100 $ time.M)

    // round #2 of Sends and Asserts
    johnnyConn.Send(Message{"send", "Are you French?"})
    johnnyConn.Send(Message{"send", "Sorry I only speak English"})
    // it's possible to chain assertions, order matters
    michelineRec.NewAssertion().
      OneToContain("French").
      OneToContain("English")
    // you can run assertions on a given recorder, with a timeout
    michelineRec.RunAssertions(100 * time.Millisecond)
  })
}
```

Thanks to `wsmock.NewGorillaConnAndRecorder` we get two structs:

- `wsmock.GorillaConn` the mocked WebSocket connection given to `wsHandler`
- `wsmock.Recorder` that records what is written by `chatHandler` to `GorillaConn` and propose an API to define assertions regarding these writes

Methods you're supposed to use on `wsmock.GorillaConn` to script the tests are:

- `Send(message any)` to script sent messages
- `Close()` if you want to explicitely close connections "client-side" (alternatively, wsmock will close them when test ends)

Then you add assertions on recorders (`rec.NewAssertion().Condition1(…).Condition2(…)`), see more in the next paragraph about wsmock condition ordering and chaining features.

You can run assertions either:

- per recorder, for instance `michelineRec.Run(100 * time.Millisecond)`
- per test: `wsmock.RunAssertions(t, 100 * time.Millisecond)` (all recorders created with `t` in ` wsmock.NewGorillaMockAndRecorder(t)` will be ran)

After `RunAssertions(…)` is finished, the message history on recorders is emptied and `wsmock` internally creates a new *round* of events. It means you can pursue scripting your test with `conn.Send(…)`, define and run new assertions on recorders, but messages from previous rounds won't be taken into account in the current round.

## Assertion concepts

With wsmock we define assertions as ordered chains of conditions (that will be validated upon messages received by a recorder). The assertion succeeds only if all of the conditions in the chain succeed in order:

```golang
rec.NewAssertion().  // this
  NextToBe("a").     // is
  NextToBe("b").     // an
  OneToBe("d")       // assertion
rec.RunAssertions(100 * time.Millisecond)
```

The preceding assertion will for instance succeed if the message history is `["a", "b", "c", "d"]` but fail with `["a", "d", "b"]` (wrong order) or `["z" "a", "b", "d"]` (unexpected first element).

Assertions may contain only one condition and you can run several assertions in parallel (with multiple calls to `NewAssertion`):

```golang
// both succeed if the only message received is "this is a short sentence"
rec.NewAssertion().OneToContain("short")    // assertion 1
rec.NewAssertion().OneToContain("sentence") // assertion 2
// but this one fails since it is supposed to match two different messages
rec.NewAssertion(). 
  OneToContain("short").
  OneToContain("sentence")
```

To sum-up, if several conditions have been attached to a recorder, they will be run in parallel, each assertion receiving the same sequence of messages from the recorder. In contrast, conditions (the building blocks of assertions) are run in series (one after the other: order matters).

## Assertion API

Check the [API documentation](https://pkg.go.dev/github.com/silently/wsmock#Assertion) for a comprehensive description of wsmock Assertion API.

`Recorder#NewAssertion()` returns a new struct that enables chaining and ordering conditions.

### Chainable conditions

The name of chainable condition methods is any combination of `Next|One + To|NotTo + Be|Check|Contain|Match`, with the following meaning:

- Prefix:
  - `Next*` means the condition should be true on the very next received message
  - `One*` means one among subsequent messages should verify the condition
- Condition:
  - `*ToBe(target any)` is successful if the message equals `target` (according to the equality operator `==`, see [spec](https://go.dev/ref/spec#Comparison_operators))
  - `*ToCheck(f Predicate)` is successful if `predicate(msg)` is true
  - `*ToContain(sub string)` is successful if the message contains `sub`
  - `*ToMatch(re regexp.Regexp)` is successful if the message contains a match of `re`
  - while `*NotTo*`s evaluate to the opposite

Here are some example:

- `rec.NewAssertion().OneToBe(target)`: one message should be equal to `target`
- `rec.NewAssertion().NextToCheck(f)`: the next (or first in that case) message should validate the predicate function `f`

### Closing conditions

The name of closing condition methods is any combination of `LastTo|LastNotTo|AllTo|NoneTo + Be|Check|Contain|Match`. They behave similarly to chaining conditions, with the following differences:

- they need the whole message history to succeed, meaning they can fail fast (for instance if `NoneToBe("a")` receives `"a"`) but can only succeed on end (timeout or connection close)
- there can be only one closing condition in an assertion (or the assertion will fail)
- Prefix:
  - `Last*` means the last message should verify the condition (fails when no message is received)
  - `All*` means all messages should verify the condition (succeeds when no message is received)
  - `None*` means no message should verify the condition (succeeds when no message is received)

If you want to test several closing conditions, don't chain them but instead create parallel assertions with multiple `.NewAssertion()`:

```golang
rec.NewAssertion().
  OneNotToBe("z").
  OneToBe("a").
  LastNoToBe("z")
rec.NewAssertion().
  NoneToMatch(regex)
```

### Condition evaluation order

Let's inspect the following assertion:

```golang
rec.NewAssertion().
  OneToBe("a").
  NextToBe("b").
  LastToBe("z")
```

The first active condition is `OneToBe("a")` and will be the only one receiving message until it is *done* (be it a success or failure). Each time a message is received `OneToBe("a")` is evaluated and its possible outcome is ternary:

- "not done" (for instance message `"?"` is received) → `OneToBe("a")` remains the only active condition
- "done and succeeded" (message `"a"` is received) → the active condition moves to `NextToBe("b")` and from now on it will receive new messages
- "done and failed" (end is reached without matching a message, either because the `RunAssertions` timeout is reached or because the conn is closed) → the assertion failed

To go on with this example, `NextToBe("b")` possible outcome is binary:

- "done and succeeded" (the next message is indeed `"b"`)
- "done and failed" (if the next message is not `"b"` or if end is reached)
- for the `Next*` conditions, "not done" is not a possible outcome

Now let say `NextToBe("b")` also succeeded and the active condition is `LastNoToBe("z")`. The assertion ends, triggering one last time the active condition.

To sum-up:
- as soon as a condition fails, the whole assertion fails and remaining conditions are not tested
- the set of possible outcomes for a condition is "not done", "done and succeeded" or "done and failed"
- the set of possible outcomes for an assertion is "succeeded" or "failed"
- when a `Next*` condition appears first in the chain (for instance `rec.NewAssertion().NextToBe("a")...`) it means it will only consider the first received message to evaluate its outcome

### OneNotToBe and NoneToBe

Beware of this potential confusion:

- `NoneToBe(x)` means "no message until end should be equal to x" → if no message is received the condition succeeds
- `OneNotToBe(x)` means "a message not equal to x is expected" → if no message is received the condition fails (and like other `One*` condition, only one message satisfying the condition is needed)

### With

The predefined set of conditions may not fit your needs. In that case you can define a custom `ConditionFunc`: 

```golang
type ConditionFunc func(end bool, latest any, all []any) (done, passed bool, err string)
```

With the following behaviour:
- when a message is received (sent from the WebSocket server handler to the recorder), `ConditionFunc` is called with `(false, latest, all)` and you have to decide if the assertion outcome is known (`done` return value). If `done` is true, you also need to return the test outcome (`passed`) and possibly an error message if `passed` is false.
- when end is reached (timeout or connection close), `ConditionFunc` is called one last time with `(true, latest, all)`. `done` is considered true whatever is returned, while `passed` and `err` do give the asserter outcome

Here is an example and how to add/chain it on a recorder:

```golang
func customCondition(end bool, latest any, _ []any) (done, passed bool, err string) {
  if end {
    done = true
    passed = false
    err = fmt.Sprintf("One message equal to %+v was expected", target)
  } else if latest == target {
    done = true
    passed = false
  }
  return
}

rec.NewAssertion().With(customCondition) // it's possible to chain it with other conditions
```

...this `customCondition` is a possible implementation of `OneNotToBe`.

## Implementation specifics

The flow of messages in a test goes like (considering a `wsHandler` server handler):
- `conn.Send("input")` → conn's serverReadCh channel → read by `wsHandler` (typically with `ReadJSON` or `ReadMessage`)
- then `wsHandler` processes the input message
- and/or/then `wsHandler` possibly writes messages (typically with `WriteJSON` or `WriteMessage`) → recorder serverWriteCh channel → forwarded by the recorder to each assertion declared on it with `NewAssertion()`

Here are some gotchas:
- `conn.Send(message any)` ensures messages are processed in arrival's order on the same `conn`, but depending on your WebSocket server handler implementation, there is no guarantee that messages sent on **several** conns will be processed in the same order they were sent
- all messages sent to the WebSocket server handler (`conn.Send(message any)`) or written by it (`WriteJSON` for instance) go through 512 buffered channels on wsmock `Recorder` type
- messages written by the server handler are stored until timeout is reached: indeed some assertions need to know the complete history of messages to decide their outcome
- **but** the message history is cleared after each run (`wsmock.RunAssertions(t, timeout)` or `rec.Run(timeout)`), which is important to know if you make several runs in the same test

## wsmock output

In case of a failing test, the output looks like:

```
--- FAIL: TestFailing (0.10s)
    --- FAIL: TestFailing/should_fail (0.10s)
        assert_failing_test.go:25: 
            In recorder#0 → assertion#1, 1 message received:
                "1"
            Error occured on write:
                [NextToBe] next message is not equal to: integration_test.Message{Kind:"chat", Payload:"notfound"}
                Failing message (of type string): 1
            
        assert_failing_test.go:25: 
            In recorder#0 → assertion#2, 3 messages received:
                "1"
                "2"
                "3"
            Error occured on end:
                [OneToCheck] no message checks predicate: github.com/silently/wsmock/integration_test_test.stringLongerThan3
```

Where:

- `recorder#0` uniquely identifies the failing recorder within `TestFailing` (`#0` maps the creation order of the recorder in `TestFailing`)
- `assertion#1` uniquely identifies the failing assertion of a given recorder (`#1` maps the creation order of the assertion on the recorder)
- messages received by the assertion are printed before the actual error

## For wsmock developers

Even if it does not follow Go conventions, we've decided to put additional tests in a separate `integration_test` folder to be able to split them in many different files without polluting the root folder.

Then you may run wsmock own tests with:

```sh
go test ./...
```

And generate coverage reports:

```sh
go test -coverpkg=./... -coverprofile cover.out ./...
go tool cover -html cover.out -o cover.html
open cover.html
```
