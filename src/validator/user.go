package validator

import "regexp"

// ValidatePassword validates the password
//
// Return true if the password is valid
func ValidatePassword(password string) bool {
	passReg := regexp.MustCompile(`^[a-zA-Z0-9!@#$%^&*()_=+]{6,32}$`)
	return passReg.MatchString(password)
}

// ValidateEmail validates the NJUPT email
//
// Return true if the email is valid
func ValidateEmail(email string) bool {
	// "^[BPFQbpfq](1[7-9]|2[0-9])([0-3])\\d{5}@njupt.edu.cn$" Matches the email format of NJUPT
	emailReg := regexp.MustCompile(`^[BPFQbpfq](1[7-9]|2[0-9])([0-3])\\d{5}@njupt.edu.cn$`)
	return emailReg.MatchString(email)
}
