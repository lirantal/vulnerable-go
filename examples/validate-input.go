package main

import (
	"fmt"
	"regexp"
	
)

func isValidTenantID(tenantID string) bool {
	// Implement a regex pattern to validate tenantID format
	// Example: only allow alphanumeric characters
	validTenantIDPattern := `^[a-zA-Z0-9]+$`
	matched, _ := regexp.MatchString(validTenantIDPattern, tenantID)
	return matched
}

func main() {

	// Example input based on the mocked storage server in fixtures/http directory of the project:
	tenantID := "3971533981712"
	isValid := isValidTenantID(tenantID)
	fmt.Println("Is tenantID valid?", isValid)
}