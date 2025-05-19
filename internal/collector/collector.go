package collector

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/gi8lino/cleaner/internal/flags"
)

// collectDeleted walks & deletes empty dirs older than cfg.Age.
// It never buffers large lists: WalkDir streams entries and Readdirnames(1) tests emptiness.
func CollectDeleted(ctx context.Context, logger *slog.Logger, cfg *flags.Config) (int64, error) {
	threshold := time.Now().Add(-cfg.Age)
	skipSet := make(map[string]struct{}, len(cfg.SkipNames))
	for _, n := range cfg.SkipNames {
		skipSet[n] = struct{}{}
	}

	var deleted int64
	err := filepath.WalkDir(cfg.RootDir, func(path string, d fs.DirEntry, err error) error {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err != nil {
			// can't stat or open—log and keep going
			logger.Error(fmt.Sprintf("skip %q", path), "error", err.Error())
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		if _, skip := skipSet[d.Name()]; skip {
			return fs.SkipDir
		}
		info, err := d.Info()
		if err != nil {
			logger.Error(fmt.Sprintf("stat %q", path), "error", err.Error())
			return fs.SkipDir
		}
		if info.ModTime().After(threshold) {
			return nil // too recent
		}
		// open & read one entry to test emptiness
		f, err := os.Open(path)
		if err != nil {
			logger.Error(fmt.Sprintf("open %q", path), "error", err.Error())
			return fs.SkipDir
		}
		names, err := f.Readdirnames(1)
		f.Close() // nolint:errcheck
		if err != nil && err != io.EOF {
			logger.Error(fmt.Sprintf("read %q", path), "error", err.Error())
			return fs.SkipDir
		}
		if len(names) == 0 {
			if removeErr := os.Remove(path); removeErr != nil {
				logger.Error(fmt.Sprintf("remove %q", path), "error", removeErr)
			} else {
				atomic.AddInt64(&deleted, 1)
			}
			return fs.SkipDir
		}
		return nil
	})

	return deleted, err
}
