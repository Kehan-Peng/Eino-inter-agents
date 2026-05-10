package domain

// -----------------------------
// Common domain objects.
// -----------------------------

type VideoFrame struct {
	URI       string  `json:"uri"`
	Timestamp float64 `json:"timestamp"`
}

type ASRSegment struct {
	Start      float64 `json:"start"`
	End        float64 `json:"end"`
	Text       string  `json:"text"`
	Confidence float64 `json:"confidence"`
}

type ProductInfo struct {
	ProductID string            `json:"product_id"`
	SKU       string            `json:"sku"`
	Title     string            `json:"title"`
	Category  string            `json:"category"`
	Brand     string            `json:"brand"`
	Price     float64           `json:"price"`
	Attrs     map[string]string `json:"attrs"`
}

type ConversionEvent struct {
	Type      string  `json:"type"` // click/add_cart/order
	Timestamp float64 `json:"timestamp"`
	ProductID string  `json:"product_id"`
	Value     float64 `json:"value"`
}

// -----------------------------
// RAG video quality review.
// -----------------------------

type VideoReviewInput struct {
	VideoID       string            `json:"video_id"`
	Frames        []VideoFrame      `json:"frames"`
	ASR           []ASRSegment      `json:"asr"`
	Product       ProductInfo       `json:"product"`
	ExtraMetadata map[string]string `json:"extra_metadata"`
}

type RuleChunk struct {
	RuleID      string            `json:"rule_id"`
	Category    string            `json:"category"`
	ProblemType string            `json:"problem_type"`
	Text        string            `json:"text"`
	Score       float64           `json:"score"`
	Metadata    map[string]string `json:"metadata"`
}

type ReviewQuery struct {
	VideoID     string      `json:"video_id"`
	Query       string      `json:"query"`
	Product     ProductInfo `json:"product"`
	FrameFacts  []string    `json:"frame_facts"`
	ASRSummary  string      `json:"asr_summary"`
	ProblemHint string      `json:"problem_hint"`
}

type RetrievedRules struct {
	VideoID string      `json:"video_id"`
	Query   ReviewQuery `json:"query"`
	Rules   []RuleChunk `json:"rules"`
}

type ReviewPrompt struct {
	VideoID string      `json:"video_id"`
	Product ProductInfo `json:"product"`
	Prompt  string      `json:"prompt"`
	Rules   []RuleChunk `json:"rules"`
}

type ReviewDecision struct {
	VideoID       string   `json:"video_id"`
	QualityLabels []string `json:"quality_labels"`
	Score         float64  `json:"score"`
	Reason        string   `json:"reason"`
	RuleIDs       []string `json:"rule_ids"`
	Confidence    float64  `json:"confidence"`
}

type VideoReviewOutput struct {
	VideoID       string   `json:"video_id"`
	QualityLabels []string `json:"quality_labels"`
	Score         float64  `json:"score"`
	Reason        string   `json:"reason"`
	RuleIDs       []string `json:"rule_ids"`
	Confidence    float64  `json:"confidence"`
	NeedHuman     bool     `json:"need_human"`
}

// -----------------------------
// Live highlight workflow.
// -----------------------------

type LiveHighlightInput struct {
	LiveID           string            `json:"live_id"`
	Frames           []VideoFrame      `json:"frames"`
	ASR              []ASRSegment      `json:"asr"`
	Products         []ProductInfo     `json:"products"`
	ConversionEvents []ConversionEvent `json:"conversion_events"`
}

type CandidateSegment struct {
	SegmentID string      `json:"segment_id"`
	LiveID    string      `json:"live_id"`
	Product   ProductInfo `json:"product"`
	Start     float64     `json:"start"`
	End       float64     `json:"end"`
	ASRText   string      `json:"asr_text"`
	Duration  float64     `json:"duration"`
}

