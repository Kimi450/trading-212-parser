package trading212

import (
	"slices"
	"strconv"

	"github.com/ansel1/merry/v2"
	"github.com/shopspring/decimal"
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
	Action                        string          `json:"Action"`
	Time                          string          `json:"Time"`
	Isin                          string          `json:"ISIN"`
	Ticker                        string          `json:"Ticker"`
	Name                          string          `json:"Name"`
	NoOfShares                    decimal.Decimal `json:"No. of shares"`
	PriceShare                    decimal.Decimal `json:"Price / share"`
	CurrencyPriceShare            string          `json:"Currency (Price / share)"`
	ExchangeRate                  decimal.Decimal `json:"Exchange rate"`
	Result                        decimal.Decimal `json:"Result"`
	CurrencyResult                string          `json:"Currency (Result)"`
	Total                         decimal.Decimal `json:"Total"`
	CurrencyTotal                 string          `json:"Currency (Total)"`
	WithholdingTax                decimal.Decimal `json:"Withholding tax"`
	CurrencyWithholdingTax        string          `json:"Currency (Withholding tax)"`
	StampDutyReserveTax           decimal.Decimal `json:"Stamp duty reserve tax"`
	CurrencyStampDutyReserveTax   string          `json:"Currency (Stamp duty reserve tax)"`
	Notes                         string          `json:"Notes"`
	ID                            string          `json:"ID"`
	CurrencyConversionFee         decimal.Decimal `json:"Currency conversion fee"`
	CurrencyCurrencyConversionFee string          `json:"Currency (Currency conversion fee)"`
}

type RecordType string

const (
	ETF   RecordType = "ETF"
	Stock RecordType = "Stock"
)

// (floatQuantity * floatPriceShare / floatExchangeRate) - floatCurrencyConversionFee
func (r *Record) GetActualPriceForQuantity(quantity decimal.Decimal, buy bool) (decimal.Decimal, error) {
	if r.NoOfShares.LessThan(quantity) {
		return decimal.NewFromInt(0),
			merry.Errorf("quantity value is more than available shares: Requested: %f Available: %f",
				quantity, r.NoOfShares)
	}
	proportionalConversionFee := r.CurrencyConversionFee.Mul(quantity).Div(r.NoOfShares)
	total := quantity.Mul(r.PriceShare).Div(r.ExchangeRate)
	if buy {
		// when selling, this fee is added to get the Total value
		// This is because the total you get is after the fee is added to
		// it representing the total cost to you
		total = total.Add(proportionalConversionFee)
	} else {
		// when selling, this fee is subtracted to get the Total value
		// This is because the total you get is after the fee is taken from it
		// to show how much you got from it (net, i.e, after the fee)
		total = total.Sub(proportionalConversionFee)
	}

	// adjust record data
	r.CurrencyConversionFee = r.CurrencyConversionFee.Sub(proportionalConversionFee)
	r.NoOfShares = r.NoOfShares.Sub(quantity)
	r.Total = r.Total.Sub(total)

	return total, nil
}

func (r *Record) GetYear() (int, error) {
	return strconv.Atoi(r.Time[0:4])
}

// I have a support ticket with Trading 212 to add this data
// to the transaction history export
func (r *Record) GetType() RecordType {
	etf := []string{"VUSA", "VUAA"}

	if slices.Contains(etf, r.Ticker) {
		return ETF
	}
	return Stock
}
