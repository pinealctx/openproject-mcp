package cmd

import "fmt"

func validatePagination(offset, pageSize int) error {
	if offset < 1 || pageSize < 1 {
		return fmt.Errorf("offset and page-size must be greater than zero")
	}
	return nil
}
