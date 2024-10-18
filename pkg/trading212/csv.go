package trading212

// https://stackoverflow.com/questions/24999079/reading-csv-file-in-go

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	"github.com/ansel1/merry/v2"
)

type Scanner struct {
	Reader *csv.Reader
	Head   []string
	Row    []string
}

func NewScanner(o io.Reader) Scanner {
	csv_o := csv.NewReader(o)
	header, err := csv_o.Read()
	if err != nil {
		return Scanner{}
	}
	return Scanner{Reader: csv_o, Head: header}
}

func (o *Scanner) Scan() bool {
	row, err := o.Reader.Read()
	if err != nil {
		return false
	}
	o.Row = row
	return true
}

func (o *Scanner) toJson() string {
	json := ""
	for i, value := range o.Row {
		value = strings.TrimSpace(value)
		if i != 0 {
			json += ",\n"
		}
		json += "\"" + o.Head[i] + "\":\"" + value + "\""
	}
	json = "{\n" + json + "\n}"
	return json
}

func (o *Scanner) ToRecord() (Record, error) {
	record := Record{}

	jsonString := []byte(o.toJson())
	recordDto := RecordDTO{}
	err := json.Unmarshal(jsonString, &recordDto)
	if err != nil {
		return record, merry.Errorf("failed to unmarshall json to record: %w", err)
	}

	// cannot find a cleaner and simpler way to do this
	record.Action = recordDto.Action

	parsedTime, err := time.Parse("2006-01-02 15:04:05", recordDto.Time)
	if err != nil {
		return record, merry.Errorf("failed to parse time: %w", err)
	}
	record.Time = parsedTime
	record.Isin = recordDto.Isin
	record.Ticker = recordDto.Ticker
	record.Name = recordDto.Name
	record.CurrencyPriceShare = recordDto.CurrencyPriceShare
	record.CurrencyResult = recordDto.CurrencyResult
	record.CurrencyTotal = recordDto.CurrencyTotal
	record.CurrencyWithholdingTax = recordDto.CurrencyWithholdingTax
	record.CurrencyStampDutyReserveTax = recordDto.CurrencyStampDutyReserveTax
	record.Notes = recordDto.Notes
	record.ID = recordDto.ID
	record.CurrencyCurrencyConversionFee = recordDto.CurrencyCurrencyConversionFee

	parseFloatIgnoreEmptyString := func(value string) (decimal.Decimal, error) {
		var out decimal.Decimal
		if regexp.MustCompile(`^\d+(\.\d+)?$`).MatchString(value) {
			out, err = decimal.NewFromString(value)
			if err != nil {
				return out, merry.Errorf("failed to parse float: %w", err)
			}
		}
		return out, nil
	}

	record.NoOfShares, err = parseFloatIgnoreEmptyString(recordDto.NoOfShares)
	if err != nil {
		return record, merry.Errorf("failed to parse float for record DTO 'NoOfShares': %w", err)
	}
	record.PriceShare, err = parseFloatIgnoreEmptyString(recordDto.PriceShare)
	if err != nil {
		return record, merry.Errorf("failed to parse float for record DTO 'PriceShare': %w", err)
	}
	record.ExchangeRate, err = parseFloatIgnoreEmptyString(recordDto.ExchangeRate)
	if err != nil {
		return record, merry.Errorf("failed to parse float for record DTO 'ExchangeRate': %w", err)
	}
	record.Result, err = parseFloatIgnoreEmptyString(recordDto.Result)
	if err != nil {
		return record, merry.Errorf("failed to parse float for record DTO 'Result': %w", err)
	}
	record.Total, err = parseFloatIgnoreEmptyString(recordDto.Total)
	if err != nil {
		return record, merry.Errorf("failed to parse float for record DTO 'Total': %w", err)
	}
	record.WithholdingTax, err = parseFloatIgnoreEmptyString(recordDto.WithholdingTax)
	if err != nil {
		return record, merry.Errorf("failed to parse float for record DTO 'WithholdingTax': %w", err)
	}
	record.StampDutyReserveTax, err = parseFloatIgnoreEmptyString(recordDto.StampDutyReserveTax)
	if err != nil {
		return record, merry.Errorf("failed to parse float for record DTO 'StampDutyReserveTax': %w", err)
	}
	record.CurrencyConversionFee, err = parseFloatIgnoreEmptyString(recordDto.CurrencyConversionFee)
	if err != nil {
		return record, merry.Errorf("failed to parse float for record DTO 'CurrencyConversionFee': %w", err)
	}
	err = record.AdjustForSplit()
	if err != nil {
		return record, merry.Errorf("failed to adjust for split: %w", err)
	}

	return record, nil
}
