package wsmock

type GorillaConn interface {
	ReadJSON(any) error
	WriteJSON(any) error
	Close() error
}

type Finder func(messages []any) bool
