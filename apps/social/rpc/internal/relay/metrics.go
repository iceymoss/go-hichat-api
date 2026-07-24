package relay

import "github.com/prometheus/client_golang/prometheus"

var (
	notificationPending  = prometheus.NewGauge(prometheus.GaugeOpts{Name: "social_notification_outbox_pending_total", Help: "Pending social notification outbox rows."})
	notificationDead     = prometheus.NewGauge(prometheus.GaugeOpts{Name: "social_notification_outbox_dead_total", Help: "Dead social notification outbox rows."})
	notificationLatency  = prometheus.NewHistogram(prometheus.HistogramOpts{Name: "social_notification_outbox_delivery_latency_seconds", Help: "Social notification delivery latency.", Buckets: prometheus.DefBuckets})
	notificationFailures = prometheus.NewCounterVec(prometheus.CounterOpts{Name: "social_notification_outbox_failures_total", Help: "Social notification relay failures."}, []string{"reason"})
)

func init() {
	prometheus.MustRegister(notificationPending, notificationDead, notificationLatency, notificationFailures)
}
