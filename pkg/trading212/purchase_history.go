package trading212

import (
	"strings"

	"github.com/ansel1/merry/v2"
	"github.com/go-logr/logr"
	"github.com/shopspring/decimal"
)

type PurchaseHistory interface {
	GetRecordQueue() RecordQueue
	Process(log logr.Logger, newRecord *Record) error
	GetProfitForYear(year int) Profits
}

type PurchaseHistoryStruct struct {
	recordQueue RecordQueue
	profits     map[int]Profits
}

func NewPurchaseHistory(recordQueue RecordQueue) PurchaseHistory {
	return &PurchaseHistoryStruct{
		recordQueue: recordQueue,
		profits:     make(map[int]Profits),
	}
}

func (q *PurchaseHistoryStruct) GetRecordQueue() RecordQueue {
	return q.recordQueue
}

func (q *PurchaseHistoryStruct) GetProfitForYear(year int) Profits {
	return q.profits[year]
}

func (q *PurchaseHistoryStruct) Process(log logr.Logger, newRecord *Record) error {
	if strings.Contains(newRecord.Action, "buy") {
		// log.Info("buy")
		q.recordQueue.Enqueue(newRecord)
	} else if strings.Contains(newRecord.Action, "sell") {
		// log.Info("sell")
		year, err := newRecord.GetYear()
		if err != nil {
			return merry.Errorf("failed to get year for record: %w", err)
		}
		profit, err := q.updateHistoryAndGetProfit(log, *newRecord)
		if err != nil {
			return merry.Errorf("failed to process new record: %w", err)
		}

		existingYearProfit := q.profits[year]

		newRecordType := newRecord.GetType()
		if newRecordType == Stock {
			existingYearProfit.Stock = existingYearProfit.Stock.Add(profit)
			q.profits[year] = existingYearProfit
		} else if newRecordType == ETF {
			existingYearProfit.ETF = existingYearProfit.ETF.Add(profit)
			q.profits[year] = existingYearProfit
		} else {
			return merry.Errorf("invalid record type: %s", newRecordType)

		}

	}

	return nil
}

func (q *PurchaseHistoryStruct) updateHistoryAndGetProfit(
	log logr.Logger, sellRecord Record) (decimal.Decimal, error) {
	var profit decimal.Decimal

	for !sellRecord.NoOfShares.Equal(decimal.NewFromInt(0)) {
		if q.recordQueue.Size() <= 0 {
			return decimal.NewFromInt(0), merry.Errorf("not enough shares available to sell: %s", sellRecord.Ticker)
		}
		currRecord := q.recordQueue.Peak()

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
			q.recordQueue.Dequeue()
		}
	}
	return profit, nil
}
