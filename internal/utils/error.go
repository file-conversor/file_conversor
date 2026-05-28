// internal/utils/error.go

package utils

// ExtractError is a helper function to extract the error from a value that may be an error or a wrapped error.
func ExtractError(a any, err error) error {
	return err
}
