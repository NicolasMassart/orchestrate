package sarama

import (
	"time"

	"github.com/consensys/orchestrate/src/infra/kafka"

	"github.com/Shopify/sarama"
	"github.com/hashicorp/go-multierror"
	healthz "github.com/heptiolabs/healthcheck"
)

type Client struct {
	sarama.Client
	addrs []string
}

var _ kafka.Client = &Client{}

func NewClient(cfg *Config, addrs []string) (*Client, error) {
	saramaCfg, err := cfg.ToKafkaConfig()
	if err != nil {
		return nil, err
	}

	client, err := sarama.NewClient(addrs, saramaCfg)
	if err != nil {
		return nil, err
	}

	return &Client{client, addrs}, nil
}

func (c *Client) Checker() error {
	gr := &multierror.Group{}
	for _, host := range c.addrs {
		gr.Go(healthz.TCPDialCheck(host, time.Second*3))
	}

	return gr.Wait().ErrorOrNil()
}
