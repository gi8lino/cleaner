package metrics

import (
	"context"
	"fmt"

	"github.com/gi8lino/cleaner/internal/flags"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

// PushMetrics registers two Gauges and pushes them via the prometheus/push package.
func PushMetrics(ctx context.Context, cfg *flags.Config, deleted int64, failed bool) error {
	// create gauges
	Failed := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: fmt.Sprintf("%s_cronjob_failed", cfg.Job),
		Help: "Whether the last cronjob run failed (1 == failed)",
	})
	Deleted := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: fmt.Sprintf("%s_deleted_dirs", cfg.Job),
		Help: "Number of empty directories deleted",
	})
	// set values
	if failed {
		Failed.Set(1)
	} else {
		Failed.Set(0)
	}
	Deleted.Set(float64(deleted))

	// prepare pusher
	p := push.New(cfg.PushGateway, cfg.Job).
		Collector(Failed).
		Collector(Deleted)
	// add all grouping labels
	for k, v := range cfg.Labels {
		p = p.Grouping(k, v)
	}

	return p.AddContext(ctx)
}
