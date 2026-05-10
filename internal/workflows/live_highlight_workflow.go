package workflows

import (
	"context"

	"github.com/cloudwego/eino/compose"

	"eino-intern-workflows/internal/domain"
	"eino-intern-workflows/internal/nodes"
)

// BuildLiveHighlightWorkflow reconstructs the long-live highlight candidate workflow:
// START -> segment -> visual_score -> asr_score -> lora_judge -> embed -> milvus_dedup -> topk_rerank -> END.
func BuildLiveHighlightWorkflow(ctx context.Context, ns nodes.LiveHighlightNodes) (compose.Runnable[domain.LiveHighlightInput, domain.HighlightOutput], error) {
	wf := compose.NewWorkflow[domain.LiveHighlightInput, domain.HighlightOutput]()

	wf.AddLambdaNode("segment_candidates", compose.InvokableLambda(ns.SegmentCandidates)).
		AddInput(compose.START)

	wf.AddLambdaNode("visual_quality_score", compose.InvokableLambda(ns.VisualQualityScore)).
		AddInput("segment_candidates")

	wf.AddLambdaNode("asr_text_quality_score", compose.InvokableLambda(ns.ASRTextQualityScore)).
		AddInput("visual_quality_score")

	wf.AddLambdaNode("lora_highlight_judge", compose.InvokableLambda(ns.LoraHighlightJudge)).
		AddInput("asr_text_quality_score")

	wf.AddLambdaNode("embed_candidates", compose.InvokableLambda(ns.EmbedCandidates)).
		AddInput("lora_highlight_judge")

	wf.AddLambdaNode("milvus_semantic_dedup", compose.InvokableLambda(ns.MilvusSemanticDedup)).
		AddInput("embed_candidates")

	wf.AddLambdaNode("topk_rerank", compose.InvokableLambda(ns.TopKRerank)).
		AddInput("milvus_semantic_dedup")

	wf.End().AddInput("topk_rerank")
	return wf.Compile(ctx)
}
