package httpcache

import (
	"context"
	"time"

	"github.com/dgraph-io/ristretto"
)

//go:generate mockgen -source=manager.go -destination=mocks/manager.go -package=mocks

type CacheManager interface {
	Get(context.Context, string) ([]byte, bool)
	Set(context.Context, string, []byte) bool
	SetWithTTL(context.Context, string, []byte, time.Duration) bool
	TTL() time.Duration
}

type cacheManager struct {
	c   *ristretto.Cache
	ttl time.Duration
}

func newManager(
	c *ristretto.Cache,
	ttl time.Duration,
) CacheManager {
	return &cacheManager{
		c:   c,
		ttl: ttl,
	}
}

func (cca *cacheManager) Get(_ context.Context, key string) ([]byte, bool) {
	if v, ok := cca.c.Get(key); ok {
		return v.([]byte), true
	}
	return nil, false
}

func (cca *cacheManager) Set(ctx context.Context, key string, value []byte) bool {
	return cca.SetWithTTL(ctx, key, value, cca.ttl)
}

func (cca *cacheManager) SetWithTTL(_ context.Context, key string, value []byte, ttl time.Duration) bool {
	return cca.c.SetWithTTL(key, value, int64(len(value)), ttl)
}

func (cca *cacheManager) TTL() time.Duration {
	return cca.ttl
}
