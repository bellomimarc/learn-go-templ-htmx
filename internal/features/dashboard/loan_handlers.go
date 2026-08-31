package dashboard

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	dashboardviews "github.com/marcello/saas-poc/internal/features/dashboard/views"
	"github.com/marcello/saas-poc/internal/loans"
)

func handleLoanApplicationPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	locale := dashboardviews.LoadLocale(r.URL.Query().Get("lang"))

	form := loans.FormData{}
	state := loans.ValidationState{HasEvaluated: false}

	if err := dashboardviews.LoanApplicationPage(locale, form, state).Render(r.Context(), w); err != nil {
		http.Error(w, "Error rendering loan application page", http.StatusInternalServerError)
		log.Printf("Error rendering loan application page: %v\n", err)
	}
}

func handleLoanValidation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	locale := dashboardviews.LoadLocale(r.URL.Query().Get("lang"))

	form := parseLoanFormData(r)
	state := loans.Validate(form, true, locale)
	time.Sleep(150 * time.Millisecond)

	if err := dashboardviews.LoanValidationResponse(locale, state).Render(r.Context(), w); err != nil {
		http.Error(w, "Error rendering validation response", http.StatusInternalServerError)
		log.Printf("Error rendering validation response: %v\n", err)
	}
}

func handleLoanSubmission(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	locale := dashboardviews.LoadLocale(r.URL.Query().Get("lang"))

	form := parseLoanFormData(r)
	state := loans.Validate(form, true, locale)

	if !state.IsValid {
		w.WriteHeader(http.StatusUnprocessableEntity)
		if err := dashboardviews.LoanSubmissionResult(locale, locale.Text("loan.submit.invalid"), "error").Render(r.Context(), w); err != nil {
			http.Error(w, "Error rendering invalid submission response", http.StatusInternalServerError)
			log.Printf("Error rendering invalid submission response: %v\n", err)
		}
		return
	}

	applicationID := fmt.Sprintf("LN-%d", time.Now().UnixNano()%1000000000)
	message := strings.NewReplacer("{id}", applicationID, "{installment}", fmt.Sprintf("%.2f", state.EstimatedInstallment)).Replace(locale.Text("loan.submit.success"))

	if err := dashboardviews.LoanSubmissionResult(locale, message, "info").Render(r.Context(), w); err != nil {
		http.Error(w, "Error rendering submission response", http.StatusInternalServerError)
		log.Printf("Error rendering submission response: %v\n", err)
	}
}

func parseLoanFormData(r *http.Request) loans.FormData {
	_ = r.ParseForm()

	return loans.FormData{
		FullName:        strings.TrimSpace(r.FormValue("full_name")),
		BirthDate:       strings.TrimSpace(r.FormValue("birth_date")),
		EmploymentType:  strings.TrimSpace(r.FormValue("employment_type")),
		AnnualIncome:    strings.TrimSpace(r.FormValue("annual_income")),
		MonthlyDebt:     strings.TrimSpace(r.FormValue("monthly_debt")),
		LoanAmount:      strings.TrimSpace(r.FormValue("loan_amount")),
		LoanYears:       strings.TrimSpace(r.FormValue("loan_years")),
		DownPayment:     strings.TrimSpace(r.FormValue("down_payment")),
		CollateralValue: strings.TrimSpace(r.FormValue("collateral_value")),
		HasGuarantor:    r.FormValue("has_guarantor") != "",
		GuarantorIncome: strings.TrimSpace(r.FormValue("guarantor_income")),
		Purpose:         strings.TrimSpace(r.FormValue("purpose")),
	}
}
