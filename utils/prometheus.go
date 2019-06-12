package utils

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func HandleSummary(summary *prometheus.SummaryVec, f func() error) error {
	start := time.Now()
	err := f()
	took := time.Now().Sub(start)

	if err != nil {
		summary.WithLabelValues(err.Error()).Observe(float64(took) / float64(time.Second))
	} else {
		summary.WithLabelValues("").Observe(float64(took) / float64(time.Second))
	}

	return err
}
