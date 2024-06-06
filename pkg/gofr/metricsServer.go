package gofr

import (
	"fmt"
	"net/http"
	"time"

	"github.com/peter-stratton/gofr/pkg/gofr/container"
	"github.com/peter-stratton/gofr/pkg/gofr/metrics"
)

type metricServer struct {
	port int
}

func newMetricServer(port int) *metricServer {
	return &metricServer{port: port}
}

func (m *metricServer) Run(c *container.Container) {
	var srv *http.Server

	if m != nil {
		c.Logf("Starting metrics server on port: %d", m.port)

		srv = &http.Server{
			Addr:              fmt.Sprintf(":%d", m.port),
			Handler:           metrics.GetHandler(c.Metrics()),
			ReadHeaderTimeout: 5 * time.Second,
		}

		c.Error(srv.ListenAndServe())
	}
}
