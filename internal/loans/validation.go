package loans

import (
	"math/big"
	"strconv"
	"strings"
	"time"

	appformat "github.com/marcello/saas-poc/internal/format"
	apptext "github.com/marcello/saas-poc/internal/text"
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
	NetMonthlyIncome     int64
	EstimatedInstallment int64
	DebtToIncomeRatio    float64
	MaximumLoanAmount    int64
	DownPaymentRatio     float64
	InterestRate         int64
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
	} else if apptext.CountNonSpaceChars(form.FullName) < 3 {
		addError("full_name", messages.Text("loan.error.full_name.length"))
	}

	if form.EmploymentType == "" {
		addError("employment_type", messages.Text("loan.error.employment_type.required"))
	}

	annualIncome, annualIncomeOK := parsePositiveCents(form.AnnualIncome)
	if !annualIncomeOK {
		addError("annual_income", messages.Text("loan.error.annual_income"))
	}

	monthlyDebt, monthlyDebtOK := parseNonNegativeCents(form.MonthlyDebt)
	if !monthlyDebtOK {
		addError("monthly_debt", messages.Text("loan.error.monthly_debt"))
	}

	loanAmount, loanAmountOK := parsePositiveCents(form.LoanAmount)
	if !loanAmountOK {
		addError("loan_amount", messages.Text("loan.error.loan_amount.positive"))
	} else if loanAmount <= 10000 {
		addError("loan_amount", messages.Text("loan.error.loan_amount.minimum"))
	}

	loanYears, loanYearsOK := parsePositiveInt(form.LoanYears)
	if !loanYearsOK {
		addError("loan_years", messages.Text("loan.error.loan_years"))
	}

	downPayment, downPaymentOK := parseNonNegativeCents(form.DownPayment)
	if !downPaymentOK {
		addError("down_payment", messages.Text("loan.error.down_payment"))
	}

	collateralValue := int64(0)
	collateralValueOK := true
	if form.CollateralValue != "" {
		collateralValue, collateralValueOK = parseNonNegativeCents(form.CollateralValue)
		if !collateralValueOK {
			addError("collateral_value", messages.Text("loan.error.collateral_value"))
		}
	}

	guarantorIncome := int64(0)
	guarantorIncomeOK := true
	if form.HasGuarantor {
		if form.GuarantorIncome == "" {
			addError("guarantor_income", messages.Text("loan.error.guarantor_income.required"))
			guarantorIncomeOK = false
		} else {
			guarantorIncome, guarantorIncomeOK = parsePositiveCents(form.GuarantorIncome)
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
		state.NetMonthlyIncome = annualIncome / 12
		state.MaximumLoanAmount = employmentMultiplier(annualIncome, form.EmploymentType)
	}

	if annualIncomeOK && loanAmountOK && loanAmount > state.MaximumLoanAmount {
		addError("loan_amount", strings.NewReplacer("{max}", appformat.FormatCents(state.MaximumLoanAmount)).Replace(messages.Text("loan.error.loan_amount.cap")))
	}

	if loanAmountOK && downPaymentOK {
		if downPayment > loanAmount {
			addError("down_payment", messages.Text("loan.error.down_payment.gt_loan"))
		} else if loanAmount > 0 {
			state.DownPaymentRatio = (float64(downPayment) / float64(loanAmount)) * 100
		}
	}

	if loanAmountOK && loanYearsOK {
		state.InterestRate = estimateInterestRate(form.EmploymentType, loanYears)
		state.TermMonths = loanYears * 12
		financedAmount := loanAmount - downPayment
		if financedAmount < 0 {
			financedAmount = 0
		}
		state.EstimatedInstallment = calculateMonthlyInstallment(financedAmount, state.InterestRate, state.TermMonths)
	}

	if state.NetMonthlyIncome > 0 && monthlyDebtOK && state.EstimatedInstallment > 0 {
		state.DebtToIncomeRatio = (float64(monthlyDebt+state.EstimatedInstallment) / float64(state.NetMonthlyIncome)) * 100
		if state.DebtToIncomeRatio > 40 {
			addError("monthly_debt", "Debt-to-income ratio exceeds 40% after adding this loan.")
		} else if state.DebtToIncomeRatio > 35 {
			addWarning("Debt-to-income ratio is above 35%; affordability is borderline.")
		}
	}

	if loanAmountOK && downPaymentOK {
		if state.DownPaymentRatio < 10 {
			hasStrongCollateral := collateralValueOK && collateralValue >= (loanAmount*12)/10
			if !form.HasGuarantor && !hasStrongCollateral {
				addError("down_payment", messages.Text("loan.error.down_payment.ratio"))
			} else {
				addWarning(messages.Text("loan.warning.down_payment.support"))
			}
		}
	}

	if loanAmountOK && loanAmount > 25000000 {
		hasCoverage := form.HasGuarantor || (collateralValueOK && collateralValue >= (loanAmount*11)/10)
		if !hasCoverage {
			addError("loan_amount", messages.Text("loan.error.loan_amount.coverage"))
		}
	}

	if form.HasGuarantor && guarantorIncomeOK && annualIncomeOK {
		ratio := float64(guarantorIncome) / float64(annualIncome)
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

func parsePositiveCents(raw string) (int64, bool) {
	value, ok := parseCents(raw)
	if !ok || value <= 0 {
		return 0, false
	}
	return value, true
}

func parseNonNegativeCents(raw string) (int64, bool) {
	value, ok := parseCents(raw)
	if !ok || value < 0 {
		return 0, false
	}
	return value, true
}

func parseCents(raw string) (int64, bool) {
	raw = strings.TrimSpace(strings.ReplaceAll(raw, ",", "."))
	parts := strings.Split(raw, ".")
	if len(parts) > 2 || raw == "" || (len(parts) == 2 && len(parts[1]) > 2) {
		return 0, false
	}
	if len(parts) == 2 && parts[0] == "" && parts[1] == "" {
		return 0, false
	}
	whole := parts[0]
	if whole == "" {
		whole = "0"
	}
	fraction := ""
	if len(parts) == 2 {
		fraction = parts[1]
	}
	for len(fraction) < 2 {
		fraction += "0"
	}
	wholeValue, err := strconv.ParseInt(whole, 10, 64)
	if err != nil || wholeValue < 0 {
		return 0, false
	}
	fractionValue, err := strconv.ParseInt(fraction, 10, 64)
	if err != nil {
		return 0, false
	}
	if wholeValue > (int64(^uint64(0)>>1)-fractionValue)/100 {
		return 0, false
	}
	return wholeValue*100 + fractionValue, true
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

func employmentMultiplier(annualIncome int64, employmentType string) int64 {
	multiplierNumerator := int64(3)
	multiplierDenominator := int64(1)
	switch employmentType {
	case "salaried":
		multiplierNumerator = 6
	case "self_employed":
		multiplierNumerator = 9
		multiplierDenominator = 2
	case "contractor":
		multiplierNumerator = 7
		multiplierDenominator = 2
	}
	return (annualIncome * multiplierNumerator) / multiplierDenominator
}

func estimateInterestRate(employmentType string, loanYears int) int64 {
	rate := int64(520)
	switch employmentType {
	case "salaried":
		rate = 450
	case "self_employed":
		rate = 540
	case "contractor":
		rate = 590
	}
	if loanYears > 20 {
		rate += 35
	}
	return rate
}

func calculateMonthlyInstallment(principal int64, annualRate int64, termMonths int) int64 {
	if principal <= 0 || termMonths <= 0 {
		return 0
	}

	monthlyRate := new(big.Rat).SetFrac(big.NewInt(annualRate), big.NewInt(120000))
	if monthlyRate.Sign() == 0 {
		return principal / int64(termMonths)
	}

	factor := new(big.Rat).Add(big.NewRat(1, 1), monthlyRate)
	base := new(big.Rat).Set(factor)
	for month := 1; month < termMonths; month++ {
		factor.Mul(factor, base)
	}
	installment := new(big.Rat).Mul(monthlyRate, factor)
	installment.Quo(installment, new(big.Rat).Sub(factor, big.NewRat(1, 1)))
	installment.Mul(installment, big.NewRat(principal, 1))
	quotient, remainder := new(big.Int).QuoRem(installment.Num(), installment.Denom(), new(big.Int))
	if new(big.Int).Mul(remainder, big.NewInt(2)).Cmp(installment.Denom()) >= 0 {
		quotient.Add(quotient, big.NewInt(1))
	}
	return quotient.Int64()
}
