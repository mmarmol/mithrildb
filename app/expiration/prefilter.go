package expiration

import "mithrildb/events"

func ShouldProcessTTL(event events.DocumentChangeEvent) bool {
	// Always process deletes to clean up TTL index
	if event.Operation == events.OpDelete {
		return true
	}

	// Only act on modifications if explicit expiration was provided and is positive
	return event.ExplicitExpiration != nil && *event.ExplicitExpiration > 0
}
