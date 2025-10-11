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
	GetProfitForYear(year int) Profits
	GetSaleAggregatesForYear(year int) Profits
}

type PurchaseHistoryStruct struct {
	recordQueue    RecordQueue
	profits        map[int]Profits
	saleAggregates map[int]Profits
}

func NewPurchaseHistory(recordQueue RecordQueue) PurchaseHistory {
	return &PurchaseHistoryStruct{
		recordQueue:    recordQueue,
		profits:        make(map[int]Profits),
		saleAggregates: make(map[int]Profits),
	}
}

func (q *PurchaseHistoryStruct) GetRecordQueue() RecordQueue {
	return q.recordQueue
}

func (q *PurchaseHistoryStruct) GetProfitForYear(year int) Profits {
	return q.profits[year]
}

func (q *PurchaseHistoryStruct) GetSaleAggregatesForYear(year int) Profits {
	return q.saleAggregates[year]
}

func (q *PurchaseHistoryStruct) Process(log logr.Logger, newRecord *Record) error {
	if !strings.Contains(newRecord.Action, "buy") && !strings.Contains(newRecord.Action, "sell") {
		return nil
	}

	log.Info(fmt.Sprintf("%-12s", newRecord.Action),
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
		profit, err := q.updateHistoryAndGetProfit(log, *newRecord)
		if err != nil {
			return merry.Errorf("failed to process new record: %w", err)
		}

		existingYearProfit := q.profits[year]
		existingYearSaleAggregate := q.saleAggregates[year]

		newRecordType := newRecord.GetType()
		switch newRecordType {
		case Stock:
			existingYearProfit.Stock = existingYearProfit.Stock.Add(profit)
			q.profits[year] = existingYearProfit
			existingYearSaleAggregate.Stock = existingYearSaleAggregate.Stock.Add(newRecord.Total)
			q.saleAggregates[year] = existingYearSaleAggregate
		case ETF:
			existingYearProfit.ETF = existingYearProfit.ETF.Add(profit)
			q.profits[year] = existingYearProfit
			existingYearSaleAggregate.ETF = existingYearSaleAggregate.ETF.Add(newRecord.Total)
			q.saleAggregates[year] = existingYearSaleAggregate
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
	log logr.Logger, sellRecord Record) (decimal.Decimal, error) {
	var profit decimal.Decimal

	for !sellRecord.NoOfShares.Equal(decimal.NewFromInt(0)) {
		if q.recordQueue.Size() <= 0 {
			return decimal.NewFromInt(0), merry.Errorf("not enough shares available to sell: %s", sellRecord.Ticker)
		}

		currRecord := q.recordQueue.Peek(0)
		lastRecord := q.recordQueue.Peek(q.recordQueue.Size() - 1)
		if TimeIsBetween(lastRecord.Time, sellRecord.Time.AddDate(0, 0, -7*4), sellRecord.Time) {
			// Fits the bill for LIFO
			log.Info("LIFO processing...", "against", lastRecord)
			currRecord = lastRecord
		}

		if sellRecord.NoOfShares.LessThanOrEqual(currRecord.NoOfShares) {
			// more shares available than to sell
			price, err := currRecord.GetActualPriceForQuantity(sellRecord.NoOfShares, true)
			if err != nil {
				return profit, merry.Errorf("failed to get price for sell action: %w", err)
			}
			profit = profit.Add(sellRecord.Total).Sub(price)
			sellRecord.NoOfShares = decimal.NewFromInt(0)

		} else {
			// save the value since the record is mutated
			currRecorNoOfShares := currRecord.NoOfShares

			// sell off all stocks in this "buy record" to get the "buy price" at market value
			buyCost, err := currRecord.GetActualPriceForQuantity(currRecord.NoOfShares, true)
			if err != nil {
				return profit, merry.Errorf("failed to get price for sell action: %w", err)
			}

			// get the profit from the sale for the number of shares you bought above
			sellProfit, err := sellRecord.GetActualPriceForQuantity(currRecorNoOfShares, false)
			if err != nil {
				return profit, merry.Errorf("failed to get price for sell action: %w", err)
			}

			profit = profit.Add(sellProfit.Sub(buyCost))

		}
		if currRecord.NoOfShares.LessThanOrEqual(decimal.NewFromInt(0)) {
			// get rid of record if it has no shares in it
			q.recordQueue.RemoveItem(currRecord)
		}
	}
	return profit, nil
}
