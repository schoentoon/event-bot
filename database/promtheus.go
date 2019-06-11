package database

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

type DBCollector struct {
	db                                                        *sql.DB
	openConns, inUseConns, idleConns, waitCount, waitDuration *prometheus.Desc
}

func NewCollector(db *sql.DB) *DBCollector {
	return &DBCollector{
		db: db,
		openConns: prometheus.NewDesc("db_open_conns",
			"Amount of open connections",
			nil, nil,
		),
		inUseConns: prometheus.NewDesc("db_in_use",
			"Amount of connections in use",
			nil, nil,
		),
		idleConns: prometheus.NewDesc("db_idle_conns",
			"Amount of idle connections",
			nil, nil,
		),
		waitCount: prometheus.NewDesc("db_wait_count",
			"The total number of connections waited for",
			nil, nil,
		),
		waitDuration: prometheus.NewDesc("db_wait_duration",
			"The total time blocked waiting for a new connection",
			nil, nil,
		),
	}
}

func (collector *DBCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.openConns
	ch <- collector.inUseConns
	ch <- collector.idleConns
	ch <- collector.waitCount
	ch <- collector.waitDuration
}

func (collector *DBCollector) Collect(ch chan<- prometheus.Metric) {
	stats := collector.db.Stats()

	ch <- prometheus.MustNewConstMetric(collector.openConns, prometheus.GaugeValue, float64(stats.OpenConnections))
	ch <- prometheus.MustNewConstMetric(collector.inUseConns, prometheus.GaugeValue, float64(stats.InUse))
	ch <- prometheus.MustNewConstMetric(collector.idleConns, prometheus.GaugeValue, float64(stats.Idle))
	ch <- prometheus.MustNewConstMetric(collector.waitCount, prometheus.GaugeValue, float64(stats.WaitCount))
	ch <- prometheus.MustNewConstMetric(collector.waitDuration, prometheus.GaugeValue, float64(stats.WaitDuration))
}
