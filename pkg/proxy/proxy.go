package proxy

import (
	"context"
	"log"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/http/httputil"
	"time"

	"github.com/esnunes/ratelimit"
)

const (
	// BucketHeaderKey ...
	BucketHeaderKey = "x-ratelimit-bucket"
	// HostHeaderKey ...
	HostHeaderKey = "x-ratelimit-host"
	// SchemeHeaderKey ...
	SchemeHeaderKey = "x-ratelimit-scheme"
)

type contextKey string

const (
	bucketContextKey = contextKey("bucket")
)

// Server ...
type Server struct {
	reverseProxy httputil.ReverseProxy
	manager      *ratelimit.Manager
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	bucket := r.Header.Get(BucketHeaderKey)
	if bucket == "" {
		http.Error(w, "invalid "+BucketHeaderKey+" header value", http.StatusBadRequest)
		return
	}

	err := s.manager.Get(bucket).Take()

	if err != nil {
		switch err {
		case ratelimit.ErrNoSlotsAvailable:
			code := http.StatusTooManyRequests
			http.Error(w, http.StatusText(code), code)
		default:
			log.Printf("Unexpected error: %v", err)

			code := http.StatusInternalServerError
			http.Error(w, http.StatusText(code), code)
		}

		return
	}

	trace := &httptrace.ClientTrace{
		GotConn: func(info httptrace.GotConnInfo) {
			s.manager.Get(bucket).Release()
		},
	}

	ctx := context.WithValue(r.Context(), bucketContextKey, bucket)
	ctx = httptrace.WithClientTrace(ctx, trace)

	s.reverseProxy.ServeHTTP(w, r.WithContext(ctx))
}

// ServerOptions ...
type ServerOptions struct {
	Burst int
	Rate  float64
	Queue int
}

// New ...
func New(o ServerOptions) *Server {
	mconfig := ratelimit.Config{
		Rate:  time.Duration(float64(time.Second) / o.Rate),
		Burst: o.Burst,
		Queue: o.Queue,
	}

	manager := &ratelimit.Manager{
		Config: mconfig,
	}

	director := func(r *http.Request) {
		r.URL.Scheme = r.Header.Get(SchemeHeaderKey)
		if r.URL.Scheme == "" {
			r.URL.Scheme = "https"
		}

		r.URL.Host = r.Header.Get(HostHeaderKey)

		r.Header.Del(BucketHeaderKey)
		r.Header.Del(SchemeHeaderKey)
		r.Header.Del(HostHeaderKey)

		r.Host = r.URL.Host
	}

	rp := httputil.ReverseProxy{
		Director:  director,
		Transport: NewTransport(manager),
	}

	return &Server{
		reverseProxy: rp,
		manager:      manager,
	}
}

// NewTransport ...
func NewTransport(manager *ratelimit.Manager) http.RoundTripper {
	defaultDialContext := (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}).DialContext

	newDialContext := func(ctx context.Context, network, addr string) (net.Conn, error) {
		bucket, _ := ctx.Value(bucketContextKey).(string)

		c, err := defaultDialContext(ctx, network, addr)
		if err != nil {
			manager.Get(bucket).Release()
		}

		return c, err
	}

	return &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           newDialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}
