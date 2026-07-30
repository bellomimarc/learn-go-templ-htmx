package dashboard

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"

	dashboardviews "github.com/marcello/saas-poc/internal/features/dashboard/views"
)

func handleLoanApplicationPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	locale := dashboardviews.LoadLocale(r.URL.Query().Get("lang"))

	form := dashboardviews.LoanFormData{}
	state := dashboardviews.LoanValidationState{HasEvaluated: false}

	if err := dashboardviews.LoanApplicationPage(locale, form, state).Render(r.Context(), w); err != nil {
		http.Error(w, "Error rendering loan application page", http.StatusInternalServerError)
		log.Printf("Error rendering loan application page: %v\n", err)
	}
}

func handleLoanValidation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	locale := dashboardviews.LoadLocale(r.URL.Query().Get("lang"))

	form := parseLoanFormData(r)
	state := validateLoanForm(form, true, locale)
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
	state := validateLoanForm(form, true, locale)

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

func parseLoanFormData(r *http.Request) dashboardviews.LoanFormData {
	_ = r.ParseForm()

	return dashboardviews.LoanFormData{
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

func validateLoanForm(form dashboardviews.LoanFormData, hasEvaluated bool, locale dashboardviews.Locale) dashboardviews.LoanValidationState {
	state := dashboardviews.LoanValidationState{HasEvaluated: hasEvaluated}
	if !hasEvaluated {
		return state
	}

	addError := func(field, msg string) {
		state.Errors = append(state.Errors, dashboardviews.ValidationIssue{Field: field, Message: msg})
	}
	addWarning := func(msg string) {
		state.Warnings = append(state.Warnings, msg)
	}

	if form.FullName == "" {
		addError("full_name", locale.Text("loan.error.full_name.required"))
	} else if nonSpaceChars(form.FullName) < 3 {
		addError("full_name", locale.Text("loan.error.full_name.length"))
	}

	if form.EmploymentType == "" {
		addError("employment_type", locale.Text("loan.error.employment_type.required"))
	}

	annualIncome, annualIncomeOK := parsePositiveFloat(form.AnnualIncome)
	if !annualIncomeOK {
		addError("annual_income", locale.Text("loan.error.annual_income"))
	}

	monthlyDebt, monthlyDebtOK := parseNonNegativeFloat(form.MonthlyDebt)
	if !monthlyDebtOK {
		addError("monthly_debt", locale.Text("loan.error.monthly_debt"))
	}

	loanAmount, loanAmountOK := parsePositiveFloat(form.LoanAmount)
	if !loanAmountOK {
		addError("loan_amount", locale.Text("loan.error.loan_amount.positive"))
	} else if loanAmount <= 100 {
		addError("loan_amount", locale.Text("loan.error.loan_amount.minimum"))
	}

	loanYears, loanYearsOK := parsePositiveInt(form.LoanYears)
	if !loanYearsOK {
		addError("loan_years", locale.Text("loan.error.loan_years"))
	}

	downPayment, downPaymentOK := parseNonNegativeFloat(form.DownPayment)
	if !downPaymentOK {
		addError("down_payment", locale.Text("loan.error.down_payment"))
	}

	collateralValue := 0.0
	collateralValueOK := true
	if form.CollateralValue != "" {
		collateralValue, collateralValueOK = parseNonNegativeFloat(form.CollateralValue)
		if !collateralValueOK {
			addError("collateral_value", locale.Text("loan.error.collateral_value"))
		}
	}

	guarantorIncome := 0.0
	guarantorIncomeOK := true
	if form.HasGuarantor {
		if form.GuarantorIncome == "" {
			addError("guarantor_income", locale.Text("loan.error.guarantor_income.required"))
			guarantorIncomeOK = false
		} else {
			guarantorIncome, guarantorIncomeOK = parsePositiveFloat(form.GuarantorIncome)
			if !guarantorIncomeOK {
				addError("guarantor_income", locale.Text("loan.error.guarantor_income.positive"))
			}
		}
	} else if form.GuarantorIncome != "" {
		addWarning(locale.Text("loan.warning.guarantor_income.unused"))
	}

	birthDate, birthDateOK := parseBirthDate(form.BirthDate)
	if !birthDateOK {
		addError("birth_date", locale.Text("loan.error.birth_date.required"))
	}

	if loanYearsOK {
		switch form.EmploymentType {
		case "contractor":
			if loanYears > 20 {
				addError("loan_years", locale.Text("loan.error.loan_years.contractor"))
			}
		case "self_employed":
			if loanYears > 25 {
				addError("loan_years", locale.Text("loan.error.loan_years.self_employed"))
			}
		}
	}

	if birthDateOK {
		age := yearsSince(birthDate, time.Now())
		if age < 18 {
			addError("birth_date", locale.Text("loan.error.birth_date.age"))
		}
		if loanYearsOK && age+loanYears > 75 {
			addError("loan_years", locale.Text("loan.error.loan_years.age"))
		}
	}

	if annualIncomeOK {
		state.NetMonthlyIncome = annualIncome / 12.0
		state.MaximumLoanAmount = annualIncome * employmentMultiplier(form.EmploymentType)
	}

	if annualIncomeOK && loanAmountOK && loanAmount > state.MaximumLoanAmount {
		addError("loan_amount", strings.NewReplacer("{max}", fmt.Sprintf("%.2f", state.MaximumLoanAmount)).Replace(locale.Text("loan.error.loan_amount.cap")))
	}

	if loanAmountOK && downPaymentOK {
		if downPayment > loanAmount {
			addError("down_payment", locale.Text("loan.error.down_payment.gt_loan"))
		} else if loanAmount > 0 {
			state.DownPaymentRatio = (downPayment / loanAmount) * 100
		}
	}

	if loanAmountOK && loanYearsOK {
		state.InterestRate = estimateInterestRate(form.EmploymentType, loanYears)
		state.TermMonths = loanYears * 12
		financedAmount := math.Max(loanAmount-downPayment, 0)
		state.EstimatedInstallment = calculateMonthlyInstallment(financedAmount, state.InterestRate, state.TermMonths)
	}

	if state.NetMonthlyIncome > 0 && monthlyDebtOK && state.EstimatedInstallment > 0 {
		state.DebtToIncomeRatio = ((monthlyDebt + state.EstimatedInstallment) / state.NetMonthlyIncome) * 100
		if state.DebtToIncomeRatio > 40 {
			addError("monthly_debt", "Debt-to-income ratio exceeds 40% after adding this loan.")
		} else if state.DebtToIncomeRatio > 35 {
			addWarning("Debt-to-income ratio is above 35%; affordability is borderline.")
		}
	}

	if loanAmountOK && downPaymentOK {
		if state.DownPaymentRatio < 10 {
			hasStrongCollateral := collateralValueOK && collateralValue >= loanAmount*1.2
			if !form.HasGuarantor && !hasStrongCollateral {
				addError("down_payment", locale.Text("loan.error.down_payment.ratio"))
			} else {
				addWarning(locale.Text("loan.warning.down_payment.support"))
			}
		}
	}

	if loanAmountOK && loanAmount > 250000 {
		hasCoverage := form.HasGuarantor || (collateralValueOK && collateralValue >= loanAmount*1.1)
		if !hasCoverage {
			addError("loan_amount", locale.Text("loan.error.loan_amount.coverage"))
		}
	}

	if form.HasGuarantor && guarantorIncomeOK && annualIncomeOK {
		ratio := guarantorIncome / annualIncome
		if ratio < 0.2 {
			addError("guarantor_income", locale.Text("loan.error.guarantor_income.ratio"))
		} else if ratio < 0.3 {
			addWarning(locale.Text("loan.warning.guarantor_income.ratio"))
		}
	}

	if form.Purpose == "business" && loanAmountOK && loanAmount > 150000 {
		addWarning(locale.Text("loan.warning.business.docs"))
	}

	state.IsValid = len(state.Errors) == 0
	return state
}

func parsePositiveFloat(raw string) (float64, bool) {
	value, err := strconv.ParseFloat(strings.ReplaceAll(raw, ",", "."), 64)
	if err != nil || value <= 0 {
		return 0, false
	}
	return value, true
}

func parseNonNegativeFloat(raw string) (float64, bool) {
	value, err := strconv.ParseFloat(strings.ReplaceAll(raw, ",", "."), 64)
	if err != nil || value < 0 {
		return 0, false
	}
	return value, true
}

func parsePositiveInt(raw string) (int, bool) {
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return 0, false
	}
	return value, true
}

func parseBirthDate(raw string) (time.Time, bool) {
	if raw == "" {
		return time.Time{}, false
	}
	value, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return time.Time{}, false
	}
	return value, true
}

