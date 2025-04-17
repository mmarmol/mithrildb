package expiration

import "time"

type Config struct {
	TickInterval   time.Duration // Intervalo entre ejecuciones del ciclo
	MaxPerCycle    int           // Máximo de elementos a procesar por ciclo
	AutoScale      bool          // Habilita o desactiva el auto-escalado
	ScaleInterval  int           // Cada cuántos ciclos aplicar lógica de escalado
	MetricsEnabled bool          // Exponer métricas (si se integra con Prometheus o similar)
}
