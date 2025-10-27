package trading212

import (
	"fmt"
	"strings"
	"time"

	"github.com/ansel1/merry/v2"
	"github.com/go-logr/logr"
	"github.com/shopspring/decimal"
)

type PurchaseHistory interface {
	GetRecordQueue() RecordQueue
	Process(log logr.Logger, newRecord *Record) error
	GetProfitForYear(year int) StockSummary
	GetSaleAggregatesForYear(year int) StockSummary
	GetLossAggregatesForYear(year int) StockSummary
	GetProfitAggregatesForYear(year int) StockSummary
}

type PurchaseHistoryStruct struct {
	recordQueue      RecordQueue
	profits          map[int]StockSummary
	saleAggregates   map[int]StockSummary
	lossAggregates   map[int]StockSummary
	profitAggregates map[int]StockSummary
}

func NewPurchaseHistory(recordQueue RecordQueue) PurchaseHistory {
	return &PurchaseHistoryStruct{
		recordQueue:      recordQueue,
		profits:          make(map[int]StockSummary),
		saleAggregates:   make(map[int]StockSummary),
		lossAggregates:   make(map[int]StockSummary),
		profitAggregates: make(map[int]StockSummary),
	}
}

func (q *PurchaseHistoryStruct) GetRecordQueue() RecordQueue {
	return q.recordQueue
}

func (q *PurchaseHistoryStruct) GetProfitForYear(year int) StockSummary {
	return q.profits[year]
}

func (q *PurchaseHistoryStruct) GetSaleAggregatesForYear(year int) StockSummary {
	return q.saleAggregates[year]
}

func (q *PurchaseHistoryStruct) GetLossAggregatesForYear(year int) StockSummary {
	return q.lossAggregates[year]
}

func (q *PurchaseHistoryStruct) GetProfitAggregatesForYear(year int) StockSummary {
	return q.profitAggregates[year]
}

func (q *PurchaseHistoryStruct) Process(log logr.Logger, newRecord *Record) error {
	if !strings.Contains(newRecord.Action, "buy") && !strings.Contains(newRecord.Action, "sell") {
		return nil
	}

	log.V(0).Info(fmt.Sprintf("%-12s", newRecord.Action),
		"ticker", fmt.Sprintf("%-5s", newRecord.Ticker),
		"date", newRecord.Time.String(),
		"PriceShare", fmt.Sprintf("%7s", newRecord.PriceShare.StringFixed(2)),
		"NoOfShares", fmt.Sprintf("%6s", newRecord.NoOfShares.StringFixed(2)),
		"splitadjusted", fmt.Sprintf("%-5t", newRecord.SplitAdjusted.Done),
	)
	if strings.Contains(newRecord.Action, "buy") {
		q.recordQueue.Append(newRecord)
	} else if strings.Contains(newRecord.Action, "sell") {
		year := newRecord.GetYear()
		sellPrice, profit, err := q.updateHistoryAndGetProfit(log, *newRecord)
		if err != nil {
			return merry.Errorf("failed to process new record: %w", err)
		}

		existingYearProfit := q.profits[year]
		existingYearSaleAggregate := q.saleAggregates[year]
		existingYearLossAggregate := q.lossAggregates[year]
		existingYearProfitAggregate := q.profitAggregates[year]

		newRecordType := newRecord.GetType()
		switch newRecordType {
		case Stock:
			existingYearProfit.Stock = existingYearProfit.Stock.Add(profit)
			q.profits[year] = existingYearProfit
			existingYearSaleAggregate.Stock = existingYearSaleAggregate.Stock.Add(sellPrice)
			q.saleAggregates[year] = existingYearSaleAggregate
			if profit.LessThan(decimal.NewFromInt(0)) {
				existingYearLossAggregate.Stock = existingYearLossAggregate.Stock.Add(profit)
				q.lossAggregates[year] = existingYearLossAggregate
			} else {
				existingYearProfitAggregate.Stock = existingYearProfitAggregate.Stock.Add(profit)
				q.profitAggregates[year] = existingYearProfitAggregate
			}
		case ETF:
			existingYearProfit.ETF = existingYearProfit.ETF.Add(profit)
			q.profits[year] = existingYearProfit
			existingYearSaleAggregate.ETF = existingYearSaleAggregate.ETF.Add(sellPrice)
			q.saleAggregates[year] = existingYearSaleAggregate
			if profit.LessThan(decimal.NewFromInt(0)) {
				existingYearLossAggregate.ETF = existingYearLossAggregate.ETF.Add(profit)
				q.lossAggregates[year] = existingYearLossAggregate
			} else {
				existingYearProfitAggregate.ETF = existingYearProfitAggregate.ETF.Add(profit)
				q.profitAggregates[year] = existingYearProfitAggregate
			}
		default:
			return merry.Errorf("invalid record type: %s", newRecordType)

		}
	}

	return nil
}

