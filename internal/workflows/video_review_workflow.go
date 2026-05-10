package workflows

import (
	"context"

	"github.com/cloudwego/eino/compose"

	"eino-intern-workflows/internal/domain"
	"eino-intern-workflows/internal/nodes"
)

// BuildVideoQualityReviewWorkflow reconstructs the internship RAG pre-review flow:
// START -> build_query -> retrieve_rules -> build_prompt -> llm_quality_judge -> validate -> END.
func BuildVideoQualityReviewWorkflow(ctx context.Context, ns nodes.VideoReviewNodes) (compose.Runnable[domain.VideoReviewInput, domain.VideoReviewOutput], error) {
	wf := compose.NewWorkflow[domain.VideoReviewInput, domain.VideoReviewOutput]()

	wf.AddLambdaNode("build_query", compose.InvokableLambda(ns.BuildReviewQuery)).
		AddInput(compose.START)

	wf.AddLambdaNode("retrieve_rules", compose.InvokableLambda(ns.RetrieveRules)).
		AddInput("build_query")

	wf.AddLambdaNode("build_prompt", compose.InvokableLambda(ns.BuildReviewPrompt)).
		AddInput("retrieve_rules")

	wf.AddLambdaNode("llm_quality_judge", compose.InvokableLambda(ns.LLMQualityJudge)).
		AddInput("build_prompt")

	wf.AddLambdaNode("validate_output", compose.InvokableLambda(ns.ValidateReviewResult)).
		AddInput("llm_quality_judge")

	wf.End().AddInput("validate_output")
	return wf.Compile(ctx)
}
