package postgres

import "errors"

var errInconsistentProductData = errors.New(
	"inconsistent persisted product data",
)