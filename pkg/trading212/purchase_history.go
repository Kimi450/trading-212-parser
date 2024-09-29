package trading212

import (
	"strings"

	"github.com/ansel1/merry/v2"
)

type PurchaseHistory interface {
	GetRecordQueue() RecordQueue
	Process(newRecord *Record) error
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

func (q *PurchaseHistoryStruct) Process(newRecord *Record) error {
	if strings.Contains(newRecord.Action, "buy") {
		println("buy")
		q.recordQueue.Enqueue(newRecord)
	} else if strings.Contains(newRecord.Action, "sell") {
		println("sell")
		year, err := newRecord.GetYear()
		if err != nil {
			return merry.Errorf("failed to get year for record: %w", err)
		}
		profit, err := q.updateHistoryAndGetProfit(*newRecord)
		if err != nil {
			return merry.Errorf("failed to process new record: %w", err)
		}

		q.profits[year] += profit
	}

	return nil
}

func (q *PurchaseHistoryStruct) updateHistoryAndGetProfit(
	sellRecord Record) (float64, error) {
	var profit float64

	for sellRecord.NoOfShares != 0 {
		currRecord := q.recordQueue.Peak()

		if sellRecord.NoOfShares <= currRecord.NoOfShares {
			// more share available than to sell
			price, err := currRecord.GetActualPriceForQuantity(sellRecord.NoOfShares, true)
			if err != nil {
				return profit, merry.Errorf("failed to get price for sell action: %w", err)
			}
			profit += (sellRecord.Total - price)
			println(sellRecord.Total)
			println(price)
			println(profit)

			sellRecord.NoOfShares = 0

		} else {
			sellRecord.NoOfShares -= currRecord.NoOfShares
		}
		if currRecord.NoOfShares <= 0 {
			// get rid of record if it has no shares in it
			q.recordQueue.Dequeue()
		}
	}

	return profit, nil
}
