package trading212

import (
	"strings"

	"github.com/ansel1/merry/v2"
	"github.com/go-logr/logr"
)

type PurchaseHistory interface {
	GetRecordQueue() RecordQueue
	Process(log logr.Logger, newRecord *Record) error
	GetProfitForYear(year int) float64
}

type PurchaseHistoryStruct struct {
	recordQueue RecordQueue
	profits     map[int]float64
}

func NewPurchaseHistory(recordQueue RecordQueue) PurchaseHistory {
	return &PurchaseHistoryStruct{
		recordQueue: recordQueue,
		profits:     make(map[int]float64),
	}
}

func (q *PurchaseHistoryStruct) GetRecordQueue() RecordQueue {
	return q.recordQueue
}

func (q *PurchaseHistoryStruct) GetProfitForYear(year int) float64 {
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

		q.profits[year] += profit
	}

	return nil
}

func (q *PurchaseHistoryStruct) updateHistoryAndGetProfit(
	log logr.Logger, sellRecord Record) (float64, error) {
	var profit float64

	for sellRecord.NoOfShares != 0 {
		if q.recordQueue.Size() <= 0 {
			return 0, merry.Errorf("not enough shares available to sell???")
		}
		currRecord := q.recordQueue.Peak()

		if sellRecord.NoOfShares <= currRecord.NoOfShares {
			// more shares available than to sell
			price, err := currRecord.GetActualPriceForQuantity(sellRecord.NoOfShares, true)
			if err != nil {
				return profit, merry.Errorf("failed to get price for sell action: %w", err)
			}
			profit += (sellRecord.Total - price)
			sellRecord.NoOfShares = 0

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

			profit += (sellProfit - buyCost)

		}
		if currRecord.NoOfShares <= 0 {
			// get rid of record if it has no shares in it
			q.recordQueue.Dequeue()
		}
	}

	return profit, nil
}
