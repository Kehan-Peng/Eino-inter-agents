package workflows

import (
	"context"

	"github.com/cloudwego/eino/compose"

	"eino-intern-workflows/internal/domain"
	"eino-intern-workflows/internal/nodes"
)

// BuildMaterialCheckWorkflow reconstructs the Go backend material plugin as a workflow:
// START -> es_recall -> synonym_normalize -> material_tree_match -> value_compare -> output -> END.
func BuildMaterialCheckWorkflow(ctx context.Context, ns nodes.MaterialNodes) (compose.Runnable[domain.MaterialCheckInput, domain.MaterialCheckOutput], error) {
	wf := compose.NewWorkflow[domain.MaterialCheckInput, domain.MaterialCheckOutput]()

	wf.AddLambdaNode("es_recall_candidates", compose.InvokableLambda(ns.ESRecallCandidates)).
		AddInput(compose.START)

	wf.AddLambdaNode("normalize_synonyms", compose.InvokableLambda(ns.NormalizeSynonyms)).
		AddInput("es_recall_candidates")

	wf.AddLambdaNode("material_tree_match", compose.InvokableLambda(ns.MaterialTreeMatch)).
		AddInput("normalize_synonyms")

	wf.AddLambdaNode("compare_value_level", compose.InvokableLambda(ns.CompareValueLevel)).
		AddInput("material_tree_match")

	wf.AddLambdaNode("build_material_output", compose.InvokableLambda(ns.BuildMaterialOutput)).
		AddInput("compare_value_level")

	wf.End().AddInput("build_material_output")
	return wf.Compile(ctx)
}
