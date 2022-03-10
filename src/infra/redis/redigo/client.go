package redigo

import (
	"github.com/consensys/orchestrate/src/infra/redis"
	redigo "github.com/gomodule/redigo/redis"
)

type Client struct {
	pool *redigo.Pool
}

var _ redis.Client = &Client{}

func New(cfg *Config) (*Client, error) {
	redisOptions, err := cfg.ToRedisOptions()
	if err != nil {
		return nil, err
	}

	pool := &redigo.Pool{
		MaxIdle:     cfg.MaxIdle,
		IdleTimeout: cfg.IdleTimeout,
		Dial: func() (redigo.Conn, error) {
			return redigo.Dial("tcp", cfg.URL(), redisOptions...)
		},
	}

	return &Client{pool: pool}, nil
}

func (nm *Client) LoadUint64(key string) (uint64, error) {
	conn := nm.pool.Get()
	defer closeConn(conn)

	reply, err := conn.Do("GET", key)
	if err != nil {
		return 0, parseRedisError(err)
	}

	value, err := redigo.Uint64(reply, nil)
	if err != nil {
		return 0, parseRedisError(err)
	}

	return value, nil
}

func (nm *Client) Set(key string, expiration int, value interface{}) error {
	conn := nm.pool.Get()
	defer closeConn(conn)

	// Set value with expiration
	_, err := conn.Do("PSETEX", key, expiration, value)
	if err != nil {
		return parseRedisError(err)
	}

	return nil
}

func (nm *Client) Delete(key string) error {
	conn := nm.pool.Get()
	defer closeConn(conn)

	// Delete value
	_, err := conn.Do("DEL", key)
	if err != nil {
		return parseRedisError(err)
	}

	return nil
}

func (nm *Client) Incr(key string) error {
	conn := nm.pool.Get()
	defer closeConn(conn)

	_, err := conn.Do("INCR", key)
	if err != nil {
		return parseRedisError(err)
	}

	return nil
}

func (nm *Client) Ping() error {
	conn := nm.pool.Get()
	defer closeConn(conn)

	_, err := conn.Do("PING")
	if err != nil {
		return parseRedisError(err)
	}

	return nil
}

func closeConn(conn redigo.Conn) {
	// There is nothing we can do if the connection fails to close
	_ = conn.Close()
}
