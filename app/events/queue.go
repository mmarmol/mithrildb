package events

type QueueReader interface {
	Next() (key []byte, value []byte, err error)
	Ack(key []byte) error
}
