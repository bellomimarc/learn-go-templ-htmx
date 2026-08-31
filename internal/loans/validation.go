package loans

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode"
)

type Messages interface {
	Text(key string) string
}

// FormData stores the raw form payload so user input can be re-rendered unchanged.
type FormData struct {
	FullName        string
	BirthDate       string
	EmploymentType  string
	AnnualIncome    string
	MonthlyDebt     string
	LoanAmount      string
	LoanYears       string
	DownPayment     string
	CollateralValue string
	HasGuarantor    bool
	GuarantorIncome string
	Purpose         string
}

// ValidationIssue links a single message to a specific form field.
type ValidationIssue struct {
	Field   string
	Message string
}

// ValidationState contains both validation results and derived financial metrics.
type ValidationState struct {
	HasEvaluated         bool
	Errors               []ValidationIssue
	Warnings             []string
	NetMonthlyIncome     float64
	EstimatedInstallment float64
	DebtToIncomeRatio    float64
	MaximumLoanAmount    float64
	DownPaymentRatio     float64
	InterestRate         float64
	TermMonths           int
	IsValid              bool
}

func Validate(form FormData, hasEvaluated bool, messages Messages) ValidationState {
	state := ValidationState{HasEvaluated: hasEvaluated}
	if !hasEvaluated {
		return state
	}

	addError := func(field, msg string) {
		state.Errors = append(state.Errors, ValidationIssue{Field: field, Message: msg})
	}
	addWarning := func(msg string) {
		state.Warnings = append(state.Warnings, msg)
	}

	if form.FullName == "" {
		addError("full_name", messages.Text("loan.error.full_name.required"))
	} else if nonSpaceChars(form.FullName) < 3 {
		addError("full_name", messages.Text("loan.error.full_name.length"))
	}

	if form.EmploymentType == "" {
		addError("employment_type", messages.Text("loan.error.employment_type.required"))
	}

	annualIncome, annualIncomeOK := parsePositiveFloat(form.AnnualIncome)
	if !annualIncomeOK {
		addError("annual_income", messages.Text("loan.error.annual_income"))
	}

	monthlyDebt, monthlyDebtOK := parseNonNegativeFloat(form.MonthlyDebt)
	if !monthlyDebtOK {
		addError("monthly_debt", messages.Text("loan.error.monthly_debt"))
	}

	loanAmount, loanAmountOK := parsePositiveFloat(form.LoanAmount)
	if !loanAmountOK {
		addError("loan_amount", messages.Text("loan.error.loan_amount.positive"))
	} else if loanAmount <= 100 {
		addError("loan_amount", messages.Text("loan.error.loan_amount.minimum"))
	}

	loanYears, loanYearsOK := parsePositiveInt(form.LoanYears)
	if !loanYearsOK {
		addError("loan_years", messages.Text("loan.error.loan_years"))
	}

	downPayment, downPaymentOK := parseNonNegativeFloat(form.DownPayment)
	if !downPaymentOK {
		addError("down_payment", messages.Text("loan.error.down_payment"))
	}

	collateralValue := 0.0
	collateralValueOK := true
	if form.CollateralValue != "" {
		collateralValue, collateralValueOK = parseNonNegativeFloat(form.CollateralValue)
		if !collateralValueOK {
			addError("collateral_value", messages.Text("loan.error.collateral_value"))
		}
	}

	guarantorIncome := 0.0
	guarantorIncomeOK := true
	if form.HasGuarantor {
		if form.GuarantorIncome == "" {
			addError("guarantor_income", messages.Text("loan.error.guarantor_income.required"))
			guarantorIncomeOK = false
		} else {
			guarantorIncome, guarantorIncomeOK = parsePositiveFloat(form.GuarantorIncome)
			if !guarantorIncomeOK {
				addError("guarantor_income", messages.Text("loan.error.guarantor_income.positive"))
			}
		}
	} else if form.GuarantorIncome != "" {
		addWarning(messages.Text("loan.warning.guarantor_income.unused"))
	}

	birthDate, birthDateOK := parseBirthDate(form.BirthDate)
	if !birthDateOK {
		addError("birth_date", messages.Text("loan.error.birth_date.required"))
	}

	if loanYearsOK {
		switch form.EmploymentType {
		case "contractor":
			if loanYears > 20 {
				addError("loan_years", messages.Text("loan.error.loan_years.contractor"))
			}
		case "self_employed":
			if loanYears > 25 {
				addError("loan_years", messages.Text("loan.error.loan_years.self_employed"))
			}
		}
	}

	if birthDateOK {
		age := yearsSince(birthDate, time.Now())
		if age < 18 {
			addError("birth_date", messages.Text("loan.error.birth_date.age"))
		}
		if loanYearsOK && age+loanYears > 75 {
			addError("loan_years", messages.Text("loan.error.loan_years.age"))
		}
	}

	if annualIncomeOK {
		state.NetMonthlyIncome = annualIncome / 12.0
		state.MaximumLoanAmount = annualIncome * employmentMultiplier(form.EmploymentType)
	}

	if annualIncomeOK && loanAmountOK && loanAmount > state.MaximumLoanAmount {
		addError("loan_amount", strings.NewReplacer("{max}", fmt.Sprintf("%.2f", state.MaximumLoanAmount)).Replace(messages.Text("loan.error.loan_amount.cap")))
	}

	if loanAmountOK && downPaymentOK {
		if downPayment > loanAmount {
			addError("down_payment", messages.Text("loan.error.down_payment.gt_loan"))
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
				addError("down_payment", messages.Text("loan.error.down_payment.ratio"))
			} else {
				addWarning(messages.Text("loan.warning.down_payment.support"))
			}
		}
	}

	if loanAmountOK && loanAmount > 250000 {
		hasCoverage := form.HasGuarantor || (collateralValueOK && collateralValue >= loanAmount*1.1)
		if !hasCoverage {
			addError("loan_amount", messages.Text("loan.error.loan_amount.coverage"))
		}
	}

	if form.HasGuarantor && guarantorIncomeOK && annualIncomeOK {
		ratio := guarantorIncome / annualIncome
		if ratio < 0.2 {
			addError("guarantor_income", messages.Text("loan.error.guarantor_income.ratio"))
		} else if ratio < 0.3 {
			addWarning(messages.Text("loan.warning.guarantor_income.ratio"))
		}
	}

	if form.Purpose == "business" && loanAmountOK && loanAmount > 150000 {
		addWarning(messages.Text("loan.warning.business.docs"))
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
