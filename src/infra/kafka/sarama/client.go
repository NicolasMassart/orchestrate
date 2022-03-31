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

func NewClient(cfg *sarama.Config, addrs []string) (*Client, error) {
	client, err := sarama.NewClient(addrs, cfg)
	if err != nil {
		return nil, err
	}

	// Retrieve and log connected brokers
	var brokers = make(map[int32]string)
	for _, v := range client.Brokers() {
		brokers[v.ID()] = v.Addr()
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
