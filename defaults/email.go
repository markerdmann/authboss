package defaults

import "net/mail"

// validateEmail checks if the given email address is valid using the ParseAddress
// function from the net/mail package.
func validateEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}