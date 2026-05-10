package llm

import (
	"context"

	"eino-intern-workflows/internal/domain"
)

// ReviewJudge models the LLM quality decision node in the RAG pre-review workflow.
type ReviewJudge interface {
	JudgeVideoQuality(ctx context.Context, prompt domain.ReviewPrompt) (domain.ReviewDecision, error)
}

// HighlightJudge models the 7B LoRA high-light decision node.
type HighlightJudge interface {
	JudgeHighlight(ctx context.Context, seg domain.CandidateSegment) (domain.LoraHighlightScore, error)
}

// Embedder models bge-large-zh-v1.5 or any embedding model.
type Embedder interface {
	EmbedText(ctx context.Context, text string) ([]float32, error)
	Dim() int
}
