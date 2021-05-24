// Copyright (c) OpenFaaS Author(s) 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package types

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	nrt "github.com/ripienaar/nats-roundtripper"
)

// NewHTTPClientReverseProxy proxies to an upstream host through the use of a http.Client
func NewHTTPClientReverseProxy(baseURL *url.URL, timeout time.Duration, maxIdleConns, maxIdleConnsPerHost int) *HTTPClientReverseProxy {
	h := HTTPClientReverseProxy{
		timeout: timeout,
	}

	h.client = http.DefaultClient
	h.timeout = timeout
	h.client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	// These overrides for the default client enable re-use of connections and prevent
	// CoreDNS from rate limiting the gateway under high traffic
	//
	// See also two similar projects where this value was updated:
	// https://github.com/prometheus/prometheus/pull/3592
	// https://github.com/minio/minio/pull/5860

	// Taken from http.DefaultTransport in Go 1.11
	h.client.Transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   timeout,
			KeepAlive: timeout,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          maxIdleConns,
		MaxIdleConnsPerHost:   maxIdleConnsPerHost,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	return &h
}

// NewNATSTransportClientReverseProxy creates a new HTTP CLient ReverseProxy that utilizes the NATS RoundTripper for the transit layer.
func NewNATSTransportClientReverseProxy(baseURL *url.URL, timeout time.Duration, address string, port int) *HTTPClientReverseProxy {
	h := HTTPClientReverseProxy{
		timeout: timeout,
	}

	h.client = http.DefaultClient
	h.timeout = timeout
	h.client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	h.client.Transport = nrt.Must(
		nrt.WithNatsServer(
			fmt.Sprintf("%s:%d", address, port),
		),
		nrt.WithPrefix("faas"),
	)

	return &h
}

// HTTPClientReverseProxy proxy to a remote BaseURL using a http.Client
type HTTPClientReverseProxy struct {
	client  *http.Client
	timeout time.Duration
}

// Do executes the request with the configured timeout
func (c HTTPClientReverseProxy) Do(req *http.Request) (*http.Response, error) {
	ctx, cancel := context.WithTimeout(req.Context(), c.Timeout())
	defer cancel()

	return c.client.Do(req.WithContext(ctx))
}

func (c HTTPClientReverseProxy) Timeout() time.Duration {
	return c.timeout
}
