package dashboard

import "github.com/marcello/saas-poc/internal/loans"

// LoanFormData stores the raw form payload so user input can be re-rendered unchanged.
type LoanFormData = loans.FormData

// ValidationIssue links a single message to a specific form field.
type ValidationIssue = loans.ValidationIssue

// LoanValidationState contains both validation results and derived financial metrics.
type LoanValidationState = loans.ValidationState