func yearsSince(birthDate time.Time, now time.Time) int {
	years := now.Year() - birthDate.Year()
	if now.Month() < birthDate.Month() || (now.Month() == birthDate.Month() && now.Day() < birthDate.Day()) {
		years--
	}
	return years
}

func employmentMultiplier(employmentType string) float64 {
	switch employmentType {
	case "salaried":
		return 6.0
	case "self_employed":
		return 4.5
	case "contractor":
		return 3.5
	default:
		return 3.0
	}
}

func estimateInterestRate(employmentType string, loanYears int) float64 {
	rate := 5.2
	switch employmentType {
	case "salaried":
		rate = 4.5
	case "self_employed":
		rate = 5.4
	case "contractor":
		rate = 5.9
	}
	if loanYears > 20 {
		rate += 0.35
	}
	return rate
}

func calculateMonthlyInstallment(principal float64, annualRate float64, termMonths int) float64 {
	if principal <= 0 || termMonths <= 0 {
		return 0
	}

	monthlyRate := (annualRate / 100.0) / 12.0
	if monthlyRate == 0 {
		return principal / float64(termMonths)
	}

	factor := math.Pow(1+monthlyRate, float64(termMonths))
	return principal * ((monthlyRate * factor) / (factor - 1))
}

func nonSpaceChars(value string) int {
	count := 0
	for _, r := range value {
		if !unicode.IsSpace(r) {
			count++
		}
	}
	return count
}
