package cache

import (
	"context"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/eko/gocache/lib/v4/store"

	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/marshaler"
	ristrettoStore "github.com/eko/gocache/store/ristretto/v4"
)

type Cache interface {
	Set(ctx context.Context, key string, value any) error
	Get(ctx context.Context, key string) (any, error)
}

func NewGoCache() (Cache, error) {
	ristrettoCache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1_000_000,
		MaxCost:     1 << 30,
		BufferItems: 64,
	})
	if err != nil {
		return nil, err
	}
	ristrettoCacheStore := ristrettoStore.NewRistretto(ristrettoCache, store.WithExpiration(1*time.Hour))

	return &goCache{
		marshaler: marshaler.New(cache.New[any](ristrettoCacheStore)),
	}, nil
}

type goCache struct {
	marshaler *marshaler.Marshaler
}

func (g goCache) Set(ctx context.Context, key string, value any) error {
	//TODO implement me
	panic("implement me")
}

func (g goCache) Get(ctx context.Context, key string) (any, error) {
	//TODO implement me
	panic("implement me")
}
