package expiration

import (
	"fmt"
	"mithrildb/config"
	"time"
)

type Config struct {
	TickInterval time.Duration // How often the expiration process runs
	MaxPerCycle  int           // Max number of documents to process per cycle
	AutoScale    bool          // Whether to enable automatic scaling of MaxPerCycle

	MinPerCycle        int     // Minimum allowed value for MaxPerCycle when scaling down
	MaxPerCycleLimit   int     // Maximum allowed value for MaxPerCycle when scaling up
	ScaleStep          int     // Number of documents to increase/decrease when adjusting MaxPerCycle
	ScaleDownThreshold float64 // If deleted < (MaxPerCycle * ScaleDownThreshold), scale down
	ScaleUpFactor      float64 // If duration < (TickInterval * ScaleUpFactor), scale up when at limit
}

// BuildFromAppConfig builds an expiration.Config from config.AppConfig.
func BuildFromAppConfig(appCfg config.AppConfig) (Config, error) {
	raw := appCfg.Expiration

	duration, err := time.ParseDuration(raw.TickInterval)
	if err != nil {
		return Config{}, fmt.Errorf("invalid expiration.TickInterval: %w", err)
	}

	return Config{
		TickInterval:       duration,
		MaxPerCycle:        raw.MaxPerCycle,
		AutoScale:          raw.AutoScale,
		MinPerCycle:        raw.MinPerCycle,
		MaxPerCycleLimit:   raw.MaxPerCycleLimit,
		ScaleStep:          raw.ScaleStep,
		ScaleDownThreshold: raw.ScaleDownThreshold,
		ScaleUpFactor:      raw.ScaleUpFactor,
	}, nil
}
