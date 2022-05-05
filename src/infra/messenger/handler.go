package messenger

import (
	"context"
)

type MessageHandler func(ctx context.Context, rawReq []byte) error
type ConsumerRequestMessageType string
