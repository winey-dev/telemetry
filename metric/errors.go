package metric

import "errors"

var (
	ErrInvalidTagValues = errors.New("invalid tag values")
	ErrRequiredTagNames = errors.New("required tag names are missing")
	ErrRequiredFields   = errors.New("required fields are missing in the item options")
)
