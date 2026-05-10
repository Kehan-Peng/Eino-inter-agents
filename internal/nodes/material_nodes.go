package nodes

import (
	"context"
	"strings"

	"eino-intern-workflows/internal/clients"
	"eino-intern-workflows/internal/domain"
)

type MaterialNode struct {
	Value    string
	Children []*MaterialNode
}

type MaterialNodes struct {
	ES              clients.MaterialSearchClient
	Root            *MaterialNode
	Synonyms        map[string]string
	ValueSenseLevel map[string]int64
	TopK            int
}

func (n MaterialNodes) ESRecallCandidates(ctx context.Context, in domain.MaterialCheckInput) (domain.MaterialRecallResult, error) {
	topK := n.TopK
	if topK <= 0 {
		topK = 5
	}
	left, err := n.ES.RecallMaterialCandidates(ctx, in.LeftMaterial, in.Category, topK)
	if err != nil {
		return domain.MaterialRecallResult{}, err
	}
	right, err := n.ES.RecallMaterialCandidates(ctx, in.RightMaterial, in.Category, topK)
	if err != nil {
		return domain.MaterialRecallResult{}, err
	}
	return domain.MaterialRecallResult{Input: in, LeftCands: left, RightCands: right}, nil
}

func (n MaterialNodes) NormalizeSynonyms(ctx context.Context, in domain.MaterialRecallResult) (domain.MaterialNormalized, error) {
	left := expandWords(in.Input.LeftMaterial, in.LeftCands, n.Synonyms)
	right := expandWords(in.Input.RightMaterial, in.RightCands, n.Synonyms)
	return domain.MaterialNormalized{Input: in.Input, LeftWords: left, RightWords: right, Candidates: in}, nil
}

func (n MaterialNodes) MaterialTreeMatch(ctx context.Context, in domain.MaterialNormalized) (domain.MaterialTreeDecision, error) {
	if intersects(in.LeftWords, in.RightWords) {
		return domain.MaterialTreeDecision{Input: in.Input, CheckResult: "材质相同"}, nil
	}
	leftPath := n.firstPath(in.LeftWords)
	rightPath := n.firstPath(in.RightWords)
	common := lastCommon(leftPath, rightPath)
	result := "材质不同"
	if common != "" && common != n.rootName() {
		result = "材质相同"
	}
	return domain.MaterialTreeDecision{
		Input:          in.Input,
		CheckResult:    result,
		LeftPath:       leftPath,
		RightPath:      rightPath,
		CommonAncestor: common,
	}, nil
}

func (n MaterialNodes) CompareValueLevel(ctx context.Context, in domain.MaterialTreeDecision) (domain.MaterialValueDecision, error) {
	leftWord, leftLevel := n.maxValueWord(in.LeftPath, in.Input.LeftMaterial)
	rightWord, rightLevel := n.maxValueWord(in.RightPath, in.Input.RightMaterial)

	valueWords := []string{}
	if leftLevel > 0 {
		valueWords = append(valueWords, leftWord)
	} else {
		valueWords = append(valueWords, "左i非高价值材质")
	}
	if rightLevel > 0 {
		valueWords = append(valueWords, rightWord)
	} else {
		valueWords = append(valueWords, "右i非高价值材质")
	}

	valueLevelResult := "材质不同"
	checkResult := in.CheckResult
	if leftLevel > rightLevel {
		valueLevelResult = "左i材质价值更高"
		checkResult = "材质不同"
	} else if leftLevel < rightLevel {
		valueLevelResult = "右i材质价值更高"
		checkResult = "材质相同"
	} else if leftLevel > 0 && rightLevel > 0 {
		valueLevelResult = "材质价值相同"
	}
	in.CheckResult = checkResult
	return domain.MaterialValueDecision{TreeDecision: in, ValueSenseWord: valueWords, ValueLevelResult: valueLevelResult}, nil
}

func (n MaterialNodes) BuildMaterialOutput(ctx context.Context, in domain.MaterialValueDecision) (domain.MaterialCheckOutput, error) {
	return domain.MaterialCheckOutput{
		CheckResult:      in.TreeDecision.CheckResult,
		ValueSenseWord:   in.ValueSenseWord,
		ValueLevelResult: in.ValueLevelResult,
		LeftPath:         in.TreeDecision.LeftPath,
		RightPath:        in.TreeDecision.RightPath,
		CommonAncestor:   in.TreeDecision.CommonAncestor,
	}, nil
}

func (n MaterialNodes) firstPath(words []string) []string {
	for _, w := range words {
		visited := []string{}
		if lowestCommonAncestorPath(n.Root, w, &visited) != nil {
			return visited
		}
	}
	return nil
}

// lowestCommonAncestorPath is intentionally named to mirror the internship code.
// Strictly speaking it is not a two-node LCA algorithm; it is a DFS target lookup
// that appends target->root path during recursion unwinding.
func lowestCommonAncestorPath(root *MaterialNode, target string, visited *[]string) *MaterialNode {
	if root == nil {
		return nil
	}
	if strings.EqualFold(root.Value, strings.TrimSpace(target)) {
		*visited = append(*visited, root.Value)
		return root
	}
	for _, child := range root.Children {
		if lowestCommonAncestorPath(child, target, visited) != nil {
			*visited = append(*visited, root.Value)
			return root
		}
	}
	return nil
}

func (n MaterialNodes) maxValueWord(path []string, raw string) (string, int64) {
	words := append([]string{normalize(raw)}, path...)
	maxWord := ""
	var maxLevel int64
	for _, w := range words {
		if lv, ok := n.ValueSenseLevel[w]; ok && lv > maxLevel {
			maxLevel = lv
			maxWord = w
		}
	}
	return maxWord, maxLevel
}

func (n MaterialNodes) rootName() string {
	if n.Root == nil {
		return ""
	}
	return n.Root.Value
}

func expandWords(raw string, cands []domain.MaterialCandidate, synonyms map[string]string) []string {
	m := map[string]bool{}
	add := func(w string) {
		w = normalize(w)
		if w != "" {
			m[w] = true
			if std, ok := synonyms[w]; ok {
				m[normalize(std)] = true
				for k, v := range synonyms {
					if normalize(v) == normalize(std) {
						m[normalize(k)] = true
					}
				}
			}
		}
	}
	add(raw)
	for _, c := range cands {
		add(c.RawText)
		add(c.StandardName)
	}
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func intersects(a, b []string) bool {
	m := map[string]bool{}
	for _, x := range a {
		m[normalize(x)] = true
	}
	for _, y := range b {
		if m[normalize(y)] {
			return true
		}
	}
	return false
}

func lastCommon(pathA, pathB []string) string {
	set := map[string]bool{}
	for _, x := range pathA {
		set[normalize(x)] = true
	}
	for _, y := range pathB {
		if set[normalize(y)] {
			return y
		}
	}
	return ""
}

func normalize(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
