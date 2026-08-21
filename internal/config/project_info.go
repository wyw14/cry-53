package config

type ProjectInfo struct {
	Name            string `json:"name"`
	RequirementID   string `json:"requirement_id"`
	GlobalSequence  int    `json:"global_sequence"`
	SourceCategory  string `json:"source_category"`
	SourceReference string `json:"source_reference"`
}

var Info = ProjectInfo{
	Name:            "数据源配置包校验与发布平台",
	RequirementID:   "GO-CG-048",
	GlobalSequence:  53,
	SourceCategory:  "代码生成",
	SourceReference: "word2.xlsx / Sheet1 / 第 444 行",
}
