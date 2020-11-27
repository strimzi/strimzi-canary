//
// Copyright Strimzi authors.
// License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
//

// Package servers contains some servers implementations
package servers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsServer defines an HTTP server exposing the metrics in Prometheus format
type MetricsServer struct {
	httpServer *http.Server
}

// NewMetricsServer returns an instance of the MetricsServer
func NewMetricsServer() *MetricsServer {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	ms := MetricsServer{}
	ms.httpServer = &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	return &ms
}

// Start runs the HTTP server in its own go routine
func (ms *MetricsServer) Start() {
	log.Printf("Starting metrics server")
	go func() {
		ms.httpServer.ListenAndServe()
	}()
}

// Stop stops the HTTP server exiting the go routine
func (ms *MetricsServer) Stop() {
	log.Printf("Stopping metrics server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	ms.httpServer.Shutdown(ctx)

	log.Printf("Metrics server closed")
}
