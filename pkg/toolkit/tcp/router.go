package tcp

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"io"
	"net"
	"strings"
	"time"

	"github.com/containous/traefik/v2/pkg/log"
)

// Router is a TCP router
type Router struct {
	routingTable   map[string]Handler
	tcpForwarder   Handler
	tlsForwarder   Handler
	tlsConfig      *tls.Config // default TLS config
	catchAllNoTLS  Handler
	hostTLSConfigs map[string]*tls.Config // TLS configs keyed by SNI
}

// ServeTCP forwards the connection to the right TCP/HTTP handler
func (r *Router) ServeTCP(conn WriteCloser) {
	if r.catchAllNoTLS != nil && len(r.routingTable) == 0 {
		r.catchAllNoTLS.ServeTCP(conn)
		return
	}

	br := bufio.NewReader(conn)
	serverName, isTLS, peeked, err := clientHelloServerName(br)
	if err != nil {
		conn.Close()
		return
	}

	// Remove read/write deadline and delegate this to underlying tcp server (for now only handled by HTTP Server)
	err = conn.SetReadDeadline(time.Time{})
	if err != nil {
		log.WithoutContext().WithError(err).Errorf("Error while setting read deadline")
	}

	err = conn.SetWriteDeadline(time.Time{})
	if err != nil {
		log.WithoutContext().WithError(err).Errorf("Error while setting write deadline")
	}

	if !isTLS {
		switch {
		case r.catchAllNoTLS != nil:
			r.catchAllNoTLS.ServeTCP(r.GetConn(conn, peeked))
		case r.tcpForwarder != nil:
			r.tcpForwarder.ServeTCP(r.GetConn(conn, peeked))
		default:
			conn.Close()
		}
		return
	}

	// FIXME Optimize and test the routing table before helloServerName
	serverName = strings.ToLower(serverName)
	if r.routingTable != nil && serverName != "" {
		if target, ok := r.routingTable[serverName]; ok {
			target.ServeTCP(r.GetConn(conn, peeked))
			return
		}
	}

	// FIXME Needs tests
	if target, ok := r.routingTable["*"]; ok {
		target.ServeTCP(r.GetConn(conn, peeked))
		return
	}

	if r.tlsForwarder != nil {
		r.tlsForwarder.ServeTCP(r.GetConn(conn, peeked))
	} else {
		conn.Close()
	}
}

// AddRoute defines a handler for a given sniHost (* is the only valid option)
func (r *Router) AddRoute(sniHost string, target Handler) {
	if r.routingTable == nil {
		r.routingTable = map[string]Handler{}
	}
	r.routingTable[strings.ToLower(sniHost)] = target
}

// AddRouteTLS defines a handler for a given sniHost and sets the matching tlsConfig
func (r *Router) AddRouteTLS(sniHost string, target Handler, config *tls.Config) {
	r.AddRoute(sniHost, &TLSHandler{
		Next:   target,
		Config: config,
	})
}

// AddRouteHTTPTLS defines a handler for a given sniHost and sets the matching tlsConfig
func (r *Router) AddRouteHTTPTLS(sniHost string, config *tls.Config) {
	if r.hostTLSConfigs == nil {
		r.hostTLSConfigs = map[string]*tls.Config{}
	}
	r.hostTLSConfigs[sniHost] = config
}

// AddCatchAllNoTLS defines the fallback tcp handler
func (r *Router) AddCatchAllNoTLS(handler Handler) {
	r.catchAllNoTLS = handler
}

// GetConn creates a connection proxy with a peeked string
func (r *Router) GetConn(conn WriteCloser, peeked string) WriteCloser {
	// FIXME should it really be on Router ?
	conn = &Conn{
		Peeked:      []byte(peeked),
		WriteCloser: conn,
	}
	return conn
}

// tcpForwarder sets the tcp handler that will forward the connections to an http handler
func (r *Router) TCPForwarder(handler Handler) {
	r.tcpForwarder = handler
}

// TLSConfig set router TLS configuration
func (r *Router) TLSConfig(config *tls.Config) {
	r.tlsConfig = config
}

