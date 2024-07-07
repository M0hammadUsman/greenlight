package validator

import (
	"net/mail"
)

type Validator struct {
	Errors map[string]string
}

func New() *Validator {
	return &Validator{make(map[string]string)}
}

// Valid returns true if the errors map doesn't contain any entries.
func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

func (v *Validator) addError(key, message string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}

// Check adds an error message to the map only if a validation check is not 'ok'.
func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.addError(key, message)
	}
}

func In(value string, list ...string) bool {
	for i := range list {
		if value == list[i] {
			return true
		}
	}
	return false
}

func ValidEmail(email string) bool {
	if _, err := mail.ParseAddress(email); err != nil {
		return false
	}
	return true
}

// Unique returns true if all string values in a slice are unique.
func Unique(values []string) bool {
	unique := make(map[string]bool)
	for _, v := range values {
		unique[v] = true
	}
	return len(values) == len(unique)
}
