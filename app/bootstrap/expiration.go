package bootstrap

import (
	"log"

	"mithrildb/config"
	"mithrildb/db"
	"mithrildb/expiration"
)

// InitExpirationService configura y arranca el job de expiración periódica.
func InitExpirationService(database *db.DB, cfg config.AppConfig) *expiration.Service {
	expCfg, err := expiration.BuildFromAppConfig(cfg)
	if err != nil {
		log.Fatalf("invalid expiration config: %v", err)
	}

	service := expiration.NewService(database, expCfg)
	service.Start()
	return service
}
