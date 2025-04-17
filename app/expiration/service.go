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
	// Escalar hacia arriba si estamos al tope y rápido
	if deleted == s.config.MaxPerCycle && duration < s.config.TickInterval/2 {
		s.config.MaxPerCycle += 100
		if s.config.MaxPerCycle > 5000 {
			s.config.MaxPerCycle = 5000
		}
		return
	}

	// Solo considerar bajar si efectivamente hubo actividad, pero poca
	if deleted > 0 && deleted < s.config.MaxPerCycle/4 && s.config.MaxPerCycle > 200 {
		s.config.MaxPerCycle -= 100
		if s.config.MaxPerCycle < 200 {
			s.config.MaxPerCycle = 200
		}
	}
}
