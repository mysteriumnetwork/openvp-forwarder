/*
 * Copyright (C) 2024 The "MysteriumNetwork/openvpn-forwarder" Authors.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/mysteriumnetwork/openvpn-forwarder/proxy"
)

type service struct {
	proxyRequestDuration              *prometheus.HistogramVec
	proxyNumberOfLiveConnections      *prometheus.GaugeVec
	proxyNumberOfProcessedConnections *prometheus.CounterVec
}

// NewMetricsService creates instance of metrics service.
func NewMetricsService() (*service, error) {
	proxyRequestDuration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "proxy_request_duration",
		Help: "Proxy request duration in seconds",
	}, []string{"request_type", "hostname"})

	if err := prometheus.Register(proxyRequestDuration); err != nil {
		return nil, err
	}

	proxyNumberOfLiveConnections := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "proxy_number_of_live_connections",
		Help: "Number of currently live connections",
	}, []string{"request_type", "hostname"})

	if err := prometheus.Register(proxyNumberOfLiveConnections); err != nil {
		return nil, err
	}

	proxyNumberOfProcessedConnections := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "proxy_number_of_processed_connections",
		Help: "Number of incoming connections which were successfully assigned and processed",
	}, []string{"request_type", "hostname"})

	if err := prometheus.Register(proxyNumberOfProcessedConnections); err != nil {
		return nil, err
	}

	return &service{
		proxyRequestDuration:              proxyRequestDuration,
		proxyNumberOfLiveConnections:      proxyNumberOfLiveConnections,
		proxyNumberOfProcessedConnections: proxyNumberOfProcessedConnections,
	}, nil
}

func (s *service) ProxyHandlerMiddleware(next func(c *proxy.Context)) func(c *proxy.Context) {
	return func(c *proxy.Context) {
		startTime := time.Now()

		go func() {
			s.proxyNumberOfLiveConnections.With(prometheus.Labels{
				"request_type": c.RequestType(),
				"hostname":     c.WaitHostname(),
			}).Inc()
		}()

		next(c)

		s.proxyNumberOfLiveConnections.With(prometheus.Labels{
			"request_type": c.RequestType(),
			"hostname":     c.Hostname(),
		}).Dec()

		s.proxyRequestDuration.With(prometheus.Labels{
			"request_type": c.RequestType(),
			"hostname":     c.Hostname(),
		}).Observe(time.Since(startTime).Seconds())

		s.proxyNumberOfProcessedConnections.With(prometheus.Labels{
			"request_type": c.RequestType(),
			"hostname":     c.Hostname(),
		}).Inc()
	}
}
