//
// Copyright Strimzi authors.
// License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
//

// Package servers contains some servers implementations
package servers

import (
	"context"
	"net/http"
	"time"

	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/strimzi/strimzi-canary/internal/services"
)

// HttpServer exposes some services over HTTP (i.e. Prometheus metrics, healthchecks)
type HttpServer struct {
	httpServer *http.Server
}

// NewHttpServer returns an instance of the HttpServer
func NewHttpServer(statusService services.StatusService) *HttpServer {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.Handle("/liveness", services.LivenessHandler())
	mux.Handle("/readiness", services.ReadinessHandler())
	mux.Handle("/status", statusService.StatusHandler())
	ms := HttpServer{}
	ms.httpServer = &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	return &ms
}

// Start runs the HTTP server in its own go routine
func (ms *HttpServer) Start() {
	glog.Infof("Starting HTTP server")
	go func() {
		ms.httpServer.ListenAndServe()
	}()
}

// Stop stops the HTTP server exiting the go routine
func (ms *HttpServer) Stop() {
	glog.Infof("Stopping HTTP server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	ms.httpServer.Shutdown(ctx)

	glog.Infof("HTTP server closed")
}
