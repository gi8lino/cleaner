package metrics

import (
	"context"
	"fmt"

	"github.com/gi8lino/cleaner/internal/flags"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

// PushMetrics pushes two gauges to the Prometheus Pushgateway:
func PushMetrics(ctx context.Context, cfg *flags.Config, deletedCount int64, failed bool) error {
	// Gauge for number of deleted directories in this run
	deletedDirsGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: fmt.Sprintf("cleaner_%s_deleted_dirs", cfg.Job),
		Help: "Number of empty directories deleted by this run",
	})

	// Gauge to indicate if this run failed (1 = failed, 0 = success)
	failedGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: fmt.Sprintf("cleaner_%s_failed", cfg.Job),
		Help: "Whether the last run failed (1 = failed, 0 = success)",
	})

	// Set gauge values
	deletedDirsGauge.Set(float64(deletedCount))
	failedGauge.Set(boolToFloat(failed))

	// Build Pushgateway pusher
	pusher := push.New(cfg.PushGateway, cfg.Job).
		Collector(deletedDirsGauge).
		Collector(failedGauge)

	// Add grouping labels
	for key, val := range cfg.Labels {
		pusher = pusher.Grouping(key, val)
	}

	// Push to Pushgateway (honors ctx for cancellation/timeouts)
	return pusher.AddContext(ctx)
}

// boolToFloat converts a bool to a float64
// True = 0, False = 1
func boolToFloat(b bool) float64 {
	if !b {
		return 1
	}

	return 0
}
