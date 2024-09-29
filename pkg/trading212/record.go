package trading212

import (
	"github.com/ansel1/merry/v2"
)

type RecordDTO struct {
	Action                        string `json:"Action"`
	Time                          string `json:"Time"`
	Isin                          string `json:"ISIN"`
	Ticker                        string `json:"Ticker"`
	Name                          string `json:"Name"`
	NoOfShares                    string `json:"No. of shares"`
	PriceShare                    string `json:"Price / share"`
	CurrencyPriceShare            string `json:"Currency (Price / share)"`
	ExchangeRate                  string `json:"Exchange rate"`
	Result                        string `json:"Result"`
	CurrencyResult                string `json:"Currency (Result)"`
	Total                         string `json:"Total"`
	CurrencyTotal                 string `json:"Currency (Total)"`
	WithholdingTax                string `json:"Withholding tax"`
	CurrencyWithholdingTax        string `json:"Currency (Withholding tax)"`
	StampDutyReserveTax           string `json:"Stamp duty reserve tax"`
	CurrencyStampDutyReserveTax   string `json:"Currency (Stamp duty reserve tax)"`
	Notes                         string `json:"Notes"`
	ID                            string `json:"ID"`
	CurrencyConversionFee         string `json:"Currency conversion fee"`
	CurrencyCurrencyConversionFee string `json:"Currency (Currency conversion fee)"`
}

type Record struct {
	Action                        string  `json:"Action"`
	Time                          string  `json:"Time"`
	Isin                          string  `json:"ISIN"`
	Ticker                        string  `json:"Ticker"`
	Name                          string  `json:"Name"`
	NoOfShares                    float64 `json:"No. of shares"`
	PriceShare                    float64 `json:"Price / share"`
	CurrencyPriceShare            string  `json:"Currency (Price / share)"`
	ExchangeRate                  float64 `json:"Exchange rate"`
	Result                        float64 `json:"Result"`
	CurrencyResult                string  `json:"Currency (Result)"`
	Total                         float64 `json:"Total"`
	CurrencyTotal                 string  `json:"Currency (Total)"`
	WithholdingTax                float64 `json:"Withholding tax"`
	CurrencyWithholdingTax        string  `json:"Currency (Withholding tax)"`
	StampDutyReserveTax           float64 `json:"Stamp duty reserve tax"`
	CurrencyStampDutyReserveTax   string  `json:"Currency (Stamp duty reserve tax)"`
	Notes                         string  `json:"Notes"`
	ID                            string  `json:"ID"`
	CurrencyConversionFee         float64 `json:"Currency conversion fee"`
	CurrencyCurrencyConversionFee string  `json:"Currency (Currency conversion fee)"`
}

// (floatQuantity * floatPriceShare / floatExchangeRate) - floatCurrencyConversionFee
func (r *Record) GetActualPriceForQuantity(quantity float64) (float64, error) {
	if r.NoOfShares < quantity {
		return 0, merry.Errorf("quantity value is more than available shares: Requested: %f Available: %f",
			quantity, r.NoOfShares)
	}
	proportionalConversionFee := r.CurrencyConversionFee * quantity / r.NoOfShares
	total := (quantity * r.PriceShare / r.ExchangeRate) + proportionalConversionFee

	// adjust record data
	r.CurrencyConversionFee -= proportionalConversionFee
	r.NoOfShares -= quantity
	r.Total -= total

	return total, nil
}
