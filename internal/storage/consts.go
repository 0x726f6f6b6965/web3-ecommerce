package storage

import (
	"errors"
)

const (
	Pk          string = "pk"
	Sk          string = "sk"
	SoftDeleted string = "soft_deleted"

	SoftDeletedIndex string = "soft_deleted_index"
	PkNotExists      string = "attribute_not_exists(pk)"
	PkExists         string = "attribute_exists(pk)"
)

var (
	UserKey    = "USER#%s"
	ProfileKey = "#PROFILE#%s"
	OrderKey   = "ORDER#%s"
	ProductKey = "PRODUCT#%s"

	ErrNotFound = errors.New("data not found")
)
