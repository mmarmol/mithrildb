package expiration

import (
	"log"
	"mithrildb/db"
	"time"
)

type Service struct {
	db     *db.DB
	config Config
	stats  Stats
}

func (s *Service) Stats() Stats {
	return s.stats
}

func NewService(db *db.DB, cfg Config) *Service {
	return &Service{
		db:     db,
		config: cfg,
	}
}

func (s *Service) Start() {
	go func() {
		ticker := time.NewTicker(s.config.TickInterval)
		defer ticker.Stop()

		for {
			start := time.Now()
			count, err := s.ProcessCycle()
			duration := time.Since(start)

			s.stats.LastRunTime = start
			s.stats.LastDuration = duration
			s.stats.LastDeleted = count
			s.stats.TotalRuns++
			s.stats.TotalDeleted += count
			s.stats.LastError = err

			s.adaptLimit(count, duration)

			if err != nil {
				log.Printf("[expiration] ❌ error: %v", err)
			} else {
				log.Printf("[expiration] ✅ deleted=%d in %s (limit=%d)", count, duration, s.config.MaxPerCycle)
			}

			<-ticker.C
		}
	}()
}

func (s *Service) ProcessCycle() (int, error) {
	now := time.Now().Unix()

	count, err := s.db.ProcessExpiredBatch(now, s.config.MaxPerCycle)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (s *Service) adaptLimit(deleted int, duration time.Duration) {
	cfg := s.config

	if !cfg.AutoScale {
		return // Autoscaling is disabled
	}

	// 📈 Scale up:
	// If we hit the current max and the cycle ran fast enough, increase MaxPerCycle
	if deleted == cfg.MaxPerCycle && duration < time.Duration(float64(cfg.TickInterval)*cfg.ScaleUpFactor) {
		s.config.MaxPerCycle += cfg.ScaleStep
		if s.config.MaxPerCycle > cfg.MaxPerCycleLimit {
			s.config.MaxPerCycle = cfg.MaxPerCycleLimit
		}
		return
	}

	// 📉 Scale down:
	// If few documents were deleted and we’re above the minimum threshold, reduce MaxPerCycle
	minDeleted := int(float64(cfg.MaxPerCycle) * cfg.ScaleDownThreshold)
	if deleted > 0 && deleted < minDeleted && s.config.MaxPerCycle > cfg.MinPerCycle {
		s.config.MaxPerCycle -= cfg.ScaleStep
		if s.config.MaxPerCycle < cfg.MinPerCycle {
			s.config.MaxPerCycle = cfg.MinPerCycle
		}
	}
}
