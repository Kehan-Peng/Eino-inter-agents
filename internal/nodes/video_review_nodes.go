package nodes

import (
	"context"
	"fmt"
	"strings"

	"eino-intern-workflows/internal/clients"
	"eino-intern-workflows/internal/domain"
	"eino-intern-workflows/internal/llm"
)

type VideoReviewNodes struct {
	Rules clients.RuleRetriever
	Judge llm.ReviewJudge
	TopK  int
}

func (n VideoReviewNodes) BuildReviewQuery(ctx context.Context, in domain.VideoReviewInput) (domain.ReviewQuery, error) {
	asrTexts := make([]string, 0, len(in.ASR))
	for _, seg := range in.ASR {
		text := strings.TrimSpace(seg.Text)
		if text != "" {
			asrTexts = append(asrTexts, text)
		}
	}
	frameFacts := []string{}
	if len(in.Frames) > 0 {
		frameFacts = append(frameFacts, fmt.Sprintf("frame_count=%d", len(in.Frames)))
	}
	problemHint := inferProblemHint(in)
	query := strings.Join([]string{
		in.Product.Category,
		in.Product.Title,
		problemHint,
		strings.Join(asrTexts, " "),
	}, " ")
	return domain.ReviewQuery{
		VideoID:     in.VideoID,
		Query:       query,
		Product:     in.Product,
		FrameFacts:  frameFacts,
		ASRSummary:  strings.Join(asrTexts, " "),
		ProblemHint: problemHint,
	}, nil
}

func (n VideoReviewNodes) RetrieveRules(ctx context.Context, in domain.ReviewQuery) (domain.RetrievedRules, error) {
	topK := n.TopK
	if topK <= 0 {
		topK = 5
	}
	rules, err := n.Rules.RetrieveRules(ctx, in, topK)
	if err != nil {
		return domain.RetrievedRules{}, err
	}
	return domain.RetrievedRules{VideoID: in.VideoID, Query: in, Rules: rules}, nil
}

func (n VideoReviewNodes) BuildReviewPrompt(ctx context.Context, in domain.RetrievedRules) (domain.ReviewPrompt, error) {
	var b strings.Builder
	b.WriteString("你是电商视频质量预审核模型。只能基于召回规则和视频内容判断，不要编造规则。\n")
	b.WriteString("输出 JSON 字段：quality_labels, score, reason, rule_ids, confidence。\n")
	b.WriteString("商品类目：" + in.Query.Product.Category + "\n")
	b.WriteString("商品标题：" + in.Query.Product.Title + "\n")
	b.WriteString("ASR 摘要：" + in.Query.ASRSummary + "\n")
	b.WriteString("视觉事实：" + strings.Join(in.Query.FrameFacts, ";") + "\n")
	b.WriteString("召回规则：\n")
	for _, r := range in.Rules {
		b.WriteString(fmt.Sprintf("- [%s] category=%s problem=%s text=%s\n", r.RuleID, r.Category, r.ProblemType, r.Text))
	}
	return domain.ReviewPrompt{VideoID: in.VideoID, Product: in.Query.Product, Prompt: b.String(), Rules: in.Rules}, nil
}

func (n VideoReviewNodes) LLMQualityJudge(ctx context.Context, in domain.ReviewPrompt) (domain.ReviewDecision, error) {
	return n.Judge.JudgeVideoQuality(ctx, in)
}

func (n VideoReviewNodes) ValidateReviewResult(ctx context.Context, in domain.ReviewDecision) (domain.VideoReviewOutput, error) {
	needHuman := in.Confidence < 0.75 || len(in.RuleIDs) == 0
	return domain.VideoReviewOutput{
		VideoID:       in.VideoID,
		QualityLabels: in.QualityLabels,
		Score:         in.Score,
		Reason:        in.Reason,
		RuleIDs:       in.RuleIDs,
		Confidence:    in.Confidence,
		NeedHuman:     needHuman,
	}, nil
}

func inferProblemHint(in domain.VideoReviewInput) string {
	// This is the deterministic part before RAG retrieval. In production it can
	// use CV feature flags such as black screen, occlusion, product mismatch, etc.
	if in.Product.Category != "" {
		return "商品展示 画质 遮挡 商品一致性 " + in.Product.Category
	}
	return "视频质量 画质 遮挡 商品一致性"
}
