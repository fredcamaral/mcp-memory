package storage

import (
	"context"
	"fmt"
	"mcp-memory/internal/circuitbreaker"
	"mcp-memory/pkg/types"
	"time"
)

// CircuitBreakerVectorStore wraps a VectorStore with circuit breaker protection
type CircuitBreakerVectorStore struct {
	store VectorStore
	cb    *circuitbreaker.CircuitBreaker
}

// NewCircuitBreakerVectorStore creates a new circuit breaker wrapped store
func NewCircuitBreakerVectorStore(store VectorStore, config *circuitbreaker.Config) *CircuitBreakerVectorStore {
	if config == nil {
		config = &circuitbreaker.Config{
			FailureThreshold:      5,
			SuccessThreshold:      2,
			Timeout:               30 * time.Second,
			MaxConcurrentRequests: 3,
			OnStateChange: func(from, to circuitbreaker.State) {
				// Log state changes
				fmt.Printf("VectorStore circuit breaker: %s -> %s\n", from, to)
			},
		}
	}

	return &CircuitBreakerVectorStore{
		store: store,
		cb:    circuitbreaker.New(config),
	}
}

// Initialize initializes the store
func (s *CircuitBreakerVectorStore) Initialize(ctx context.Context) error {
	return s.cb.Execute(ctx, func(ctx context.Context) error {
		return s.store.Initialize(ctx)
	})
}

// Store stores a chunk
func (s *CircuitBreakerVectorStore) Store(ctx context.Context, chunk types.ConversationChunk) error {
	return s.cb.Execute(ctx, func(ctx context.Context) error {
		return s.store.Store(ctx, chunk)
	})
}

// Search performs a search with fallback to empty results
func (s *CircuitBreakerVectorStore) Search(ctx context.Context, query types.MemoryQuery, embeddings []float64) (*types.SearchResults, error) {
	var result *types.SearchResults

	err := s.cb.ExecuteWithFallback(ctx,
		func(ctx context.Context) error {
			var err error
			result, err = s.store.Search(ctx, query, embeddings)
			return err
		},
		func(ctx context.Context, cbErr error) error {
			// Return empty results on circuit breaker failure
			result = &types.SearchResults{
				Results: []types.SearchResult{},
				Total:   0,
			}
			return nil
		},
	)

	return result, err
}

// GetByID gets a chunk by ID
func (s *CircuitBreakerVectorStore) GetByID(ctx context.Context, id string) (*types.ConversationChunk, error) {
	var result *types.ConversationChunk

	err := s.cb.Execute(ctx, func(ctx context.Context) error {
		var err error
		result, err = s.store.GetByID(ctx, id)
		return err
	})

	return result, err
}

// ListByRepository lists chunks by repository
func (s *CircuitBreakerVectorStore) ListByRepository(ctx context.Context, repository string, limit int, offset int) ([]types.ConversationChunk, error) {
	var result []types.ConversationChunk

	err := s.cb.ExecuteWithFallback(ctx,
		func(ctx context.Context) error {
			var err error
			result, err = s.store.ListByRepository(ctx, repository, limit, offset)
			return err
		},
		func(ctx context.Context, cbErr error) error {
			// Return empty list on circuit breaker failure
			result = []types.ConversationChunk{}
			return nil
		},
	)

	return result, err
}

// ListBySession lists chunks by session ID
func (s *CircuitBreakerVectorStore) ListBySession(ctx context.Context, sessionID string) ([]types.ConversationChunk, error) {
	var result []types.ConversationChunk

	err := s.cb.ExecuteWithFallback(ctx,
		func(ctx context.Context) error {
			var err error
			result, err = s.store.ListBySession(ctx, sessionID)
			return err
		},
		func(ctx context.Context, cbErr error) error {
			// Return empty list on circuit breaker failure
			result = []types.ConversationChunk{}
			return nil
		},
	)

	return result, err
}

// Delete deletes a chunk
func (s *CircuitBreakerVectorStore) Delete(ctx context.Context, id string) error {
	return s.cb.Execute(ctx, func(ctx context.Context) error {
		return s.store.Delete(ctx, id)
	})
}

// Update updates a chunk
func (s *CircuitBreakerVectorStore) Update(ctx context.Context, chunk types.ConversationChunk) error {
	return s.cb.Execute(ctx, func(ctx context.Context) error {
		return s.store.Update(ctx, chunk)
	})
}

// HealthCheck performs a health check
func (s *CircuitBreakerVectorStore) HealthCheck(ctx context.Context) error {
	return s.cb.Execute(ctx, func(ctx context.Context) error {
		return s.store.HealthCheck(ctx)
	})
}

