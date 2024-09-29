package trading212

import (
	"fmt"
	"math"
	"strconv"

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
func (r *Record) GetActualPriceForQuantity(quantity float64, buy bool) (float64, error) {
	var err error
	if r.NoOfShares < quantity {
		return 0, merry.Errorf("quantity value is more than available shares: Requested: %f Available: %f",
			quantity, r.NoOfShares)
	}
	proportionalConversionFee := r.CurrencyConversionFee * quantity / r.NoOfShares
	total := (quantity * r.PriceShare / r.ExchangeRate)
	if buy {
		// when selling, this fee is added to get the Total value (idk why)
		total += proportionalConversionFee
	} else {
		// when selling, this fee is subtracted to get the Total value (idk why)
		total -= proportionalConversionFee
	}

	// adjust record data
	r.CurrencyConversionFee -= proportionalConversionFee
	r.NoOfShares -= quantity
	r.Total -= total

	r.Total, err = RoundFloatFast(r.Total, 2)
	if err != nil {
		return 0, merry.Errorf("failed to adjust float precision: %w", err)
	}
	r.NoOfShares, err = RoundFloatFast(r.NoOfShares, 2)
	if err != nil {
		return 0, merry.Errorf("failed to adjust float precision: %w", err)
	}
	r.CurrencyConversionFee, err = RoundFloatFast(r.CurrencyConversionFee, 2)
	if err != nil {
		return 0, merry.Errorf("failed to adjust float precision: %w", err)
	}

	return total, nil
}

func RoundFloatFast(f float64, prec int) (float64, error) {
	// https: //stackoverflow.com/questions/18390266/how-can-we-truncate-float64-type-to-a-particular-precision
	mul := math.Pow10(prec)
	if mul == 0 {
		return 0, nil
	}

	product := f * mul
	var roundingErr error
	if product > float64(math.MaxInt64) {
		roundingErr = fmt.Errorf("unsafe round: float64=%+v, places=%d", f, prec)
	}

	return math.Round(product) / mul, roundingErr
}

func (r *Record) GetYear() (int, error) {
	return strconv.Atoi(r.Time[0:4])
}
