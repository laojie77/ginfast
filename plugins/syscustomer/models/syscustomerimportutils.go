package models

import "regexp"

var customerImportMobilePattern = regexp.MustCompile(`^1[3-9]\d{9}$`)

func IsValidImportCustomerMobile(mobile string) bool {
	return customerImportMobilePattern.MatchString(mobile)
}