func TimeIsBetween(t, min, max time.Time) bool {
	if min.After(max) {
		min, max = max, min
	}
	return (t.Equal(min) || t.After(min)) && (t.Equal(max) || t.Before(max))
}

// FIFO default
// If sold withing 4 weeks of purchase, LIFO will apply when needed
// If bought within 4 weeks of sale, if a loss occurs on the initial disposal,
// then this loss can only be offset against a gain on the sale of shares of
// the same class which were purchased within 4 weeks of that sale.
func (q *PurchaseHistoryStruct) updateHistoryAndGetProfit(
	log logr.Logger, sellRecord Record) (decimal.Decimal, decimal.Decimal, error) {
	var buyPrice, sellPrice, profit, totalSale decimal.Decimal
	var err error

	for !sellRecord.NoOfShares.Equal(decimal.NewFromInt(0)) {
		if q.recordQueue.Size() <= 0 {
			return sellPrice, profit, merry.Errorf("not enough shares available to sell: %s", sellRecord.Ticker)
		}

		buyRecord := q.recordQueue.Peek(0)
		lastRecord := q.recordQueue.Peek(q.recordQueue.Size() - 1)
		if TimeIsBetween(lastRecord.Time, sellRecord.Time.AddDate(0, 0, -7*4), sellRecord.Time) {
			// Fits the bill for LIFO
			log.V(0).Info("LIFO processing...", "against", lastRecord)
			buyRecord = lastRecord
		}

		buyExchangeRateOverride := &buyRecord.ExchangeRate
		if sellRecord.CurrencyTotal != buyRecord.CurrencyTotal &&
			buyRecord.ExchangeRate.Equal(decimal.NewFromInt(1)) {
			// this means that conversion was taken place after purchase and selling were done
			// in the same currency as the buy. So we use the sell exchange rate to
			// culculate the buy price too to have like-for-like calculations

			buyExchangeRateOverride = &sellRecord.ExchangeRate
			log.V(0).Info("currency override",
				"old", buyRecord.ExchangeRate,
				"new", buyExchangeRateOverride)
		}

		if sellRecord.NoOfShares.LessThanOrEqual(buyRecord.NoOfShares) {
			// more shares available than to sell
			// TODO: make this currency aware for sales in a diff currency than buy (to use the same exchjange rate as the sale)
			// get price of shares to be sold
			buyPrice, err = buyRecord.GetActualPriceForQuantity(
				sellRecord.NoOfShares, buyExchangeRateOverride, true)
			if err != nil {
				return sellPrice, profit, merry.Errorf("failed to get buy price for sell action: %w", err)
			}

			// get price of sale action
			sellPrice, err = sellRecord.GetActualPriceForQuantity(
				sellRecord.NoOfShares, nil, false)
			if err != nil {
				return sellPrice, profit, merry.Errorf("failed to get sell price for sell action: %w", err)
			}

			sellRecord.NoOfShares = decimal.NewFromInt(0)

		} else {
			// save the value since the record is mutated
			currRecordNoOfShares := buyRecord.NoOfShares

			// sell off all stocks in this "buy record" to get the "buy price" at market value
			buyPrice, err = buyRecord.GetActualPriceForQuantity(
				buyRecord.NoOfShares, buyExchangeRateOverride, true)
			if err != nil {
				return sellPrice, profit, merry.Errorf("failed to get price for sell action: %w", err)
			}

			// get the profit from the sale for the number of shares you bought above
			sellPrice, err = sellRecord.GetActualPriceForQuantity(
				currRecordNoOfShares, nil, false)
			if err != nil {
				return sellPrice, profit, merry.Errorf("failed to get price for sell action: %w", err)
			}

		}
		totalSale = totalSale.Add(sellPrice)
		profit = profit.Add(sellPrice.Sub(buyPrice))

		if buyRecord.NoOfShares.LessThanOrEqual(decimal.NewFromInt(0)) {
			// get rid of record if it has no shares in it
			q.recordQueue.RemoveItem(buyRecord)
		}
	}
	return totalSale, profit, nil
}
