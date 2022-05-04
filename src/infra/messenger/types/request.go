package types

type ConsumerRequestMessageType string

type ConsumerRequestMessage struct {
	Type ConsumerRequestMessageType
	Body []byte
}
