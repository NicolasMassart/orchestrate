package entities

type RequestMessageType string

type Message struct {
	Type   RequestMessageType
	Body   []byte
	Offset int64
	Commit func() error `json:"-"`
}
