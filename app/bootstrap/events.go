package bootstrap // o donde agrupes las inicializaciones

import (
	"log"

	"mithrildb/db"
	"mithrildb/events"
	"mithrildb/expiration"
)

// InitEventSystem sets up the main event queue, fanout, and TTL listener.
func InitEventSystem(database *db.DB) {
	// Inicializa el CF para la cola principal de eventos
	eventCF, err := database.EnsureSystemColumnFamily(events.EventQueueCF)
	if err != nil {
		log.Fatalf("cannot init system.eventqueue column family: %v", err)
	}

	eventQueue, err := events.NewRocksQueue(database.TransactionDB, eventCF)
	if err != nil {
		log.Fatalf("failed to initialize event queue: %v", err)
	}

	events.InitEventQueue(eventQueue)

	// Inicializa y registra el listener de expiración (TTL)
	ttlCF, err := database.EnsureSystemColumnFamily("system.ttlqueue")
	if err != nil {
		log.Fatalf("cannot init system.ttlqueue: %v", err)
	}

	ttlQueue, err := events.NewRocksQueue(database.TransactionDB, ttlCF)
	if err != nil {
		log.Fatalf("failed to init TTL queue: %v", err)
	}

	events.RegisterListener(events.Listener{
		Name:      "expiration",
		Queue:     ttlQueue,
		PreFilter: expiration.ShouldProcessTTL,
	})

	events.StartFanout(eventQueue)

	// Inicializa el worker de expiración que consume system.ttlqueue
	expiration.NewListener(database, ttlQueue).Start()
}
