package dto

import "github.com/jinzhu/copier"

// CopyStruct copies fields from src to dst using jinzhu/copier.
// Returns error if field copy fails.
func CopyStruct(dst, src interface{}) error {
	return copier.Copy(dst, src)
}
