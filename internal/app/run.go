package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/signal"
	"syscall"

	"github.com/gi8lino/cleaner/internal/collector"
	"github.com/gi8lino/cleaner/internal/flags"
	"github.com/gi8lino/cleaner/internal/logging"
	"github.com/gi8lino/cleaner/internal/metrics"
)

func Run(ctx context.Context, version, gitCommit string, args []string, out io.Writer) error {
	// Parse flags
	cfg, err := flags.ParseFlags(version, args, out)
	if err != nil {
		var helpErr *flags.HelpRequested
		if errors.As(err, &helpErr) {
			fmt.Fprint(out, helpErr.Error()) // nolint:errcheck
			return nil
		}
		return fmt.Errorf("parsing error: %w", err)
	}

	// Setup logger
	logger := logging.SetupLogger(cfg.LogFormat, cfg.Debug, out)
	logger.Info("Starting collector",
		"version", version,
		"commit", gitCommit,
	)

	// Make a cancellable context on SIGINT/SIGTERM
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Guarantee metrics are pushed no matter how we exit
	var (
		deleted int64
		walkErr error
	)
	defer func() {
		// if the context was canceled, treat as a failure
		failed := walkErr != nil || ctx.Err() != nil
		if pushErr := metrics.PushMetrics(ctx, cfg, deleted, failed); pushErr != nil {
			logger.Error("pushgateway error", "error", pushErr)
		}
	}()

	// Short-circuit if we got canceled before starting
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Do the walk & delete
	deleted, walkErr = collector.CollectDeleted(ctx, logger, cfg)
	if walkErr != nil {
		return fmt.Errorf("cleanup failed: %w", walkErr)
	}

	logger.Info(fmt.Sprintf("Deleted %d empty directories older than %s", deleted, cfg.Age))
	return nil
}
