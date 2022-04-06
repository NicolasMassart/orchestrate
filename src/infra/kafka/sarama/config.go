package sarama

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"time"

	"github.com/Shopify/sarama"
)

var rebalanceStrategy = map[string]sarama.BalanceStrategy{
	"Range":      sarama.BalanceStrategyRange,
	"RoundRobin": sarama.BalanceStrategyRoundRobin,
	"Sticky":     sarama.BalanceStrategySticky,
}

type Config struct {
	URLs              []string
	Version           string
	ClientID          string
	SASLMechanism     string
	SASLUser          string
	SASLPassword      string
	SASLSCRAMAuthzID  string
	TLSCert           string
	TLSKey            string
	TLSCA             string
	GroupName         string
	RebalanceStrategy string
	MaxWaitTime       time.Duration
	MaxProcessingTime time.Duration
	SessionTimeout    time.Duration
	HeartbeatInterval time.Duration
	RebalanceTimeout  time.Duration
	SASLEnable        bool
	SASLHandshake     bool
	TSLEnable         bool
	TLSSkipVerify     bool
	ReturnErrors      bool
	ReturnSuccesses   bool
	AutoCommit        bool
}

func (cfg *Config) ToKafkaConfig() (*sarama.Config, error) {
	saramaCfg := sarama.NewConfig()

	var err error
	saramaCfg.Version, err = sarama.ParseKafkaVersion(cfg.Version)
	if err != nil {
		return nil, err
	}

	saramaCfg.ClientID = cfg.ClientID

	// Consumer config
	saramaCfg.Consumer.Return.Errors = cfg.ReturnErrors
	saramaCfg.Consumer.Offsets.AutoCommit.Enable = cfg.AutoCommit
	saramaCfg.Consumer.MaxWaitTime = cfg.MaxWaitTime
	saramaCfg.Consumer.MaxProcessingTime = cfg.MaxProcessingTime
	saramaCfg.Consumer.Group.Session.Timeout = cfg.SessionTimeout
	saramaCfg.Consumer.Group.Heartbeat.Interval = cfg.HeartbeatInterval
	saramaCfg.Consumer.Group.Rebalance.Timeout = cfg.RebalanceTimeout
	if strategy, ok := rebalanceStrategy[cfg.RebalanceStrategy]; ok {
		saramaCfg.Consumer.Group.Rebalance.Strategy = strategy
	}

	// Producer config
	saramaCfg.Producer.Return.Errors = cfg.ReturnErrors
	saramaCfg.Producer.Return.Successes = cfg.ReturnSuccesses

	saramaCfg.Net.SASL.Enable = cfg.SASLEnable
	saramaCfg.Net.SASL.Mechanism = sarama.SASLMechanism(cfg.SASLMechanism)
	saramaCfg.Net.SASL.Handshake = cfg.SASLHandshake
	saramaCfg.Net.SASL.User = cfg.SASLUser
	saramaCfg.Net.SASL.Password = cfg.SASLPassword
	saramaCfg.Net.SASL.SCRAMAuthzID = cfg.SASLSCRAMAuthzID

	saramaCfg.Net.TLS.Enable = cfg.TSLEnable
	if saramaCfg.Net.TLS.Enable {
		saramaCfg.Net.TLS.Config, err = cfg.getTLSConfig()
		if err != nil {
			return nil, err
		}
	}

	if err := saramaCfg.Validate(); err != nil {
		return nil, err
	}

	return saramaCfg, nil
}

func (cfg *Config) getTLSConfig() (*tls.Config, error) {
	tlsConfig := tls.Config{}
	var err error

	if cfg.TLSCert != "" && cfg.TLSKey != "" {
		// Load client cert
		var cert tls.Certificate
		cert, err = tls.LoadX509KeyPair(cfg.TLSCert, cfg.TLSKey)
		if err != nil {
			return &tls.Config{}, err
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	if cfg.TLSCert != "" {
		// Load CA cert
		var caCert []byte
		caCert, err = ioutil.ReadFile(cfg.TLSCA)
		if err != nil {
			return &tls.Config{}, err
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tlsConfig.RootCAs = caCertPool
	}

	tlsConfig.InsecureSkipVerify = cfg.TLSSkipVerify

	return &tlsConfig, err
}
