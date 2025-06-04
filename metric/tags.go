package metric

type ConstraintTags struct {
	TagNames  []string
	TagValues []string
}

func NewConstraintTags(tagNames, tagValues []string) ConstraintTags {
	return ConstraintTags{
		TagNames:  tagNames,
		TagValues: tagValues,
	}
}
func (ct ConstraintTags) IsValid() bool {
	if len(ct.TagNames) != len(ct.TagValues) {
		return false
	}
	for _, tagValue := range ct.TagValues {
		if tagValue == "" {
			return false
		}
	}
	return true
}

func (ct ConstraintTags) IsEmpty() bool {
	if !ct.IsValid() {
		return false
	}
	return len(ct.TagNames) == 0 && len(ct.TagValues) == 0
}

func (ct ConstraintTags) Len() int {
	if !ct.IsValid() {
		return 0
	}
	return len(ct.TagNames)
}
func makeTagValues(constraintValues, values []string) []string {
	if len(constraintValues) == 0 {
		return values
	}
	if len(values) == 0 {
		return constraintValues
	}

	tagValues := make([]string, 0, len(constraintValues)+len(values))
	tagValues = append(tagValues, constraintValues...)
	tagValues = append(tagValues, values...)
	return tagValues
}