// GetStats gets store statistics with fallback
func (s *CircuitBreakerVectorStore) GetStats(ctx context.Context) (*StoreStats, error) {
	var result *StoreStats

	err := s.cb.ExecuteWithFallback(ctx,
		func(ctx context.Context) error {
			var err error
			result, err = s.store.GetStats(ctx)
			return err
		},
		func(ctx context.Context, cbErr error) error {
			// Return empty stats on circuit breaker failure
			result = &StoreStats{
				TotalChunks:  0,
				ChunksByType: make(map[string]int64),
				ChunksByRepo: make(map[string]int64),
			}
			return nil
		},
	)

	return result, err
}

// Cleanup performs cleanup
func (s *CircuitBreakerVectorStore) Cleanup(ctx context.Context, retentionDays int) (int, error) {
	var result int

	err := s.cb.Execute(ctx, func(ctx context.Context) error {
		var err error
		result, err = s.store.Cleanup(ctx, retentionDays)
		return err
	})

	return result, err
}

// Close closes the store
func (s *CircuitBreakerVectorStore) Close() error {
	// Don't use circuit breaker for close operations
	return s.store.Close()
}

// GetCircuitBreakerStats returns circuit breaker statistics
func (s *CircuitBreakerVectorStore) GetCircuitBreakerStats() circuitbreaker.Stats {
	return s.cb.GetStats()
}

// Additional methods for service compatibility (with circuit breaker)

// GetAllChunks gets all chunks with circuit breaker protection
func (s *CircuitBreakerVectorStore) GetAllChunks(ctx context.Context) ([]types.ConversationChunk, error) {
	var result []types.ConversationChunk

	err := s.cb.ExecuteWithFallback(ctx,
		func(ctx context.Context) error {
			var err error
			result, err = s.store.GetAllChunks(ctx)
			return err
		},
		func(ctx context.Context, cbErr error) error {
			// Return empty list on circuit breaker failure
			result = []types.ConversationChunk{}
			return nil
		},
	)

	return result, err
}

// DeleteCollection deletes collection with circuit breaker protection
func (s *CircuitBreakerVectorStore) DeleteCollection(ctx context.Context, collection string) error {
	return s.cb.Execute(ctx, func(ctx context.Context) error {
		return s.store.DeleteCollection(ctx, collection)
	})
}

// ListCollections lists collections with circuit breaker protection
func (s *CircuitBreakerVectorStore) ListCollections(ctx context.Context) ([]string, error) {
	var result []string

	err := s.cb.ExecuteWithFallback(ctx,
		func(ctx context.Context) error {
			var err error
			result, err = s.store.ListCollections(ctx)
			return err
		},
		func(ctx context.Context, cbErr error) error {
			// Return empty list on circuit breaker failure
			result = []string{}
			return nil
		},
	)

	return result, err
}

// FindSimilar finds similar chunks with circuit breaker protection
func (s *CircuitBreakerVectorStore) FindSimilar(ctx context.Context, content string, chunkType *types.ChunkType, limit int) ([]types.ConversationChunk, error) {
	var result []types.ConversationChunk

	err := s.cb.ExecuteWithFallback(ctx,
		func(ctx context.Context) error {
			var err error
			result, err = s.store.FindSimilar(ctx, content, chunkType, limit)
			return err
		},
		func(ctx context.Context, cbErr error) error {
			// Return empty list on circuit breaker failure
			result = []types.ConversationChunk{}
			return nil
		},
	)

	return result, err
}

// StoreChunk stores chunk with circuit breaker protection
func (s *CircuitBreakerVectorStore) StoreChunk(ctx context.Context, chunk types.ConversationChunk) error {
	return s.cb.Execute(ctx, func(ctx context.Context) error {
		return s.store.StoreChunk(ctx, chunk)
	})
}

// BatchStore stores chunks in batch with circuit breaker protection
func (s *CircuitBreakerVectorStore) BatchStore(ctx context.Context, chunks []types.ConversationChunk) (*BatchResult, error) {
	var result *BatchResult

	err := s.cb.ExecuteWithFallback(ctx,
		func(ctx context.Context) error {
			var err error
			result, err = s.store.BatchStore(ctx, chunks)
			return err
		},
		func(ctx context.Context, cbErr error) error {
			// Return failed result on circuit breaker failure
			result = &BatchResult{
				Success:      0,
				Failed:       len(chunks),
				Errors:       []string{"circuit breaker open"},
				ProcessedIDs: []string{},
			}
			return nil
		},
	)

	return result, err
}

// BatchDelete deletes chunks in batch with circuit breaker protection
func (s *CircuitBreakerVectorStore) BatchDelete(ctx context.Context, ids []string) (*BatchResult, error) {
	var result *BatchResult

	err := s.cb.ExecuteWithFallback(ctx,
		func(ctx context.Context) error {
			var err error
			result, err = s.store.BatchDelete(ctx, ids)
			return err
		},
		func(ctx context.Context, cbErr error) error {
			// Return failed result on circuit breaker failure
			result = &BatchResult{
				Success:      0,
				Failed:       len(ids),
				Errors:       []string{"circuit breaker open"},
				ProcessedIDs: ids,
			}
			return nil
		},
	)

	return result, err
}
