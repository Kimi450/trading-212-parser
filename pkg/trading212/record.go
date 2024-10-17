package trading212

import (
	"slices"
	"strconv"
	"time"

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

type SplitAdjusted struct {
	Done bool
}

type Record struct {
	SplitAdjusted

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

func (r *Record) AdjustForSplit() error {
	_, denominator, err := SplitAdjustmentRequired(*r)
	if err != nil {
		return err
	}

	if denominator == 1 { // no adjustment needed
		return nil
	}

	r.NoOfShares = r.NoOfShares.Mul(decimal.NewFromInt(denominator))
	r.PriceShare = r.PriceShare.Div(decimal.NewFromInt(denominator))
	r.SplitAdjusted.Done = true

	return nil
}

func SplitAdjustmentRequired(record Record) (numerator, dedenmoniator int64, err error) {
	parsedTime, err := time.Parse("2006-01-02 15:04:05", record.Time[:19])
	if err != nil {
		return 0, 0, err
	}

	// https://companiesmarketcap.com/eur

	if record.Ticker == "AAPL" {
		// all good for post 2022
		return 1, 1, nil
	}
	if record.Ticker == "AMD" {
		// all good for post 2022
		return 1, 1, nil
	}
	if record.Ticker == "AMZN" &&
		parsedTime.Before(time.Date(2022, 6, 6, 0, 0, 0, 0, time.UTC)) &&
		parsedTime.After(time.Date(1999, 9, 2, 0, 0, 0, 0, time.UTC)) {
		return 1, 20, nil
	}
	if record.Ticker == "BA" {
		// all good for post 2022
		return 1, 1, nil
	}
	if record.Ticker == "BABA" {
		// all good for post 2022
		return 1, 1, nil
	}
	if record.Ticker == "BARC" {
		// all good for post 2022
		return 1, 1, nil
	}
	if record.Ticker == "BRT3" {
		// ETF
		return 1, 1, nil
	}
	if record.Ticker == "CRUDP" {
		// idk
		return 1, 1, nil
	}
	if record.Ticker == "CRWD" {
		// all good for post 2022
		return 1, 1, nil
	}
	if record.Ticker == "DIS" {
		// all good for post 2022
		return 1, 1, nil
	}
	if record.Ticker == "GME" &&
		parsedTime.Before(time.Date(2022, 7, 22, 0, 0, 0, 0, time.UTC)) &&
		parsedTime.After(time.Date(2007, 3, 19, 0, 0, 0, 0, time.UTC)) {
		return 1, 4, nil
	}
	if record.Ticker == "GOOGL" &&
		parsedTime.Before(time.Date(2022, 7, 18, 0, 0, 0, 0, time.UTC)) &&
		parsedTime.After(time.Date(2015, 4, 27, 0, 0, 0, 0, time.UTC)) {
		return 1, 20, nil
	}
	if record.Ticker == "INTC" {
		// all good for post 2022
		return 1, 1, nil
	}
	if record.Ticker == "LUNR" {
		// all good for post 2022
		return 1, 1, nil
	}
	if record.Ticker == "META" {
		// all good for post 2022
		return 1, 1, nil
	}
	if record.Ticker == "MRNA" {
		// all good for post 2022
		return 1, 1, nil
	}
	if record.Ticker == "MSFT" {
		// all good for post 2022
		return 1, 1, nil
	}
	if record.Ticker == "NFLX" {
		// all good for post 2022
		return 1, 1, nil
	}
	if record.Ticker == "NKE" {
		// all good for post 2022
		return 1, 1, nil
	}
	if record.Ticker == "NVDA" &&
		parsedTime.Before(time.Date(2024, 6, 10, 0, 0, 0, 0, time.UTC)) &&
		parsedTime.After(time.Date(2021, 7, 20, 0, 0, 0, 0, time.UTC)) {
		return 1, 10, nil
	}
	if record.Ticker == "RACE" {
		// all good for post 2022
		return 1, 1, nil
	}
	if record.Ticker == "RIGD" {
		// all good for post 2022
		return 1, 1, nil
	}
	if record.Ticker == "RR" {
		// all good for post 2022
		return 1, 1, nil
	}
	if record.Ticker == "RS" {
		// all good for post 2022
		return 1, 1, nil
	}
	if record.Ticker == "SPCE" &&
		parsedTime.Before(time.Date(2024, 6, 17, 0, 0, 0, 0, time.UTC)) {
		// all good for post 2022
		return 1, 20, nil
	}
	if record.Ticker == "SPOT" {
		// all good for post 2022
		return 1, 1, nil
	}
	if record.Ticker == "TSLA" &&
		parsedTime.Before(time.Date(2022, 8, 25, 0, 0, 0, 0, time.UTC)) &&
		parsedTime.After(time.Date(2020, 8, 31, 0, 0, 0, 0, time.UTC)) {
		return 1, 3, nil
	}
	if record.Ticker == "TWTR" {
		// all good for post 2022
		return 1, 1, nil
	}
	if record.Ticker == "VUAA" {
		// ETF
		return 1, 1, nil
	}
	if record.Ticker == "VUSA" {
		// ETF
		return 1, 1, nil
	}

	// Unknown
	return 1, 1, nil
}