// tlsForwarder sets the tcp handler that will forward the TLS connections to an http handler
func (r *Router) TLSForwarder(handler Handler) {
	for sniHost, tlsConf := range r.hostTLSConfigs {
		r.AddRouteTLS(sniHost, handler, tlsConf)
	}

	r.tlsForwarder = &TLSHandler{
		Next:   handler,
		Config: r.tlsConfig,
	}
}

// Conn is a connection proxy that handles Peeked bytes
type Conn struct {
	// Peeked are the bytes that have been read from Conn for the
	// purposes of route matching, but have not yet been consumed
	// by Read calls. It set to nil by Read when fully consumed.
	Peeked []byte

	// Conn is the underlying connection.
	// It can be type asserted against *net.TCPConn or other types
	// as needed. It should not be read from directly unless
	// Peeked is nil.
	WriteCloser
}

// Read reads bytes from the connection (using the buffer prior to actually reading)
func (c *Conn) Read(p []byte) (n int, err error) {
	if len(c.Peeked) > 0 {
		n = copy(p, c.Peeked)
		c.Peeked = c.Peeked[n:]
		if len(c.Peeked) == 0 {
			c.Peeked = nil
		}
		return n, nil
	}
	return c.WriteCloser.Read(p)
}

// clientHelloServerName returns the SNI server name inside the TLS ClientHello,
// without consuming any bytes from br.
// On any error, the empty string is returned.
func clientHelloServerName(br *bufio.Reader) (sni string, isTLS bool, peeked string, err error) {
	hdr, err := br.Peek(1)
	if err != nil {
		opErr, ok := err.(*net.OpError)
		if err != io.EOF && (!ok || !opErr.Timeout()) {
			log.WithoutContext().Debugf("Error while Peeking first byte: %s", err)
		}
		return "", false, "", err
	}

	// No valid TLS record has a type of 0x80, however SSLv2 handshakes
	// start with a uint16 length where the MSB is set and the first record
	// is always < 256 bytes long. Therefore typ == 0x80 strongly suggests
	// an SSLv2 client.
	const recordTypeSSLv2 = 0x80
	const recordTypeHandshake = 0x16
	if hdr[0] != recordTypeHandshake {
		if hdr[0] == recordTypeSSLv2 {
			// we consider SSLv2 as TLS and it will be refuse by real TLS handshake.
			return "", true, getPeeked(br), nil
		}
		return "", false, getPeeked(br), nil // Not TLS.
	}

	const recordHeaderLen = 5
	hdr, err = br.Peek(recordHeaderLen)
	if err != nil {
		log.Errorf("Error while Peeking hello: %s", err)
		return "", false, getPeeked(br), nil
	}

	recLen := int(hdr[3])<<8 | int(hdr[4]) // ignoring version in hdr[1:3]
	helloBytes, err := br.Peek(recordHeaderLen + recLen)
	if err != nil {
		log.Errorf("Error while Hello: %s", err)
		return "", true, getPeeked(br), nil
	}

	server := tls.Server(sniSniffConn{r: bytes.NewReader(helloBytes)}, &tls.Config{
		GetConfigForClient: func(hello *tls.ClientHelloInfo) (*tls.Config, error) {
			sni = hello.ServerName
			return nil, nil
		},
	})
	_ = server.Handshake()

	return sni, true, getPeeked(br), nil
}

func getPeeked(br *bufio.Reader) string {
	peeked, err := br.Peek(br.Buffered())
	if err != nil {
		log.Errorf("Could not get anything: %s", err)
		return ""
	}
	return string(peeked)
}

// sniSniffConn is a net.Conn that reads from r, fails on Writes,
// and crashes otherwise.
type sniSniffConn struct {
	r        io.Reader
	net.Conn // nil; crash on any unexpected use
}

// Read reads from the underlying reader
func (c sniSniffConn) Read(p []byte) (int, error) { return c.r.Read(p) }

// Write crashes all the time
func (sniSniffConn) Write(p []byte) (int, error) { return 0, io.EOF }
