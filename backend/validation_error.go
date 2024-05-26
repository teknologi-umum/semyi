package main

import "fmt"

type ValidationError struct {
	Issues []ValidationIssue `json:"issues"`
}

type ValidationIssue struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func NewValidationError() *ValidationError {
	return &ValidationError{
		Issues: []ValidationIssue{},
	}
}

func (v *ValidationError) AddIssue(field, message string) {
	v.Issues = append(v.Issues, ValidationIssue{
		Field:   field,
		Message: message,
	})
}

func (v *ValidationError) Error() string {
	return fmt.Sprintf("validation error: %v", v.Issues)
}

func (v *ValidationError) HasIssues() bool {
	return len(v.Issues) > 0
}
