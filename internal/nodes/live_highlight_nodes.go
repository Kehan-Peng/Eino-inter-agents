package nodes

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"eino-intern-workflows/internal/domain"
	"eino-intern-workflows/internal/llm"
	"eino-intern-workflows/internal/vector"
)

type LiveHighlightNodes struct {
	Judge          llm.HighlightJudge
	Embedder       llm.Embedder
	VectorStore    vector.VectorStore
	Collection     string
	DedupThreshold float64
	TopK           int
}

func (n LiveHighlightNodes) SegmentCandidates(ctx context.Context, in domain.LiveHighlightInput) (domain.CandidateSet, error) {
	// Interview reconstruction:
	// product keyword + ASR timestamp + cart switch signal -> 30-120s segments.
	// Here we use ASR blocks as a deterministic placeholder for candidate generation.
	var out []domain.CandidateSegment
	if len(in.Products) == 0 {
		return domain.CandidateSet{LiveID: in.LiveID}, nil
	}
	for i, seg := range in.ASR {
		product := in.Products[i%len(in.Products)]
		start, end := seg.Start, seg.End
		if end-start < 30 {
			end = start + 30
		}
		if end-start > 120 {
			end = start + 120
		}
		out = append(out, domain.CandidateSegment{
			SegmentID: fmt.Sprintf("%s_seg_%04d", in.LiveID, i),
			LiveID:    in.LiveID,
			Product:   product,
			Start:     start,
			End:       end,
			Duration:  end - start,
			ASRText:   strings.TrimSpace(seg.Text),
		})
	}
	return domain.CandidateSet{LiveID: in.LiveID, Candidates: out}, nil
}

func (n LiveHighlightNodes) VisualQualityScore(ctx context.Context, in domain.CandidateSet) (domain.ScoredSegmentSet, error) {
	segments := make([]domain.ScoredSegment, 0, len(in.Candidates))
	for _, c := range in.Candidates {
		vf := domain.VisualFeature{
			SegmentID:       c.SegmentID,
			Sharpness:       0.70,
			Stability:       0.75,
			BlackScreen:     1.00,
			Occlusion:       0.85,
			CenterScore:     0.72,
			BackgroundClean: 0.70,
			Lighting:        0.78,
			ProductDisplay:  0.80,
		}
		vf.VisualScore = 0.12*vf.Sharpness + 0.12*vf.Stability + 0.08*vf.BlackScreen + 0.08*vf.Occlusion + 0.12*vf.CenterScore + 0.12*vf.BackgroundClean + 0.18*vf.Lighting + 0.18*vf.ProductDisplay
		segments = append(segments, domain.ScoredSegment{CandidateSegment: c, VisualFeature: vf})
	}
	return domain.ScoredSegmentSet{LiveID: in.LiveID, Segments: segments}, nil
}

func (n LiveHighlightNodes) ASRTextQualityScore(ctx context.Context, in domain.ScoredSegmentSet) (domain.ScoredSegmentSet, error) {
	for i := range in.Segments {
		text := in.Segments[i].ASRText
		effective := calcEffectiveSentenceRatio(text)
		repetition := calcRepetitionRatio(text)
		asrScore := clamp01(0.7*effective + 0.2*(1-repetition) + 0.1*0.85)
		in.Segments[i].ASRFeature = domain.ASRTextFeature{
			SegmentID:         in.Segments[i].SegmentID,
			EffectiveRatio:    effective,
			RepetitionRatio:   repetition,
			MeanASRConfidence: 0.85,
			ASRScore:          asrScore,
		}
	}
	return in, nil
}

func (n LiveHighlightNodes) LoraHighlightJudge(ctx context.Context, in domain.ScoredSegmentSet) (domain.ScoredSegmentSet, error) {
	for i := range in.Segments {
		score, err := n.Judge.JudgeHighlight(ctx, in.Segments[i].CandidateSegment)
		if err != nil {
			return domain.ScoredSegmentSet{}, err
		}
		in.Segments[i].LoraScore = score
		in.Segments[i].FinalScore = clamp01(
			0.3*in.Segments[i].VisualFeature.VisualScore +
				0.3*in.Segments[i].ASRFeature.ASRScore +
				0.2*score.SellingPoint +
				0.2*score.ConversionHint,
		)
	}
	return in, nil
}

