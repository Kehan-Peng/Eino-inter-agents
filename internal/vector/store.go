package vector

import "context"

// VectorRecord is the unified vector object used by Milvus-backed workflows.
type VectorRecord struct {
	ID       string
	Vector   []float32
	Metadata map[string]any
}

type SearchResult struct {
	ID       string
	Score    float64
	Metadata map[string]any
}

// VectorStore abstracts Milvus operations used by RAG retrieval and semantic dedup.
// Production implementation should use github.com/milvus-io/milvus-sdk-go/v2/client.
type VectorStore interface {
	Upsert(ctx context.Context, collection string, records []VectorRecord) error
	Search(ctx context.Context, collection string, vector []float32, topK int, filter string) ([]SearchResult, error)
}
