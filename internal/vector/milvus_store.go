package vector

import (
	"context"
	"fmt"
)

// MilvusStore is a thin adapter boundary for Milvus.
//
// Keep the workflow independent from the concrete Milvus SDK because Milvus SDK
// types tend to leak collection schema and query options into business code.
// Replace the TODO sections with github.com/milvus-io/milvus-sdk-go/v2/client calls:
//   - client.NewGrpcClient(ctx, addr)
//   - entity.NewColumnFloatVector(vectorField, dim, vectors)
//   - client.Insert / client.Upsert
//   - client.Search with metric type COSINE or IP if vectors are normalized
//   - IVF_FLAT index for the live-highlight candidate scale in this project.
type MilvusStore struct {
	Addr        string
	VectorField string
	MetricType  string // COSINE / IP / L2
}

func NewMilvusStore(addr string) *MilvusStore {
	return &MilvusStore{
		Addr:        addr,
		VectorField: "embedding",
		MetricType:  "COSINE",
	}
}

func (s *MilvusStore) Upsert(ctx context.Context, collection string, records []VectorRecord) error {
	if collection == "" {
		return fmt.Errorf("milvus collection is empty")
	}
	if len(records) == 0 {
		return nil
	}
	// TODO: call Milvus SDK Insert/Upsert here.
	// The workflow deliberately depends on VectorStore, so this method can be
	// replaced by a real SDK implementation without touching workflow nodes.
	return nil
}

func (s *MilvusStore) Search(ctx context.Context, collection string, vector []float32, topK int, filter string) ([]SearchResult, error) {
	if collection == "" {
		return nil, fmt.Errorf("milvus collection is empty")
	}
	if len(vector) == 0 {
		return nil, fmt.Errorf("query vector is empty")
	}
	if topK <= 0 {
		topK = 10
	}
	// TODO: call Milvus SDK Search here.
	// Return scores as cosine similarity in [0,1] when vectors are normalized.
	return []SearchResult{}, nil
}
