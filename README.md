```
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
    rec.RunAssertions(300 * time.Millisecond)
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