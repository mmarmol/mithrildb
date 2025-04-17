package expiration

import "time"

type Stats struct {
	LastRunTime  time.Time     // Cuándo fue la última ejecución
	LastDuration time.Duration // Cuánto tardó el último ciclo
	LastDeleted  int           // Cuántos documentos expiraron en el último ciclo
	LastError    error         // Último error del ciclo (si hubo)

	TotalRuns    int // Total de ciclos ejecutados
	TotalDeleted int // Total de documentos expirados en todos los ciclos
}