func (n LiveHighlightNodes) EmbedCandidates(ctx context.Context, in domain.ScoredSegmentSet) (domain.ScoredSegmentSet, error) {
	if n.Embedder == nil {
		return in, nil
	}
	for i := range in.Segments {
		vec, err := n.Embedder.EmbedText(ctx, in.Segments[i].ASRText)
		if err != nil {
			return domain.ScoredSegmentSet{}, err
		}
		in.Segments[i].Embedding = vec
	}
	return in, nil
}

func (n LiveHighlightNodes) MilvusSemanticDedup(ctx context.Context, in domain.ScoredSegmentSet) (domain.ScoredSegmentSet, error) {
	threshold := n.DedupThreshold
	if threshold <= 0 {
		threshold = 0.88
	}
	collection := n.Collection
	if collection == "" {
		collection = "live_highlight_segments"
	}

	sort.SliceStable(in.Segments, func(i, j int) bool { return in.Segments[i].FinalScore > in.Segments[j].FinalScore })
	kept := make([]domain.ScoredSegment, 0, len(in.Segments))

	for _, seg := range in.Segments {
		if len(seg.Embedding) == 0 || n.VectorStore == nil {
			kept = append(kept, seg)
			continue
		}
		filter := fmt.Sprintf("live_id == '%s' && product_id == '%s'", seg.LiveID, seg.Product.ProductID)
		res, err := n.VectorStore.Search(ctx, collection, seg.Embedding, 5, filter)
		if err != nil {
			return domain.ScoredSegmentSet{}, err
		}
		duplicate := false
		for _, r := range res {
			if r.Score >= threshold {
				duplicate = true
				break
			}
		}
		if duplicate {
			continue
		}
		kept = append(kept, seg)
		err = n.VectorStore.Upsert(ctx, collection, []vector.VectorRecord{{
			ID:     seg.SegmentID,
			Vector: seg.Embedding,
			Metadata: map[string]any{
				"live_id":    seg.LiveID,
				"product_id": seg.Product.ProductID,
				"sku":        seg.Product.SKU,
				"score":      seg.FinalScore,
			},
		}})
		if err != nil {
			return domain.ScoredSegmentSet{}, err
		}
	}
	return domain.ScoredSegmentSet{LiveID: in.LiveID, Segments: kept}, nil
}

func (n LiveHighlightNodes) TopKRerank(ctx context.Context, in domain.ScoredSegmentSet) (domain.HighlightOutput, error) {
	topK := n.TopK
	if topK <= 0 {
		topK = 15
	}
	sort.SliceStable(in.Segments, func(i, j int) bool {
		return rerankScore(in.Segments[i]) > rerankScore(in.Segments[j])
	})
	selected := make([]domain.ScoredSegment, 0, topK)
	seenSKU := map[string]int{}
	for _, seg := range in.Segments {
		if seenSKU[seg.Product.SKU] >= 2 {
			continue
		}
		selected = append(selected, seg)
		seenSKU[seg.Product.SKU]++
		if len(selected) >= topK {
			break
		}
	}
	return domain.HighlightOutput{LiveID: in.LiveID, TopSegments: selected, TotalInput: len(in.Segments), AfterDedup: len(in.Segments)}, nil
}

func rerankScore(seg domain.ScoredSegment) float64 {
	return seg.FinalScore + categoryGuideBonus(seg)
}

func categoryGuideBonus(seg domain.ScoredSegment) float64 {
	s := strings.ToLower(seg.ASRText)
	switch seg.Product.Category {
	case "女装":
		if strings.Contains(s, "搭配") || strings.Contains(s, "版型") {
			return 0.03
		}
	case "男装":
		if strings.Contains(s, "尺码") || strings.Contains(s, "上身") {
			return 0.03
		}
	case "休闲服":
		if strings.Contains(s, "促销") || strings.Contains(s, "颜色") {
			return 0.03
		}
	}
	return 0
}

func calcEffectiveSentenceRatio(text string) float64 {
	if strings.TrimSpace(text) == "" {
		return 0
	}
	keywords := []string{"价格", "优惠", "材质", "尺码", "上身", "颜色", "加购", "下单", "链接", "卖点"}
	hit := 0
	for _, k := range keywords {
		if strings.Contains(text, k) {
			hit++
		}
	}
	return clamp01(float64(hit) / 4.0)
}

func calcRepetitionRatio(text string) float64 {
	words := strings.Fields(text)
	if len(words) <= 1 {
		return 0
	}
	seen := map[string]int{}
	dup := 0
	for _, w := range words {
		seen[w]++
		if seen[w] > 1 {
			dup++
		}
	}
	return clamp01(float64(dup) / float64(len(words)))
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
