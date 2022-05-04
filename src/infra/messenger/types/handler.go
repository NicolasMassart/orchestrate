package types

import (
	"context"
)

type MessageHandler func(ctx context.Context, rawReq []byte) error