type CandidateSet struct {
	LiveID     string             `json:"live_id"`
	Candidates []CandidateSegment `json:"candidates"`
}

type VisualFeature struct {
	SegmentID       string  `json:"segment_id"`
	Sharpness       float64 `json:"sharpness"`
	Stability       float64 `json:"stability"`
	BlackScreen     float64 `json:"black_screen"`
	Occlusion       float64 `json:"occlusion"`
	CenterScore     float64 `json:"center_score"`
	BackgroundClean float64 `json:"background_clean"`
	Lighting        float64 `json:"lighting"`
	ProductDisplay  float64 `json:"product_display"`
	VisualScore     float64 `json:"visual_score"`
}

type ASRTextFeature struct {
	SegmentID         string  `json:"segment_id"`
	EffectiveRatio    float64 `json:"effective_ratio"`
	RepetitionRatio   float64 `json:"repetition_ratio"`
	MeanASRConfidence float64 `json:"mean_asr_confidence"`
	ASRScore          float64 `json:"asr_score"`
}

type LoraHighlightScore struct {
	SegmentID      string  `json:"segment_id"`
	IsHighlight    bool    `json:"is_highlight"`
	QualityScore   float64 `json:"quality_score"`
	SellingPoint   float64 `json:"selling_point"`
	ConversionHint float64 `json:"conversion_hint"`
	Reason         string  `json:"reason"`
}

type ScoredSegment struct {
	CandidateSegment
	VisualFeature VisualFeature      `json:"visual_feature"`
	ASRFeature    ASRTextFeature     `json:"asr_feature"`
	LoraScore     LoraHighlightScore `json:"lora_score"`
	FinalScore    float64            `json:"final_score"`
	Embedding     []float32          `json:"embedding,omitempty"`
}

type ScoredSegmentSet struct {
	LiveID   string          `json:"live_id"`
	Segments []ScoredSegment `json:"segments"`
}

type HighlightOutput struct {
	LiveID      string          `json:"live_id"`
	TopSegments []ScoredSegment `json:"top_segments"`
	TotalInput  int             `json:"total_input"`
	AfterDedup  int             `json:"after_dedup"`
}

// -----------------------------
// Material Check workflow.
// -----------------------------

type MaterialCheckInput struct {
	LeftMaterial  string `json:"left_material"`
	RightMaterial string `json:"right_material"`
	Category      string `json:"category"`
}

type MaterialCandidate struct {
	RawText      string  `json:"raw_text"`
	StandardName string  `json:"standard_name"`
	Score        float64 `json:"score"`
}

type MaterialRecallResult struct {
	Input      MaterialCheckInput  `json:"input"`
	LeftCands  []MaterialCandidate `json:"left_candidates"`
	RightCands []MaterialCandidate `json:"right_candidates"`
}

type MaterialNormalized struct {
	Input      MaterialCheckInput   `json:"input"`
	LeftWords  []string             `json:"left_words"`
	RightWords []string             `json:"right_words"`
	Candidates MaterialRecallResult `json:"candidates"`
}

type MaterialTreeDecision struct {
	Input          MaterialCheckInput `json:"input"`
	CheckResult    string             `json:"check_result"`
	LeftPath       []string           `json:"left_path"`
	RightPath      []string           `json:"right_path"`
	CommonAncestor string             `json:"common_ancestor"`
}

type MaterialValueDecision struct {
	TreeDecision     MaterialTreeDecision `json:"tree_decision"`
	ValueSenseWord   []string             `json:"value_sense_word"`
	ValueLevelResult string               `json:"value_level_result"`
}

type MaterialCheckOutput struct {
	CheckResult      string   `json:"check_result"`
	ValueSenseWord   []string `json:"value_sense_word"`
	ValueLevelResult string   `json:"value_level_result"`
	LeftPath         []string `json:"left_path"`
	RightPath        []string `json:"right_path"`
	CommonAncestor   string   `json:"common_ancestor"`
}
