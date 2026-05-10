package clients

import (
	"context"

	"eino-intern-workflows/internal/domain"
)

// RuleRetriever is the RAG retrieval boundary. A production implementation can
// combine keyword retrieval, vector retrieval and reranking.
type RuleRetriever interface {
	RetrieveRules(ctx context.Context, query domain.ReviewQuery, topK int) ([]domain.RuleChunk, error)
}

// MaterialSearchClient recalls candidate standard material words from ES keyword fields.
type MaterialSearchClient interface {
	RecallMaterialCandidates(ctx context.Context, rawText string, category string, topK int) ([]domain.MaterialCandidate, error)
}
