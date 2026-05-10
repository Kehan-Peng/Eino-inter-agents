package main

import (
	"context"
	"fmt"
	"hash/fnv"
	"strings"

	"eino-intern-workflows/internal/clients"
	"eino-intern-workflows/internal/domain"
	"eino-intern-workflows/internal/llm"
	"eino-intern-workflows/internal/nodes"
	"eino-intern-workflows/internal/vector"
	"eino-intern-workflows/internal/workflows"
)

func main() {
	ctx := context.Background()

	// 1. Video quality pre-review workflow.
	reviewWF, _ := workflows.BuildVideoQualityReviewWorkflow(ctx, nodes.VideoReviewNodes{
		Rules: mockRuleRetriever{},
		Judge: mockReviewJudge{},
		TopK:  5,
	})
	reviewOut, _ := reviewWF.Invoke(ctx, domain.VideoReviewInput{
		VideoID: "v_001",
		ASR:     []domain.ASRSegment{{Start: 0, End: 30, Text: "这件衣服材质很舒服，价格优惠，马上上链接"}},
		Product: domain.ProductInfo{ProductID: "p_1", Title: "休闲外套", Category: "休闲服"},
	})
	fmt.Printf("review output: %+v\n", reviewOut)

	// 2. Live highlight workflow with Milvus boundary represented by MemoryStore here.
	highlightWF, _ := workflows.BuildLiveHighlightWorkflow(ctx, nodes.LiveHighlightNodes{
		Judge:          mockHighlightJudge{},
		Embedder:       mockEmbedder{dim: 1024},
		VectorStore:    vector.NewMemoryStore(), // replace with vector.NewMilvusStore("host:19530") in prod
		Collection:     "live_highlight_segments",
		DedupThreshold: 0.88,
		TopK:           15,
	})
	highlightOut, _ := highlightWF.Invoke(ctx, domain.LiveHighlightInput{
		LiveID:   "live_001",
		Products: []domain.ProductInfo{{ProductID: "p_1", SKU: "sku_1", Title: "休闲外套", Category: "休闲服"}},
		ASR: []domain.ASRSegment{
			{Start: 0, End: 50, Text: "这件休闲服现在有促销，颜色好看，上身效果很好，马上下单"},
			{Start: 60, End: 100, Text: "这件休闲服现在有促销，颜色好看，上身效果很好，马上下单"},
		},
	})
	fmt.Printf("highlight output: top=%d\n", len(highlightOut.TopSegments))

	// 3. Material Check workflow.
	materialWF, _ := workflows.BuildMaterialCheckWorkflow(ctx, nodes.MaterialNodes{
		ES: mockMaterialES{},
		Root: &nodes.MaterialNode{Value: "材质", Children: []*nodes.MaterialNode{
			{Value: "皮革", Children: []*nodes.MaterialNode{
				{Value: "牛皮", Children: []*nodes.MaterialNode{{Value: "头层牛皮"}, {Value: "二层牛皮"}}},
			}},
		}},
		Synonyms: map[string]string{
			"二层牛皮革":      "二层牛皮",
			"二层牛皮（除牛反绒）": "二层牛皮",
			"除牛反绒":       "二层牛皮",
		},
		ValueSenseLevel: map[string]int64{"头层牛皮": 5, "二层牛皮": 4, "牛皮": 3},
		TopK:            5,
	})
	materialOut, _ := materialWF.Invoke(ctx, domain.MaterialCheckInput{LeftMaterial: "牛皮", RightMaterial: "头层牛皮", Category: "服饰"})
	fmt.Printf("material output: %+v\n", materialOut)
}

// -----------------------------
// Mock clients for local demo.
// -----------------------------

var _ clients.RuleRetriever = mockRuleRetriever{}
var _ llm.ReviewJudge = mockReviewJudge{}
var _ llm.HighlightJudge = mockHighlightJudge{}
var _ llm.Embedder = mockEmbedder{}
var _ clients.MaterialSearchClient = mockMaterialES{}

type mockRuleRetriever struct{}

func (mockRuleRetriever) RetrieveRules(ctx context.Context, query domain.ReviewQuery, topK int) ([]domain.RuleChunk, error) {
	return []domain.RuleChunk{{RuleID: "rule_display_clear", Category: query.Product.Category, ProblemType: "商品展示", Text: "商品主体需要清晰展示，不得被大面积遮挡。", Score: 0.91}}, nil
}

type mockReviewJudge struct{}

func (mockReviewJudge) JudgeVideoQuality(ctx context.Context, prompt domain.ReviewPrompt) (domain.ReviewDecision, error) {
	return domain.ReviewDecision{VideoID: prompt.VideoID, QualityLabels: []string{"商品展示正常"}, Score: 0.86, Reason: "商品讲解和展示信息较完整", RuleIDs: []string{"rule_display_clear"}, Confidence: 0.88}, nil
}

type mockHighlightJudge struct{}

func (mockHighlightJudge) JudgeHighlight(ctx context.Context, seg domain.CandidateSegment) (domain.LoraHighlightScore, error) {
	text := seg.ASRText
	sp := 0.5
	cv := 0.5
	if strings.Contains(text, "促销") || strings.Contains(text, "下单") || strings.Contains(text, "链接") {
		sp = 0.82
		cv = 0.76
	}
	return domain.LoraHighlightScore{SegmentID: seg.SegmentID, IsHighlight: sp > 0.7, QualityScore: 0.8, SellingPoint: sp, ConversionHint: cv, Reason: "卖点和促单信息集中"}, nil
}

type mockEmbedder struct{ dim int }

func (m mockEmbedder) Dim() int { return m.dim }
func (m mockEmbedder) EmbedText(ctx context.Context, text string) ([]float32, error) {
	vec := make([]float32, m.dim)
	for _, token := range strings.Fields(text) {
		h := fnv.New32a()
		_, _ = h.Write([]byte(token))
		vec[int(h.Sum32())%m.dim] += 1
	}
	return vec, nil
}

type mockMaterialES struct{}

func (mockMaterialES) RecallMaterialCandidates(ctx context.Context, rawText string, category string, topK int) ([]domain.MaterialCandidate, error) {
	std := strings.TrimSpace(rawText)
	return []domain.MaterialCandidate{{RawText: rawText, StandardName: std, Score: 1.0}}, nil
}
