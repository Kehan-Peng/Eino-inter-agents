package vector

import (
	"context"
	"math"
	"sort"
	"sync"
)

// MemoryStore is useful for local tests. Replace it with MilvusStore in production.
type MemoryStore struct {
	mu          sync.RWMutex
	collections map[string][]VectorRecord
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{collections: make(map[string][]VectorRecord)}
}

func (s *MemoryStore) Upsert(ctx context.Context, collection string, records []VectorRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.collections[collection] = append(s.collections[collection], records...)
	return nil
}

func (s *MemoryStore) Search(ctx context.Context, collection string, vector []float32, topK int, filter string) ([]SearchResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []SearchResult
	for _, r := range s.collections[collection] {
		out = append(out, SearchResult{ID: r.ID, Score: cosine(vector, r.Vector), Metadata: r.Metadata})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Score > out[j].Score })
	if topK > 0 && len(out) > topK {
		out = out[:topK]
	}
	return out, nil
}

func cosine(a, b []float32) float64 {
	if len(a) == 0 || len(a) != len(b) {
		return 0
	}
	var dot, na, nb float64
	for i := range a {
		x, y := float64(a[i]), float64(b[i])
		dot += x * y
		na += x * x
		nb += y * y
	}
	if na == 0 || nb == 0 {
		return 0
	}
	return dot / (math.Sqrt(na) * math.Sqrt(nb))
}
