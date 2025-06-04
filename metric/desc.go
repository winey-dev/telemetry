package metric

type Desc struct {
	Category       string
	SubCategory    string
	ItemName       string
	Description    string
	ConstraintTags ConstraintTags
	TagNames       []string
}

func NewDesc(category, subCategory, itemName, description string, constraintTags ConstraintTags, TagNames ...string) *Desc {
	return &Desc{
		Category:       category,
		SubCategory:    subCategory,
		ItemName:       itemName,
		Description:    description,
		ConstraintTags: constraintTags,
		TagNames:       TagNames,
	}
}

func (d *Desc) TagNamesWithConstraint() []string {
	if d.ConstraintTags.Len() == 0 {
		return d.TagNames
	}
	tagNames := make([]string, 0, len(d.TagNames)+d.ConstraintTags.Len())
	tagNames = append(tagNames, d.TagNames...)
	tagNames = append(tagNames, d.ConstraintTags.TagNames...)
	return tagNames
}

func (d *Desc) String() string {
	return d.Category + "." + d.SubCategory + "." + d.ItemName
}
