package metrics

import (
	"errors"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	log "github.com/sirupsen/logrus"
)

const namespace = "library_service"

func init() {
	goCollector := collectors.NewGoCollector(
		collectors.WithGoCollectorRuntimeMetrics(
			collectors.MetricsAll,
		),
	)

	processCollector := collectors.NewProcessCollector(
		collectors.ProcessCollectorOpts{},
	)

	if err := prometheus.Register(goCollector); err != nil {
		var alreadyRegisteredError prometheus.AlreadyRegisteredError
		if !errors.As(err, &alreadyRegisteredError) {
			log.Printf("failed to register GoCollector: %v", err)
		}
	}

	if err := prometheus.Register(processCollector); err != nil {
		var alreadyRegisteredError prometheus.AlreadyRegisteredError
		if !errors.As(err, &alreadyRegisteredError) {
			log.Printf("failed to register ProcessCollector: %v", err)
		}
	}
}
