package model

type Budget struct {
	Name       string        `json:"name"`
	DisplayName string       `json:"displayName,omitempty"`
	Amount     *BudgetAmount `json:"amount"`
	ThresholdRules []*ThresholdRule `json:"thresholdRules,omitempty"`
}

type BudgetAmount struct {
	SpecifiedAmount *Money `json:"specifiedAmount"`
}

type Money struct {
	Units        int64  `json:"units,string"`
	Nanos        int32  `json:"nanos"`
	CurrencyCode string `json:"currencyCode"`
}

type ThresholdRule struct {
	ThresholdPercent float64 `json:"thresholdPercent"`
	SpendBasis       string  `json:"spendBasis,omitempty"`
}
