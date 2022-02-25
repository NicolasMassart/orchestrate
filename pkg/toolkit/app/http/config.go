package http

import (
	"crypto/tls"
	"time"

	"github.com/consensys/orchestrate/pkg/toolkit/app/auth/key"
	"github.com/spf13/viper"
)

type Config struct {
	XAPIKey               string
	Authorization         string
	Timeout               time.Duration
	KeepAlive             time.Duration
	IdleConnTimeout       time.Duration
	TLSHandshakeTimeout   time.Duration
	ExpectContinueTimeout time.Duration
	ClientCert            *tls.Certificate
	MaxIdleConnsPerHost   int
	InsecureSkipVerify    bool
	MultiTenancy          bool
	AuthHeaderForward     bool
}

func NewDefaultConfig() *Config {
	return &Config{
		MaxIdleConnsPerHost:   200,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		Timeout:               30 * time.Second,
		KeepAlive:             30 * time.Second,
		AuthHeaderForward:     true,
		InsecureSkipVerify:    false,
	}
}

func NewConfig(vipr *viper.Viper) *Config {
	cfg := NewDefaultConfig()
	cfg.XAPIKey = vipr.GetString(key.APIKeyViperKey)
	return cfg
}
