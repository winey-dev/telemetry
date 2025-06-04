package metric

import "github.com/winey-dev/telemetry/dto"

type Metric interface {
	Desc() *Desc
	Write(*dto.Metric) error
}

type Opts struct {
	Category       string
	SubCategory    string
	ItemName       string
	Description    string
	ConstraintTags ConstraintTags
}
