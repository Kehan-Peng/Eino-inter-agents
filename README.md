# Eino Internship Workflow Reconstruction

This repo reconstructs several internship workflows as Go files using Eino-style Workflow orchestration:

1. `VideoQualityReviewWorkflow`: RAG-based video quality pre-review.
2. `LiveHighlightWorkflow`: long live replay candidate segmentation, scoring, BGE+Milvus deduplication and Top-K rerank.
3. `MaterialCheckWorkflow`: Go backend material check plugin with ES recall, synonym normalization, material tree path matching and value-level comparison.

The code intentionally separates deterministic business nodes from LLM/model/vector-store clients. This matches Eino Workflow's design idea: node input/output is business-specific, while Workflow maps fields across nodes.

## Important

- This is a code skeleton for interview and architecture explanation.
- External systems are represented by interfaces: LLM, ES, ASR/OCR, Milvus, embedding model.
- Replace mock implementations with real SDK clients before production use.

