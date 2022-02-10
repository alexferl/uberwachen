package storage

import "context"

// Storage a common interface for all database storage
type Storage interface {
	Init(ctx context.Context) error
	Get(ctx context.Context, name string, item interface{}) error
	GetAll(ctx context.Context, items interface{}) error
	Set(ctx context.Context, data interface{}) error
	Delete(ctx context.Context, name string) error
	Update(ctx context.Context, name string, data interface{}) error
}
