package db

import "errors"

var (
	ErrKeyAlreadyExists    = errors.New("key already exists")
	ErrKeyNotFound         = errors.New("key does not exist")
	ErrRevisionMismatch    = errors.New("revision mismatch")
	ErrInvalidColumnFamily = errors.New("invalid column family")
	ErrEmptyKey            = errors.New("key cannot be empty")
	ErrNilValue            = errors.New("value cannot be nil")
	ErrFamilyExists        = errors.New("column family already exists")
)
