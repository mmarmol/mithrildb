package db

import "errors"

var (
	ErrKeyAlreadyExists          = errors.New("key already exists")
	ErrKeyNotFound               = errors.New("key does not exist")
	ErrRevisionMismatch          = errors.New("revision mismatch")
	ErrInvalidColumnFamily       = errors.New("invalid column family")
	ErrInvalidUserColumnFamily   = errors.New("invalid user column family")
	ErrInvalidSystemColumnFamily = errors.New("invalid system column family")
	ErrEmptyKey                  = errors.New("key cannot be empty")
	ErrNilValue                  = errors.New("value cannot be nil")
	ErrFamilyExists              = errors.New("column family already exists")
	ErrInvalidListType           = errors.New("document is not a valid list")
	ErrEmptyList                 = errors.New("list is empty")
	ErrInvalidSetType            = errors.New("document is not a valid set")
)
