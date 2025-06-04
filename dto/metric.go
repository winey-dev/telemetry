package dto

type Metric struct {
	Category    string   `json:"category"`
	SubCategory string   `json:"sub_category"`
	ItemName    string   `json:"item_name"`
	Description string   `json:"description"`
	TagNames    []string `json:"tag_names"`
	TagValues   []string `json:"tag_values"`
	Value       float64  `json:"value"`
}
