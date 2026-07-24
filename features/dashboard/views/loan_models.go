package dashboard

// LoanFormData stores the raw form payload so user input can be re-rendered unchanged.
type LoanFormData struct {
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

// LoanValidationState contains both validation results and derived financial metrics.
type LoanValidationState struct {
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
