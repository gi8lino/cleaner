package flags

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/gi8lino/cleaner/internal/logging"
	flag "github.com/spf13/pflag"
)

// Config holds all runtime parameters.
type Config struct {
	Debug       bool
	LogFormat   logging.LogFormat
	RootDir     string
	Age         time.Duration
	PushGateway string
	Job         string
	SkipNames   []string
	Labels      map[string]string // arbitrary grouping labels for Pushgateway
}

// HelpRequested is returned when the user requests --help or --version.
type HelpRequested struct {
	Message string // Printed message
}

// Error returns the message to print for HelpRequested.
func (e *HelpRequested) Error() string {
	return e.Message
}

// ParseFlags binds pflag flags into a Config.
// Uses Duration for age, StringSlice for skip & labels.
func ParseFlags(version string, args []string, out io.Writer) (*Config, error) {
	var (
		debug     = false
		root      = ""
		age       = 7 * 24 * time.Hour
		pushURL   = "http://prometheus-pushgateway.monitoring.svc.cluster.local:9091"
		job       = "cleaner"
		skip      = []string{}
		labelArgs []string
	)

	fs := flag.NewFlagSet("cleaner", flag.ContinueOnError)
	fs.SortFlags = false
	fs.SetOutput(out)

	// Logging flags
	fs.BoolVar(&debug, "debug", debug, "Enable debug logging")
	logFormat := fs.StringP("log-format", "l", "text", "Log format: text or json")

	// Core cleanup flags
	fs.StringVarP(&root, "root", "r", root, "root directory to clean")
	fs.DurationVarP(&age, "older-than", "o", age, "minimum age (duration) of empty dirs to delete")
	fs.StringVarP(&pushURL, "pushgateway-url", "u", pushURL, "Prometheus Pushgateway URL")
	fs.StringVarP(&job, "job-name", "j", job, "metric namespace / job name")
	fs.StringSliceVar(&skip, "skip", skip, "directories to skip (can repeat)")
	fs.StringSliceVar(&labelArgs, "label", nil, "extra labels for Pushgateway in key=value form")

	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: %s [flags]\n\nFlags:\n", strings.ToLower(fs.Name())) // nolint:errcheck
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	// coerce log-format into our enum
	if *logFormat != "json" && *logFormat != "text" {
		return nil, fmt.Errorf("invalid log format: '%s'", *logFormat)
	}
	lf := logging.LogFormat(*logFormat)

	// build label map (we'll validate key=value in Validate)
	labels := map[string]string{"job": job}
	for _, kv := range labelArgs {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid label %q: must be key=value", kv)
		}
		labels[parts[0]] = parts[1]
	}

	return &Config{
		Debug:       debug,
		LogFormat:   lf,
		RootDir:     root,
		Age:         age,
		PushGateway: pushURL,
		Job:         job,
		SkipNames:   skip,
		Labels:      labels,
	}, nil
}

// Validate checks that the Config fields satisfy all requirements.
func (c *Config) Validate() error {
	if c.RootDir == "" {
		return errors.New("flag --root is required and cannot be empty")
	}
	if c.Age <= 0 {
		return errors.New("flag --older-than must be > 0")
	}
	switch c.LogFormat {
	case logging.LogFormatText, logging.LogFormatJSON:
		// ok
	default:
		return fmt.Errorf("invalid log format: %q (allowed: text, json)", c.LogFormat)
	}
	if c.Job == "" {
		return errors.New("flag --job-name cannot be empty")
	}
	for k, v := range c.Labels {
		if strings.TrimSpace(k) == "" || strings.TrimSpace(v) == "" {
			return fmt.Errorf("invalid label %q=%q: neither key nor value can be empty", k, v)
		}
	}
	return nil
}
