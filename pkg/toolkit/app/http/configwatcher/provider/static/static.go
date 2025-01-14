package static

import (
	"context"
	"fmt"

	"github.com/consensys/orchestrate/pkg/toolkit/app/http/configwatcher/provider"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
)

type Provider struct {
	msg    provider.Message
	logger *log.Logger
}

func New(msg provider.Message) *Provider {
	return &Provider{
		msg:    msg,
		logger: log.NewLogger().SetComponent("configwatcher"),
	}
}

func (p *Provider) Provide(_ context.Context, msgs chan<- provider.Message) error {
	p.logger.WithField("provider", fmt.Sprintf("%T", p)).
		Trace("start providing")
	msgs <- p.msg
	return nil
}
